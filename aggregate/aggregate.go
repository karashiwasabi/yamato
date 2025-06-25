// File: aggregate/aggregate.go
package aggregate

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"YAMATO/usage"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

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
	RawCount      string `json:"-"`
	HK            string `json:"-"`
	HS            string `json:"-"`
	HU            string `json:"-"`
	JSN           string `json:"-"`
	JSU           string `json:"-"`
	JSSN          string `json:"-"`
}

func AggregateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromRaw, toRaw := q.Get("from"), q.Get("to")
	filter := strings.TrimSpace(q.Get("filter"))
	log.Printf("▶ Aggregate from=%s to=%s filter=%q fullQuery=%s",
		fromRaw, toRaw, filter, r.URL.RawQuery)

	if fromRaw == "" || toRaw == "" {
		http.Error(w, "from/to は必須です", http.StatusBadRequest)
		return
	}
	from := strings.ReplaceAll(fromRaw, "-", "")
	to := strings.ReplaceAll(toRaw, "-", "")

	var details []Detail

	// --- DAT抽出 ---
	datArgs := []interface{}{from, to}
	sb := &strings.Builder{}
	sb.WriteString(`
SELECT
  COALESCE(m.MA009JC009YJCode,'') AS yj,
  d.DatDate AS date,
  COALESCE(m.MA018JC018ShouhinMei,'') AS productName,
  CASE d.DatDeliveryFlag WHEN '1' THEN '納品'
                        WHEN '2' THEN '返品'
                        ELSE d.DatDeliveryFlag END AS type,
  d.DatQuantity AS rawCount,
  COALESCE(m.MA039JC039HousouTaniTani,'') AS unit,
  '' AS packaging,
  d.DatUnitPrice AS unitPrice,
  d.DatSubtotal  AS subtotal,
  d.DatExpiryDate AS expiryDate,
  d.DatLotNumber  AS lotNumber,
  d.CurrentOroshiCode AS oroshiCode,
  d.DatReceiptNumber  AS receiptNumber,
  d.DatLineNumber     AS lineNumber,
  COALESCE(m.MA037JC037HousouKeitai,'')        AS hk,
  COALESCE(m.MA044JC044HousouSouryouSuuchi,'') AS hs,
  COALESCE(m.MA039JC039HousouTaniTani,'')      AS hu,
  COALESCE(m.MA131JA006HousouSuuryouSuuchi,'') AS jsn,
  COALESCE(m.MA132JA007HousouSuuryouTaniCode,'') AS jsu,
  COALESCE(m.MA133JA008HousouSouryouSuuchi,'') AS jssn
FROM datrecords d
LEFT JOIN ma0 m ON d.DatJanCode = m.MA000JC000JanCode
WHERE d.DatDate BETWEEN ? AND ?`)

	// テキストフィルタ
	if filter != "" {
		sb.WriteString(" AND m.MA018JC018ShouhinMei LIKE ?")
		datArgs = append(datArgs, "%"+filter+"%")
	}
	// チェックボックスフィルタ
	if q.Get("doyaku") == "1" {
		sb.WriteString(" AND m.MA061JC061Doyaku='1'")
	}
	if q.Get("gekiyaku") == "1" {
		sb.WriteString(" AND m.MA062JC062Gekiyaku='1'")
	}
	if q.Get("mayaku") == "1" {
		sb.WriteString(" AND m.MA063JC063Mayaku='1'")
	}
	if q.Get("kakuseizai") == "1" {
		sb.WriteString(" AND m.MA065JC065Kakuseizai='1'")
	}
	if q.Get("kakuseizaiGenryou") == "1" {
		sb.WriteString(" AND m.MA066JC066KakuseizaiGenryou='1'")
	}
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			datArgs = append(datArgs, v)
		}
		sb.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	datQuery := sb.String()
	log.Printf("▶ DAT SQL: %s\n   args=%v", datQuery, datArgs)
	rows, err := DB.Query(datQuery, datArgs...)
	if err != nil {
		http.Error(w, "DBエラー(DAT)", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var d Detail
		if err := rows.Scan(
			&d.YJ, &d.Date, &d.ProductName, &d.Type,
			&d.RawCount, &d.Unit, &d.Packaging,
			&d.UnitPrice, &d.Subtotal, &d.ExpiryDate,
			&d.LotNumber, &d.OroshiCode, &d.ReceiptNumber, &d.LineNumber,
			&d.HK, &d.HS, &d.HU, &d.JSN, &d.JSU, &d.JSSN,
		); err != nil {
			continue
		}
		hsVal, _ := strconv.Atoi(strings.TrimLeft(d.HS, "0"))
		rcVal, _ := strconv.Atoi(strings.TrimLeft(d.RawCount, "0"))
		d.Quantity = strconv.Itoa(hsVal * rcVal)
		d.Count = d.RawCount
		inner := d.JSN + d.HU + "×" + d.JSSN
		if d.JSU != "" && d.JSU != "0" {
			if nm := usage.GetTaniName(d.JSU); nm != "" {
				inner += nm
			}
		}
		d.Packaging = d.HK + d.HS + d.HU + "(" + inner + ")"
		details = append(details, d)
	}

	// --- USAGE抽出（DATと同じチェックボックス条件を適用） ---
	uArgs := []interface{}{from, to}
	ub := &strings.Builder{}
	ub.WriteString(`
SELECT
  u.usageDate,
  u.usageYjCode,
  u.usageJanCode,
  COALESCE(m.MA018JC018ShouhinMei,'') AS productName,
  u.usageAmount,
  u.usageUnitName
FROM usagerecords u
LEFT JOIN ma0 m ON u.usageJanCode = m.MA000JC000JanCode
WHERE u.usageDate BETWEEN ? AND ?`)

	// 同じチェックボックス
	if filter != "" {
		ub.WriteString(" AND m.MA018JC018ShouhinMei LIKE ?")
		uArgs = append(uArgs, "%"+filter+"%")
	}
	if q.Get("doyaku") == "1" {
		ub.WriteString(" AND m.MA061JC061Doyaku='1'")
	}
	if q.Get("gekiyaku") == "1" {
		ub.WriteString(" AND m.MA062JC062Gekiyaku='1'")
	}
	if q.Get("mayaku") == "1" {
		ub.WriteString(" AND m.MA063JC063Mayaku='1'")
	}
	if q.Get("kakuseizai") == "1" {
		ub.WriteString(" AND m.MA065JC065Kakuseizai='1'")
	}
	if q.Get("kakuseizaiGenryou") == "1" {
		ub.WriteString(" AND m.MA066JC066KakuseizaiGenryou='1'")
	}
	if ks := q.Get("kouseishinyaku"); ks != "" {
		parts := strings.Split(ks, ",")
		ph := make([]string, len(parts))
		for i, v := range parts {
			ph[i] = "?"
			uArgs = append(uArgs, v)
		}
		ub.WriteString(" AND m.MA064JC064Kouseishinyaku IN(" + strings.Join(ph, ",") + ")")
	}

	usageQuery := ub.String()
	log.Printf("▶ USAGE SQL: %s\n   args=%v", usageQuery, uArgs)
	rowsU, err := DB.Query(usageQuery, uArgs...)
	if err == nil {
		defer rowsU.Close()
		for rowsU.Next() {
			var d Detail
			var date, yj, jan, pname, amt, unitName string
			if err := rowsU.Scan(&date, &yj, &jan, &pname, &amt, &unitName); err != nil {
				continue
			}
			d.YJ = yj
			d.Date = date
			d.ProductName = pname
			d.Type = "処方"
			d.Quantity = amt
			d.Unit = unitName
			d.Count = ""
			details = append(details, d)
		}
	}

	// グループ化・ソート
	groups := make(map[string][]Detail)
	for _, d := range details {
		groups[d.YJ] = append(groups[d.YJ], d)
	}
	for k, list := range groups {
		sort.Slice(list, func(i, j int) bool { return list[i].Date < list[j].Date })
		groups[k] = list
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(groups)
}
