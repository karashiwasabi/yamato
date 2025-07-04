package aggregate

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"YAMATO/usage"
)

var DB *sql.DB

// SetDB は main から DB を受け取ります
func SetDB(db *sql.DB) {
	DB = db
}

// Detail は /aggregate が返す明細行
type Detail struct {
	YJ            string `json:"yj"`
	ProductName   string `json:"productName"`
	Date          string `json:"date"`
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

	// 内部用
	RawCount     string `json:"-"`
	HK           string `json:"-"`
	HS           string `json:"-"`
	HU           string `json:"-"`
	JSN          string `json:"-"`
	JSU          string `json:"-"`
	JSSN         string `json:"-"`
	PackagingKey string `json:"-"`
}

// YJResult は YJ コード単位のまとめ
type YJResult struct {
	ProductName string              `json:"productName"`
	Groups      map[string][]Detail `json:"groups"`
}

// parseParams は from/to とフィルタを取得
func parseParams(r *http.Request) (from, to string, q url.Values, errMsg string, code int) {
	q = r.URL.Query()
	fr, tr := q.Get("from"), q.Get("to")
	if fr == "" || tr == "" {
		return "", "", nil, "from/to は必須です", http.StatusBadRequest
	}
	from = strings.ReplaceAll(fr, "-", "")
	to = strings.ReplaceAll(tr, "-", "")
	return from, to, q, "", 0
}

// fetchDatDetails は DAT レコードを取り Detail に変換
func fetchDatDetails(from, to string, q url.Values) ([]Detail, error) {
	var details []Detail
	args := []interface{}{from, to}
	sb := &strings.Builder{}
	sb.WriteString(`
SELECT
  -- YJ: MA0 → MA2 フォールバック
  COALESCE(NULLIF(m.MA009JC009YJCode,''), m2.MA2YjCode, '')                  AS yj,
  -- 商品名: MA0 → MA2
  COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei, '')            AS productName,
  d.DatDate                                                               AS date,
  CASE d.DatDeliveryFlag WHEN '1' THEN '納品'
                         WHEN '2' THEN '返品'
                         ELSE d.DatDeliveryFlag END                         AS type,
  d.DatQuantity                                                           AS rawCount,
  -- 単位（包装単位コード→名称は Go 側で補完）
  COALESCE(NULLIF(m.MA039JC039HousouTaniTani,''), m2.HousouTaniUnit, '')    AS unit,
  ''                                                                       AS packaging,
  d.DatUnitPrice                                                           AS unitPrice,
  d.DatSubtotal                                                            AS subtotal,
  d.DatExpiryDate                                                          AS expiryDate,
  d.DatLotNumber                                                           AS lotNumber,
  d.CurrentOroshiCode                                                      AS oroshiCode,
  d.DatReceiptNumber                                                       AS receiptNumber,
  d.DatLineNumber                                                          AS lineNumber,
  -- 包装情報: MA0 → MA2
  COALESCE(NULLIF(m.MA037JC037HousouKeitai,''), m2.HousouKeitai, '')       AS hk,
  COALESCE(NULLIF(m.MA044JC044HousouSouryouSuuchi,''), CAST(m2.HousouSouryouNumber AS TEXT), '') AS hs,
  COALESCE(NULLIF(m.MA039JC039HousouTaniTani,''), m2.HousouTaniUnit, '')    AS hu,
  COALESCE(NULLIF(m.MA131JA006HousouSuuryouSuuchi,''), CAST(m2.JanHousouSuuryouNumber AS TEXT), '') AS jsn,
  COALESCE(NULLIF(m.MA132JA007HousouSuuryouTaniCode,''), m2.JanHousouSuuryouUnit, '')   AS jsu,
  COALESCE(NULLIF(m.MA133JA008HousouSouryouSuuchi,''), CAST(m2.JanHousouSouryouNumber AS TEXT), '')    AS jssn
FROM datrecords d
LEFT JOIN ma0 m  ON d.DatJanCode = m.MA000JC000JanCode
LEFT JOIN ma2 m2 ON d.DatJanCode = m2.MA2JanCode
WHERE d.DatDate BETWEEN ? AND ?
`)
	if f := q.Get("filter"); f != "" {
		sb.WriteString(" AND COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei) LIKE ?")

		args = append(args, "%"+f+"%")
	}
	for _, c := range []struct{ name, col string }{
		{"doyaku", "MA061JC061Doyaku"},
		{"gekiyaku", "MA062JC062Gekiyaku"},
		{"mayaku", "MA063JC063Mayaku"},
		{"kakuseizai", "MA065JC065Kakuseizai"},
		{"kakuseizaiGenryou", "MA066JC066KakuseizaiGenryou"},
	} {
		if q.Get(c.name) == "1" {
			sb.WriteString(" AND m." + c.col + "='1'")
		}
	}
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			args = append(args, v)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	query := sb.String()
	log.Printf("▶ DAT SQL: %s\n   args=%v", query, args)

	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Printf("▶ DAT Query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Detail
		if err := rows.Scan(
			&d.YJ, &d.ProductName, &d.Date, &d.Type,
			&d.RawCount, &d.Unit, &d.Packaging,
			&d.UnitPrice, &d.Subtotal, &d.ExpiryDate,
			&d.LotNumber, &d.OroshiCode, &d.ReceiptNumber, &d.LineNumber,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
		); err != nil {
			log.Printf("▶ DAT Scan error: %v", err)
			continue
		}
		// 単位・包装単位コード→名称
		if nm := usage.GetTaniName(d.Unit); nm != "" {
			d.Unit = nm
		}
		if nm := usage.GetTaniName(d.HU); nm != "" {
			d.HU = nm
		}

		// 数量計算
		hsVal, _ := strconv.Atoi(strings.TrimLeft(d.HS, "0"))
		rcVal, _ := strconv.Atoi(strings.TrimLeft(d.RawCount, "0"))
		d.Quantity = strconv.Itoa(hsVal * rcVal)
		d.Count = d.RawCount

		// Packaging文字列
		inner := d.JSN + d.HU + "×" + d.JSSN
		if d.JSU != "" && d.JSU != "0" {
			if nm := usage.GetTaniName(d.JSU); nm != "" {
				inner += nm
			}
		}
		d.Packaging = d.HK + d.HS + d.HU + "(" + inner + ")"

		details = append(details, d)
	}
	return details, nil
}

