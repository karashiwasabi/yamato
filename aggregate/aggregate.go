// File: aggregate/aggregate.go
package aggregate

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
)

// DB は main.go からセットされる DB ハンドル
var DB *sql.DB

// SetDB は外部から DB を注入します
func SetDB(db *sql.DB) {
	DB = db
}

// Detail は集計結果の 1 行を表します
type Detail struct {
	YJ            string `json:"yj"`
	Date          string `json:"date"`
	ProductName   string `json:"productName"`
	Type          string `json:"type"`
	Quantity      string `json:"quantity"`
	Unit          string `json:"unit"`
	Packaging     string `json:"packaging"`
	Count         string `json:"count"`
	UnitPrice     string `json:"unitPrice"`
	Subtotal      string `json:"subtotal"`
	ExpiryDate    string `json:"expiryDate"`
	LotNumber     string `json:"lotNumber"`
	OroshiCode    string `json:"oroshiCode"`
	ReceiptNumber string `json:"receiptNumber"`
	LineNumber    string `json:"lineNumber"`
}

// AggregateHandler は GET /aggregate を処理します。
// クエリパラメータ:
//
//	from              期間開始 (YYYY-MM-DD)
//	to                期間終了 (YYYY-MM-DD)
//	filter            商品名フィルタ(部分一致)
//	doyaku            毒薬 (1 を指定すると毒薬のみ)
//	gekiyaku          劇薬
//	mayaku            麻薬
//	kouseishinyaku    向精神薬 (1,2,3 の CSV)
//	kakuseizai        覚せい剤
//	kakuseizaiGenryou 覚せい剤原料
func AggregateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// 必須パラメータ取得
	rawFrom := q.Get("from")
	rawTo := q.Get("to")
	if rawFrom == "" || rawTo == "" {
		http.Error(w, "from, to を YYYY-MM-DD 形式で指定してください", http.StatusBadRequest)
		return
	}
	// YYYY-MM-DD → YYYYMMDD
	from := strings.ReplaceAll(rawFrom, "-", "")
	to := strings.ReplaceAll(rawTo, "-", "")

	// 任意フィルタ取得
	filter := strings.TrimSpace(q.Get("filter"))

	// --- DAT レコード取得 ---
	// SQL ビルド
	var argsDAT []interface{}
	argsDAT = append(argsDAT, from, to)
	sb := strings.Builder{}
	sb.WriteString(`
SELECT
  COALESCE(m.MA009JC009YJCode, '')         AS yj,
  d.DatDate                               AS date,
  COALESCE(m.MA018JC018ShouhinMei, '')    AS productName,
  CASE d.DatDeliveryFlag
    WHEN '1' THEN '納品'
    WHEN '2' THEN '返品'
    ELSE d.DatDeliveryFlag
  END                                      AS type,
  d.DatQuantity                           AS quantity,
  COALESCE(m.MA039JC039HousouTaniTani, '') AS unit,
  ''                                       AS packaging,
  d.DatQuantity                           AS count,
  d.DatUnitPrice                          AS unitPrice,
  d.DatSubtotal                           AS subtotal,
  d.DatExpiryDate                         AS expiryDate,
  d.DatLotNumber                          AS lotNumber,
  d.CurrentOroshiCode                     AS oroshiCode,
  d.DatReceiptNumber                      AS receiptNumber,
  d.DatLineNumber                         AS lineNumber
FROM datrecords d
LEFT JOIN ma0 m ON d.DatJanCode = m.MA000JC000JanCode
WHERE d.DatDate BETWEEN ? AND ?`)

	// 商品名フィルタ
	if filter != "" {
		sb.WriteString(" AND m.MA018JC018ShouhinMei LIKE ?")
		argsDAT = append(argsDAT, "%"+filter+"%")
	}

	// 毒薬・劇薬・麻薬フラグ
	if q.Get("doyaku") != "" {
		sb.WriteString(" AND m.MA061JC061Doyaku = '1'")
	}
	if q.Get("gekiyaku") != "" {
		sb.WriteString(" AND m.MA062JC062Gekiyaku = '1'")
	}
	if q.Get("mayaku") != "" {
		sb.WriteString(" AND m.MA063JC063Mayaku = '1'")
	}

	// 向精神薬フラグ (複数選択可)
	if ks := q.Get("kouseishinyaku"); ks != "" {
		codes := strings.Split(ks, ",")
		// プレースホルダー作成
		ph := make([]string, len(codes))
		for i, c := range codes {
			ph[i] = "?"
			argsDAT = append(argsDAT, c)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN (" + strings.Join(ph, ",") + ")")
	}

	// 覚せい剤フラグ
	if q.Get("kakuseizai") != "" {
		sb.WriteString(" AND m.MA065JC065Kakuseizai = '1'")
	}
	if q.Get("kakuseizaiGenryou") != "" {
		sb.WriteString(" AND m.MA066JC066KakuseizaiGenryou = '1'")
	}

	sqlDAT := sb.String()
	rowsDAT, err := DB.Query(sqlDAT, argsDAT...)
	if err != nil {
		log.Println("aggregate DAT query error:", err)
		http.Error(w, "DBエラー(DAT)", http.StatusInternalServerError)
		return
	}
	defer rowsDAT.Close()

	var details []Detail
	for rowsDAT.Next() {
		var d Detail
		if err := rowsDAT.Scan(
			&d.YJ, &d.Date, &d.ProductName, &d.Type,
			&d.Quantity, &d.Unit, &d.Packaging, &d.Count,
			&d.UnitPrice, &d.Subtotal, &d.ExpiryDate, &d.LotNumber,
			&d.OroshiCode, &d.ReceiptNumber, &d.LineNumber,
		); err != nil {
			log.Println("scan DAT row error:", err)
			continue
		}
		details = append(details, d)
	}

	// --- USAGE レコード取得 ---
	// 同じフラグ条件を適用するために ma0 を JOIN
	var argsUsage []interface{}
	argsUsage = append(argsUsage, from, to)
	sb2 := strings.Builder{}
	sb2.WriteString(`
SELECT
  u.usageYjCode         AS yj,
  u.usageDate           AS date,
  u.usageProductName    AS productName,
  '処方'                AS type,
  u.usageAmount         AS quantity,
  u.usageUnitName       AS unit,
  ''                     AS packaging,
  ''                     AS count,
  ''                     AS unitPrice,
  ''                     AS subtotal,
  ''                     AS expiryDate,
  ''                     AS lotNumber,
  ''                     AS oroshiCode,
  ''                     AS receiptNumber,
  ''                     AS lineNumber
FROM usagerecords u
LEFT JOIN ma0 m ON u.usageJanCode = m.MA000JC000JanCode
WHERE u.usageDate BETWEEN ? AND ?`)

	// フィルタ再利用
	if filter != "" {
		sb2.WriteString(" AND u.usageProductName LIKE ?")
		argsUsage = append(argsUsage, "%"+filter+"%")
	}
	if q.Get("doyaku") != "" {
		sb2.WriteString(" AND m.MA061JC061Doyaku = '1'")
	}
	if q.Get("gekiyaku") != "" {
		sb2.WriteString(" AND m.MA062JC062Gekiyaku = '1'")
	}
	if q.Get("mayaku") != "" {
		sb2.WriteString(" AND m.MA063JC063Mayaku = '1'")
	}
	if ks := q.Get("kouseishinyaku"); ks != "" {
		codes := strings.Split(ks, ",")
		ph := make([]string, len(codes))
		for i, c := range codes {
			ph[i] = "?"
			argsUsage = append(argsUsage, c)
		}
		sb2.WriteString(" AND m.MA064JC064Kouseishinyaku IN (" + strings.Join(ph, ",") + ")")
	}
	if q.Get("kakuseizai") != "" {
		sb2.WriteString(" AND m.MA065JC065Kakuseizai = '1'")
	}
	if q.Get("kakuseizaiGenryou") != "" {
		sb2.WriteString(" AND m.MA066JC066KakuseizaiGenryou = '1'")
	}

	sqlUsage := sb2.String()
	rowsUsage, err := DB.Query(sqlUsage, argsUsage...)
	if err != nil {
		log.Println("aggregate USAGE query error:", err)
		http.Error(w, "DBエラー(USAGE)", http.StatusInternalServerError)
		return
	}
	defer rowsUsage.Close()

	for rowsUsage.Next() {
		var d Detail
		if err := rowsUsage.Scan(
			&d.YJ, &d.Date, &d.ProductName, &d.Type,
			&d.Quantity, &d.Unit, &d.Packaging, &d.Count,
			&d.UnitPrice, &d.Subtotal, &d.ExpiryDate, &d.LotNumber,
			&d.OroshiCode, &d.ReceiptNumber, &d.LineNumber,
		); err != nil {
			log.Println("scan USAGE row error:", err)
			continue
		}
		details = append(details, d)
	}

	// --- グループ化・ソート・レスポンス ---
	groups := make(map[string][]Detail)
	for _, d := range details {
		groups[d.YJ] = append(groups[d.YJ], d)
	}
	for yj, list := range groups {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Date < list[j].Date
		})
		groups[yj] = list
	}

	log.Printf("aggregate: total rows = %d", len(details))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		log.Println("aggregate JSON encode error:", err)
	}
}
