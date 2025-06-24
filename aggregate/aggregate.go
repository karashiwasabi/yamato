// File: aggregate/aggregate.go
// Package aggregate implements the /aggregate endpoint.
// It fetches DAT and USAGE records in a date range,
// joins DAT with MA0 to get YJ and unit, groups by YJ,
// sorts each group by date, and returns JSON—without productName,
// since you’ll load that later from MA0.

package aggregate

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
)

// DB is injected by main.go
var DB *sql.DB

// SetDB provides the DB handle to this package.
func SetDB(db *sql.DB) {
	DB = db
}

// Detail is one row of the aggregated result, minus productName.
type Detail struct {
	YJ            string `json:"yj"`
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
}

// AggregateHandler handles GET /aggregate?from=YYYY-MM-DD&to=YYYY-MM-DD
func AggregateHandler(w http.ResponseWriter, r *http.Request) {
	// 1) parse & normalize query dates
	rawFrom := r.URL.Query().Get("from")
	rawTo := r.URL.Query().Get("to")
	if rawFrom == "" || rawTo == "" {
		http.Error(w, "from,to を YYYY-MM-DD 形式で指定してください", http.StatusBadRequest)
		return
	}
	from := strings.ReplaceAll(rawFrom, "-", "") // e.g. "20250601"
	to := strings.ReplaceAll(rawTo, "-", "")     // e.g. "20250630"
	log.Printf("aggregate: from=%s to=%s", from, to)

	var details []Detail

	// 2) DAT records + MA0 join for YJ & unit
	rows1, err := DB.Query(`
    SELECT
      COALESCE(m.MA009JC009YJCode, '')         AS yj,
      d.DatDate                                AS date,
      CASE d.DatDeliveryFlag
        WHEN '1' THEN '納品'
        WHEN '2' THEN '返品'
        ELSE d.DatDeliveryFlag
      END                                      AS type,
      d.DatQuantity                            AS quantity,
      COALESCE(m.MA039JC039HousouTaniTani, '') AS unit,
      ''                                       AS packaging,
      d.DatQuantity                            AS count,
      d.DatUnitPrice                           AS unit_price,
      d.DatSubtotal                            AS subtotal,
      d.DatExpiryDate                          AS expiry_date,
      d.DatLotNumber                           AS lot_number,
      d.CurrentOroshiCode                      AS oroshi_code,
      d.DatReceiptNumber                       AS receipt_number,
      d.DatLineNumber                          AS line_number
    FROM datrecords d
    LEFT JOIN ma0 m
      ON d.DatJanCode = m.MA000JC000JanCode
    WHERE d.DatDate BETWEEN ? AND ?
  `, from, to)
	if err != nil {
		log.Println("aggregate DAT query error:", err)
		http.Error(w, "DBエラー(DAT)", http.StatusInternalServerError)
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		var d Detail
		if err := rows1.Scan(
			&d.YJ,
			&d.Date,
			&d.Type,
			&d.Quantity,
			&d.Unit,
			&d.Packaging,
			&d.Count,
			&d.UnitPrice,
			&d.Subtotal,
			&d.ExpiryDate,
			&d.LotNumber,
			&d.OroshiCode,
			&d.ReceiptNumber,
			&d.LineNumber,
		); err != nil {
			log.Println("scan DAT row:", err)
			continue
		}
		details = append(details, d)
	}

	// 3) USAGE records
	rows2, err := DB.Query(`
    SELECT
      usageYjCode       AS yj,
      usageDate         AS date,
      '処方'            AS type,
      usageAmount       AS quantity,
      usageUnitName     AS unit,
      ''                AS packaging,
      ''                AS count,
      ''                AS unit_price,
      ''                AS subtotal,
      ''                AS expiry_date,
      ''                AS lot_number,
      ''                AS oroshi_code,
      ''                AS receipt_number,
      ''                AS line_number
    FROM usagerecords
    WHERE usageDate BETWEEN ? AND ?
  `, from, to)
	if err != nil {
		log.Println("aggregate USAGE query error:", err)
		http.Error(w, "DBエラー(USAGE)", http.StatusInternalServerError)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var d Detail
		if err := rows2.Scan(
			&d.YJ,
			&d.Date,
			&d.Type,
			&d.Quantity,
			&d.Unit,
			&d.Packaging,
			&d.Count,
			&d.UnitPrice,
			&d.Subtotal,
			&d.ExpiryDate,
			&d.LotNumber,
			&d.OroshiCode,
			&d.ReceiptNumber,
			&d.LineNumber,
		); err != nil {
			log.Println("scan USAGE row:", err)
			continue
		}
		details = append(details, d)
	}

	// 4) group by YJ
	groups := make(map[string][]Detail)
	for _, d := range details {
		groups[d.YJ] = append(groups[d.YJ], d)
	}

	// 5) sort each group by date
	for yj, list := range groups {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Date < list[j].Date
		})
		groups[yj] = list
	}
	log.Printf("aggregate: total rows = %d", len(details))

	// 6) return JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		log.Println("aggregate JSON encode error:", err)
	}
}

// ma0Handler returns {"productName": "..."} for a given YJ code.
func ma0Handler(w http.ResponseWriter, r *http.Request) {
	yj := r.URL.Query().Get("yj")
	if yj == "" {
		http.Error(w, "yj を指定してください", http.StatusBadRequest)
		return
	}

	// JCShms テーブルから JC018ShouhinMei を1件取得
	const sqlq = `
    SELECT JC018ShouhinMei
      FROM jcshms
     WHERE JC009YJCode = ?
     LIMIT 1;
  `
	var name string
	if err := DB.QueryRow(sqlq, yj).Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			name = "" // 見つからなければ空文字
		} else {
			log.Println("ma0Handler query error:", err)
			http.Error(w, "DBエラー", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"productName": name})
}