// fetchUsageDetails は USAGE レコードを取り Detail に変換します
func fetchUsageDetails(from, to string, q url.Values) ([]Detail, error) {
	var details []Detail
	args := []interface{}{from, to}
	sb := &strings.Builder{}

	sb.WriteString(`
SELECT
  u.usageDate                                               AS date,
  -- YJ フォールバック
  COALESCE(NULLIF(m.MA009JC009YJCode,''), m2.MA2YjCode, '') AS yj,
  -- 商品名 フォールバック
  COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei, '') AS productName,
  u.usageAmount                                            AS rawCount,
  u.usageUnitName                                          AS unit,
  -- 包装情報: MA0 → MA2
  COALESCE(NULLIF(m.MA037JC037HousouKeitai,''), m2.HousouKeitai, '')        AS hk,
  COALESCE(NULLIF(m.MA044JC044HousouSouryouSuuchi,''), CAST(m2.HousouSouryouNumber AS TEXT), '') AS hs,
  COALESCE(NULLIF(m.MA039JC039HousouTaniTani,''), m2.HousouTaniUnit, '')    AS hu,
  COALESCE(NULLIF(m.MA131JA006HousouSuuryouSuuchi,''), CAST(m2.JanHousouSuuryouNumber AS TEXT), '') AS jsn,
  COALESCE(NULLIF(m.MA132JA007HousouSuuryouTaniCode,''), m2.JanHousouSuuryouUnit, '')   AS jsu,
  COALESCE(NULLIF(m.MA133JA008HousouSouryouSuuchi,''), CAST(m2.JanHousouSouryouNumber AS TEXT), '')    AS jssn
FROM usagerecords u
LEFT JOIN ma0  m  ON u.usageJanCode = m.MA000JC000JanCode
LEFT JOIN ma2  m2 ON u.usageJanCode = m2.MA2JanCode
WHERE u.usageDate BETWEEN ? AND ?
`)

	// 商品名フィルタ
	if f := q.Get("filter"); f != "" {
		sb.WriteString(" AND COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei) LIKE ?")
		args = append(args, "%"+f+"%")
	}

	// 毒薬／劇薬／麻薬フラグ
	for _, c := range []struct{ name, col string }{
		{"doyaku", "MA061JC061Doyaku"},
		{"gekiyaku", "MA062JC062Gekiyaku"},
		{"mayaku", "MA063JC063Mayaku"},
		{"kakuseizai", "MA065JC065Kakuseizai"},
		{"kakuseizaiGenryou", "MA066JC066KakuseizaiGenryou"},
	} {
		if q.Get(c.name) == "1" {
			sb.WriteString(" AND m." + c.col + "='1'")
		}
	}

	// 向1/2/3 フィルタ
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			args = append(args, v)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	query := sb.String()
	log.Printf("▶ USAGE SQL: %s\n   args=%v", query, args)

	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Printf("▶ USAGE Query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Detail
		var rawCount, unitName string

		if err := rows.Scan(
			&d.Date,
			&d.YJ,
			&d.ProductName,
			&rawCount,
			&unitName,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
		); err != nil {
			log.Printf("▶ USAGE Scan error: %v", err)
			continue
		}

		d.Type = "処方"
		d.Quantity = rawCount
		d.Unit = unitName

		// 単位名称補完
		if nm := usage.GetTaniName(d.HU); nm != "" {
			d.HU = nm
		}

		d.Count = ""

		// PackagingKey + Packaging 組み立て
		d.PackagingKey = d.HK + d.JSN + d.HU
		inner := d.JSN + d.HU + "×" + d.JSSN
		if d.JSU != "" && d.JSU != "0" {
			if nm := usage.GetTaniName(d.JSU); nm != "" {
				inner += nm
			}
		}
		d.Packaging = d.HK + d.HS + d.HU + "(" + inner + ")"

		details = append(details, d)
	}

	return details, nil
}

