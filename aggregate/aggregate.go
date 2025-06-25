// File: aggregate/aggregate.go
package aggregate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"YAMATO/usage"
)

// DB は main.go からセットされる DB ハンドル
var DB *sql.DB

// SetDB によって外部から DB を注入
func SetDB(db *sql.DB) {
	DB = db
}

// Detail は集計結果１行分を表す
type Detail struct {
	YJ            string `json:"yj"`
	Date          string `json:"date"`
	ProductName   string `json:"productName"`
	Type          string `json:"type"`
	Quantity      string `json:"quantity"` // HS×RawCount の計算結果
	Unit          string `json:"unit"`
	Packaging     string `json:"packaging"` // 組み立てた包装文字列
	Count         string `json:"count"`     // 元の個数（RawCount）
	UnitPrice     string `json:"unitPrice"`
	Subtotal      string `json:"subtotal"`
	ExpiryDate    string `json:"expiryDate"`
	LotNumber     string `json:"lotNumber"`
	OroshiCode    string `json:"oroshiCode"`
	ReceiptNumber string `json:"receiptNumber"`
	LineNumber    string `json:"lineNumber"`

	// 補助フィールド（JSONに含めない）
	RawCount string `json:"-"`
	HK       string `json:"-"`
	HS       string `json:"-"`
	HU       string `json:"-"`
	JSN      string `json:"-"`
	JSU      string `json:"-"`
	JSSN     string `json:"-"`
}

// AggregateHandler は GET /aggregate を処理
func AggregateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromRaw, toRaw := q.Get("from"), q.Get("to")
	if fromRaw == "" || toRaw == "" {
		http.Error(w, "from/to は必須です", http.StatusBadRequest)
		return
	}
	from := strings.ReplaceAll(fromRaw, "-", "")
	to := strings.ReplaceAll(toRaw, "-", "")

	// --- DAT レコード抽出 ---
	args := []interface{}{from, to}
	sb := &strings.Builder{}
	sb.WriteString(`
SELECT
  COALESCE(m.MA009JC009YJCode,'')               AS yj,
  d.DatDate                                     AS date,
  COALESCE(m.MA018JC018ShouhinMei,'')            AS productName,
  CASE d.DatDeliveryFlag
    WHEN '1' THEN '納品'
    WHEN '2' THEN '返品'
    ELSE d.DatDeliveryFlag
  END                                            AS type,
  d.DatQuantity                                 AS rawCount,
  COALESCE(m.MA039JC039HousouTaniTani,'')        AS unit,
  ''                                             AS packaging,
  d.DatUnitPrice                                AS unitPrice,
  d.DatSubtotal                                 AS subtotal,
  d.DatExpiryDate                               AS expiryDate,
  d.DatLotNumber                                AS lotNumber,
  d.CurrentOroshiCode                           AS oroshiCode,
  d.DatReceiptNumber                            AS receiptNumber,
  d.DatLineNumber                               AS lineNumber,
  COALESCE(m.MA037JC037HousouKeitai,'')          AS hk,
  COALESCE(m.MA044JC044HousouSouryouSuuchi,'')   AS hs,
  COALESCE(m.MA039JC039HousouTaniTani,'')        AS hu,
  COALESCE(m.MA131JA006HousouSuuryouSuuchi,'')   AS jsn,
  COALESCE(m.MA132JA007HousouSuuryouTaniCode,'') AS jsu,
  COALESCE(m.MA133JA008HousouSouryouSuuchi,'')   AS jssn
FROM datrecords d
LEFT JOIN ma0 m ON d.DatJanCode = m.MA000JC000JanCode
WHERE d.DatDate BETWEEN ? AND ?`)
	if filter := q.Get("filter"); filter != "" {
		sb.WriteString(" AND m.MA018JC018ShouhinMei LIKE ?")
		args = append(args, "%"+filter+"%")
	}
	queryDAT := sb.String()

	rows, err := DB.Query(queryDAT, args...)
	if err != nil {
		log.Println("DAT query error:", err)
		http.Error(w, "DBエラー(DAT)", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var details []Detail
	for rows.Next() {
		var d Detail
		if err := rows.Scan(
			&d.YJ, &d.Date, &d.ProductName, &d.Type,
			&d.RawCount, &d.Unit, &d.Packaging,
			&d.UnitPrice, &d.Subtotal, &d.ExpiryDate,
			&d.LotNumber, &d.OroshiCode, &d.ReceiptNumber, &d.LineNumber,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
		); err != nil {
			log.Println("scan DAT row error:", err)
			continue
		}

		// 数量 = HS × rawCount
		hsVal, _ := strconv.Atoi(strings.TrimLeft(d.HS, "0"))
		rcVal, _ := strconv.Atoi(strings.TrimLeft(d.RawCount, "0"))
		d.Quantity = fmt.Sprintf("%d", hsVal*rcVal)
		d.Count = d.RawCount

		// 包装文字列組み立て + ログ出力確認
		outer := d.HK + d.HS + d.HU
		inner := d.JSN + d.HU + "×" + d.JSSN

		if d.JSU != "" && d.JSU != "0" {
			// TANIマップから単位名称取得
			extra := usage.GetTaniName(d.JSU)
			log.Printf("TANI lookup: code=%q → name=%q", d.JSU, extra)
			if extra != "" {
				inner += extra
			}
		} else {
			log.Printf("TANI lookup skipped for code=%q", d.JSU)
		}

		d.Packaging = outer + "(" + inner + ")"
		log.Printf("Packaged: %q", d.Packaging)

		details = append(details, d)
	}

	// --- USAGE レコード抽出（既存ロジック）---
	// ParseUsageFile → usage.UsageRecord を Detail にマッピングし、details に append

	// --- グループ化・ソート ---
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

	// --- JSON 出力 ---
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		log.Println("JSON encode error:", err)
	}
}
