package inout

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"YAMATO/ma0"
)

var DB *sql.DB

// InoutRecord は得意先レコード
type InoutRecord struct {
	InoutCode  string `json:"inoutcode"`
	Name       string `json:"name"`
	OroshiCode string `json:"oroshicode"`
}

// ProductRec は /api/inout/search の結果レコード
type ProductRec struct {
	Jan             string  `json:"jan"`             // JC000JanCode
	YJ              string  `json:"yj"`              // JC009YJCode
	Name            string  `json:"name"`            // JC018ShouhinMei
	Spec            string  `json:"spec"`            // JC020KikakuYouryou
	PackCount       float64 `json:"packCount"`       // JC038HousouTaniSuuchi
	PackTotal       float64 `json:"packTotal"`       // JC044HousouSouryouSuuchi
	Coef            float64 `json:"coef"`            // JC048HousouYakkaKeisuu
	UnitYaku        float64 `json:"unitYaku"`        // JC049GenTaniYakka
	PackagingUnit   string  `json:"unitName"`        // JC039HousouTaniTani
	PackQtyNumber   float64 `json:"packQtyNumber"`   // JA006HousouSuuryouSuuchi
	PackQtyUnitCode int     `json:"packQtyUnitCode"` // JA007HousouSuuryouTaniCode
}

func init() {
	http.HandleFunc("/api/inout/search", productSearchHandler)
}

// Handler は /api/inout の GET/POST を処理
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listInout(w)
	case http.MethodPost:
		saveInout(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// listInout は inout テーブル全件を JSON 返却
func listInout(w http.ResponseWriter) {
	rows, err := DB.Query(`
    SELECT inoutcode, name, oroshicode
      FROM inout
     ORDER BY inoutcode
  `)
	if err != nil {
		log.Printf("[inout] list error: %v", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []InoutRecord
	for rows.Next() {
		var rec InoutRecord
		if err := rows.Scan(&rec.InoutCode, &rec.Name, &rec.OroshiCode); err != nil {
			log.Printf("[inout] scan error: %v", err)
			continue
		}
		list = append(list, rec)
	}
	writeJSON(w, list)
}

// saveInout は新規得意先を登録して返却
func saveInout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		OroshiCode string `json:"oroshicode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	seq, err := ma0.NextSequence(DB, "INOUT")
	if err != nil {
		log.Printf("[inout] seq error: %v", err)
		http.Error(w, "Sequence Error", http.StatusInternalServerError)
		return
	}
	if _, err := DB.Exec(
		`INSERT INTO inout(inoutcode, name, oroshicode) VALUES(?, ?, ?)`,
		seq, req.Name, req.OroshiCode,
	); err != nil {
		log.Printf("[inout] insert error: %v", err)
		http.Error(w, "Insert Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, InoutRecord{InoutCode: seq, Name: req.Name, OroshiCode: req.OroshiCode})
}

// productSearchHandler は薬品検索＋包装情報を返す
func productSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := "%" + q.Get("name") + "%"
	spec := "%" + q.Get("spec") + "%"

	rows, err := DB.Query(`
    SELECT
      j.JC000JanCode,
      j.JC009YJCode,
      j.JC018ShouhinMei,
      j.JC020KikakuYouryou,
      j.JC038HousouTaniSuuchi,
      j.JC044HousouSouryouSuuchi,
      j.JC048HousouYakkaKeisuu,
      j.JC049GenTaniYakka,
      j.JC039HousouTaniTani,
      m2.JA006HousouSuuryouSuuchi   AS packQtyNumber,
      m2.JA007HousouSuuryouTaniCode AS packQtyUnitCode
    FROM jcshms AS j
LEFT JOIN jancode AS m2
      ON j.JC000JanCode = m2.JA001JanCode
   WHERE j.JC018ShouhinMei LIKE ?
     AND j.JC020KikakuYouryou LIKE ?
   LIMIT 100
  `, name, spec)
	if err != nil {
		log.Printf("[inout] search error: %v", err)
		http.Error(w, "Search Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []ProductRec
	for rows.Next() {
		var p ProductRec
		if err := rows.Scan(
			&p.Jan, &p.YJ, &p.Name, &p.Spec,
			&p.PackCount, &p.PackTotal, &p.Coef, &p.UnitYaku, &p.PackagingUnit,
			&p.PackQtyNumber, &p.PackQtyUnitCode,
		); err != nil {
			log.Printf("[inout] scan error: %v", err)
			continue
		}
		out = append(out, p)
	}
	writeJSON(w, out)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