// groupDetails は Detail を YJ→PackagingKey でまとめる
func groupDetails(details []Detail) map[string]YJResult {
	tmp := make(map[string]map[string][]Detail)
	for i := range details {
		d := &details[i]
		d.PackagingKey = d.HK + d.JSN + d.HU
		if tmp[d.YJ] == nil {
			tmp[d.YJ] = make(map[string][]Detail)
		}
		tmp[d.YJ][d.PackagingKey] = append(tmp[d.YJ][d.PackagingKey], *d)
	}
	for _, pkMap := range tmp {
		for pk, list := range pkMap {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Date < list[j].Date
			})
			pkMap[pk] = list
		}
	}
	resp := make(map[string]YJResult, len(tmp))
	for yj, pkMap := range tmp {
		name := ""
		for _, list := range pkMap {
			if len(list) > 0 {
				name = list[0].ProductName
				break
			}
		}
		resp[yj] = YJResult{ProductName: name, Groups: pkMap}
	}
	return resp
}

// renderResponse は JSON レスポンスを返す
func renderResponse(w http.ResponseWriter, data map[string]YJResult) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(data) == 0 {
		_, err := w.Write([]byte(`{}`)) // ← 空でも JSON 保証
		return err
	}
	return json.NewEncoder(w).Encode(data)
}

// AggregateHandler は /aggregate エンドポイント
func AggregateHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[AGGREGATE panic] %v", rec)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	from, to, q, errMsg, code := parseParams(r)
	if errMsg != "" {
		log.Printf("[AGGREGATE] invalid params: %s", errMsg)
		http.Error(w, errMsg, code)
		return
	}

	// ① 各種フェッチ処理
	dats, err := fetchDatDetails(from, to, q)
	if err != nil {
		log.Printf("[AGGREGATE] fetchDatDetails error: %v", err)
		http.Error(w, "DAT Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	usgs, err := fetchUsageDetails(from, to, q)
	if err != nil {
		log.Printf("[AGGREGATE] fetchUsageDetails error: %v", err)
		http.Error(w, "USAGE Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	invs, err := fetchInvDetails(from, to, q)
	if err != nil {
		log.Printf("[AGGREGATE] fetchInvDetails error: %v", err)
		http.Error(w, "INV Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	iods, err := fetchIodDetails(from, to, q)
	if err != nil {
		log.Printf("[AGGREGATE] fetchIodDetails error: %v", err)
		http.Error(w, "IOD Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ② 全件を append して一括 group
	all := append([]Detail{}, dats...)
	all = append(all, usgs...)
	all = append(all, invs...)
	all = append(all, iods...)

	log.Printf("[AGGREGATE] grouped %d items", len(all))
	resp := groupDetails(all)

	// ③ JSONとして返却
	if err := renderResponse(w, resp); err != nil {
		log.Printf("[AGGREGATE] renderResponse error: %v", err)
	}
}

// fetchInvDetails は inventory テーブルの棚卸データを取り Detail に変換します。
// filter／毒劇麻フラグを q から受け取り WHERE に適用します。
func fetchInvDetails(from, to string, q url.Values) ([]Detail, error) {
	var details []Detail

	// SQL ビルダ
	args := []interface{}{from, to}
	sb := &strings.Builder{}
	sb.WriteString(`
SELECT
  inv.invDate                                           AS date,
  -- YJ／商品名：MA0→MA2 フォールバック
  COALESCE(NULLIF(m.MA009JC009YJCode,''), m2.MA2YjCode, '')      AS yj,
  COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei, '') AS productName,
  '棚卸'                                                 AS type,
  -- 生数量／包装単位
  CAST(inv.invJanHousouSuuryouNumber AS TEXT)            AS rawCount,
  inv.InvHousouTaniUnit                                  AS unit,
  CAST(inv.qty AS TEXT)                                  AS quantity,
  -- 包装情報: MA0→MA2 フォールバック
  COALESCE(NULLIF(m.MA037JC037HousouKeitai,''), m2.HousouKeitai, '')                          AS hk,
  COALESCE(NULLIF(m.MA044JC044HousouSouryouSuuchi,''), CAST(m2.HousouSouryouNumber AS TEXT), '') AS hs,
  COALESCE(NULLIF(m.MA039JC039HousouTaniTani,''), m2.HousouTaniUnit, '')                         AS hu,
  COALESCE(NULLIF(m.MA131JA006HousouSuuryouSuuchi,''), CAST(m2.JanHousouSuuryouNumber AS TEXT), '') AS jsn,
  COALESCE(NULLIF(m.MA132JA007HousouSuuryouTaniCode,''), m2.JanHousouSuuryouUnit, '')             AS jsu,
  COALESCE(NULLIF(m.MA133JA008HousouSouryouSuuchi,''), CAST(m2.JanHousouSouryouNumber AS TEXT), '') AS jssn
FROM inventory inv
LEFT JOIN ma0  m  ON inv.invJanCode = m.MA000JC000JanCode
LEFT JOIN ma2  m2 ON inv.invJanCode = m2.MA2JanCode
WHERE inv.invDate BETWEEN ? AND ?`)

	// 商品名フィルタ
	if f := q.Get("filter"); f != "" {
		sb.WriteString(" AND inv.invProductName LIKE ?")
		args = append(args, "%"+f+"%")
	}

	// 毒薬／劇薬／麻薬フラグ
	for _, c := range []struct{ name, col string }{
		{"doyaku", "MA061JC061Doyaku"},
		{"gekiyaku", "MA062JC062Gekiyaku"},
		{"mayaku", "MA063JC063Mayaku"},
		{"kakuseizai", "MA065JC065Kakuseizai"},
		{"kakuseizaiGenryou", "MA066JC066KakuseizaiGenryou"},
	} {
		if q.Get(c.name) == "1" {
			sb.WriteString(" AND m." + c.col + "='1'")
		}
	}

	// 向精神薬フラグ
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			args = append(args, v)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	query := sb.String()
	log.Printf("▶ INV SQL: %s\n   args=%v", query, args)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Detail
		if err := rows.Scan(
			&d.Date,
			&d.YJ,
			&d.ProductName,
			&d.Type,
			&d.RawCount,
			&d.Unit, // ここにコードが入っている
			&d.Quantity,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
		); err != nil {
			log.Printf("▶ INV Scan error: %v", err)
			continue
		}

		// ↓↓↓ 追加 ↓↓↓
		// コード→名称変換: Unit, HU, JSU
		if nm := usage.GetTaniName(d.Unit); nm != "" {
			d.Unit = nm
		}
		if nm := usage.GetTaniName(d.HU); nm != "" {
			d.HU = nm
		}
		if nm := usage.GetTaniName(d.JSU); nm != "" {
			d.JSU = nm
		}
		// ↑↑↑ 追加 ↑↑↑

		// Count はパック数表示用
		d.Count = ""

		// PackagingKey & Packaging 組み立て
		d.PackagingKey = d.HK + d.JSN + d.HU
		inner := d.JSN + d.HU + "×" + d.JSSN
		if d.JSU != "" && d.JSU != "0" {
			inner += d.JSU
		}
		d.Packaging = d.HK + d.HS + d.HU + "(" + inner + ")"

		details = append(details, d)
	}
	return details, nil
}

// fetchIodDetails は IOD テーブルを読み込み Detail に変換します。
// 商品名フィルタや毒薬・劇薬・麻薬・向精神薬条件を q から受け取って WHERE に適用します。
func fetchIodDetails(from, to string, q url.Values) ([]Detail, error) {
	var details []Detail

	// SQL ビルダ
	args := []interface{}{from, to}
	sb := &strings.Builder{}
	sb.WriteString(`
SELECT
  iod.iodDate                                               AS date,
  -- YJ／商品名: MA0→MA2 フォールバック
  COALESCE(NULLIF(m.MA009JC009YJCode,''), m2.MA2YjCode, '')          AS yj,
  COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei, '')      AS productName,
  CASE iodType WHEN '3' THEN '出庫' WHEN '4' THEN '入庫' ELSE iodType END AS type,
  CAST(iod.iodJanQuantity   AS TEXT)                             AS rawCount,
  iod.iodUnit                                                 AS unit,
  CAST(iod.iodQuantity      AS TEXT)                             AS quantity,
  -- 包装情報: MA0→MA2 フォールバック
  COALESCE(NULLIF(m.MA037JC037HousouKeitai,''), m2.HousouKeitai, '')                      AS hk,
  COALESCE(NULLIF(m.MA044JC044HousouSouryouSuuchi,''), CAST(m2.HousouSouryouNumber AS TEXT), '') AS hs,
  COALESCE(NULLIF(m.MA039JC039HousouTaniTani,''), m2.HousouTaniUnit, '')                    AS hu,
  COALESCE(NULLIF(m.MA131JA006HousouSuuryouSuuchi,''), CAST(m2.JanHousouSuuryouNumber AS TEXT), '') AS jsn,
  COALESCE(NULLIF(m.MA132JA007HousouSuuryouTaniCode,''), m2.JanHousouSuuryouUnit, '')        AS jsu,
  COALESCE(NULLIF(m.MA133JA008HousouSouryouSuuchi,''), CAST(m2.JanHousouSouryouNumber AS TEXT), '') AS jssn,
  iod.iodOroshiCode                                              AS oroshiCode,
  iod.iodReceiptNumber                                           AS receiptNumber,
  CAST(iod.iodLineNumber  AS TEXT)                               AS lineNumber
FROM iod
LEFT JOIN ma0  m  ON iod.iodJan = m.MA000JC000JanCode
LEFT JOIN ma2  m2 ON iod.iodJan = m2.MA2JanCode
WHERE iod.iodDate BETWEEN ? AND ?`)

	// 商品名フィルタ
	if f := q.Get("filter"); f != "" {
		sb.WriteString(" AND COALESCE(NULLIF(m.MA018JC018ShouhinMei,''), m2.Shouhinmei) LIKE ?")
		args = append(args, "%"+f+"%")
	}

	// 毒薬／劇薬／麻薬フラグ
	for _, c := range []struct{ name, col string }{
		{"doyaku", "MA061JC061Doyaku"},
		{"gekiyaku", "MA062JC062Gekiyaku"},
		{"mayaku", "MA063JC063Mayaku"},
		{"kakuseizai", "MA065JC065Kakuseizai"},
		{"kakuseizaiGenryou", "MA066JC066KakuseizaiGenryou"},
	} {
		if q.Get(c.name) == "1" {
			sb.WriteString(" AND m." + c.col + "='1'")
		}
	}

	// 向精神薬フラグ
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			args = append(args, v)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	query := sb.String()
	log.Printf("▶ IOD SQL: %s\n   args=%v", query, args)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Detail
		if err := rows.Scan(
			&d.Date,
			&d.YJ,
			&d.ProductName,
			&d.Type,
			&d.RawCount,
			&d.Unit,
			&d.Quantity,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
			&d.OroshiCode,
			&d.ReceiptNumber,
			&d.LineNumber,
		); err != nil {
			log.Printf("▶ IOD Scan error: %v", err)
			continue
		}

		// ← ここでコード→名称に変換
		if nm := usage.GetTaniName(d.Unit); nm != "" {
			d.Unit = nm
		}

		// Count はパック数
		d.Count = d.RawCount

		// PackagingKey + Packaging 組み立て
		d.PackagingKey = d.HK + d.JSN + d.HU
		inner := d.JSN + d.HU + "×" + d.JSSN
		if d.JSU != "" && d.JSU != "0" {
			if nm := usage.GetTaniName(d.JSU); nm != "" {
				inner += nm
			}
		}
		d.Packaging = d.HK + d.HS + d.HU + "(" + inner + ")"

		details = append(details, d)
	}

	return details, nil
}
