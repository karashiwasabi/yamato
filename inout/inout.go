// File: YAMATO/inout/inout.go
package inout

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"YAMATO/ma0"
)

var DB *sql.DB

// InoutRecord は得意先マスターを表します
type InoutRecord struct {
	InoutCode  string `json:"inoutcode"`
	Name       string `json:"name"`
	OroshiCode string `json:"oroshicode"`
}

// ProductRec は /api/inout/search の結果レコードです
type ProductRec struct {
	YJ              string  `json:"yj"`
	Jan             string  `json:"jan"`
	Name            string  `json:"name"`
	Spec            string  `json:"spec"`
	PackQtyNumber   float64 `json:"packQtyNumber"`
	PackQtyUnitCode int     `json:"packQtyUnitCode"`
	PackTotal       float64 `json:"packTotal"`
	Coef            float64 `json:"coef"`
	UnitName        string  `json:"unitName"`
	UnitYaku        float64 `json:"unitYaku"`
}

// IODRecord は出庫・入庫明細DTOです
type IODRecord struct {
	IodJan           string  `json:"iodJan"`
	IodDate          string  `json:"iodDate"`
	IodType          int     `json:"iodType"`
	IodJanQuantity   float64 `json:"iodJanQuantity"`
	IodJanUnit       string  `json:"iodJanUnit"`
	IodQuantity      float64 `json:"iodQuantity"`
	IodUnit          string  `json:"iodUnit"`
	IodPackaging     string  `json:"iodPackaging"`
	IodUnitPrice     float64 `json:"iodUnitPrice"`
	IodSubtotal      float64 `json:"iodSubtotal"`
	IodExpiryDate    string  `json:"iodExpiryDate"`
	IodLotNumber     string  `json:"iodLotNumber"`
	IodOroshiCode    string  `json:"iodOroshiCode"`
	IodReceiptNumber string  `json:"iodReceiptNumber"`
	IodLineNumber    int     `json:"iodLineNumber"`
}

// Handler は /api/inout の GET/POST を処理します
func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listClients(w)
	case http.MethodPost:
		saveClient(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// listClients は得意先一覧を返却します
func listClients(w http.ResponseWriter) {
	rows, err := DB.Query(`SELECT inoutcode, name, oroshicode FROM inout ORDER BY inoutcode`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []InoutRecord
	for rows.Next() {
		var rec InoutRecord
		if err := rows.Scan(&rec.InoutCode, &rec.Name, &rec.OroshiCode); err != nil {
			log.Println("inout scan error:", err)
			continue
		}
		out = append(out, rec)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(out)
}

// saveClient は POST で送られてきた得意先を inout テーブルに登録します
func saveClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		OroshiCode string `json:"oroshicode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// シーケンス発行
	seq, err := ma0.NextSequence(DB, "INOUT")
	if err != nil {
		log.Println("INOUT sequence error:", err)
		http.Error(w, "Sequence Error", http.StatusInternalServerError)
		return
	}

	// inout テーブルに挿入
	if _, err := DB.Exec(
		`INSERT INTO inout(inoutcode, name, oroshicode) VALUES(?,?,?)`,
		seq, req.Name, req.OroshiCode,
	); err != nil {
		log.Println("inout insert error:", err)
		http.Error(w, "Insert Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ProductSearchHandler は /api/inout/search の GET リクエストを処理します。
func ProductSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := "%" + q.Get("name") + "%"
	spec := "%" + q.Get("spec") + "%"

	rows, err := DB.Query(`
      SELECT
        j.JC009YJCode,
        j.JC000JanCode,
        j.JC018ShouhinMei,
        j.JC020KikakuYouryou AS spec,
        m2.JA006HousouSuuryouSuuchi   AS packQtyNumber,
        m2.JA007HousouSuuryouTaniCode AS packQtyUnitCode,
        j.JC044HousouSouryouSuuchi    AS packTotal,
        COALESCE(NULLIF(j.JC048HousouYakkaKeisuu, ''), '0') AS coef,
        j.JC039HousouTaniTani         AS unitName,
        COALESCE(NULLIF(j.JC049GenTaniYakka, ''), '0') AS unitYaku
      FROM jcshms AS j
      LEFT JOIN jancode AS m2
        ON j.JC000JanCode = m2.JA001JanCode
      WHERE (
          j.JC018ShouhinMei           LIKE ?   -- 商品名
       OR j.JC022ShouhinMeiKanaSortYou LIKE ?   -- かなソート用商品名
      )
        AND j.JC020KikakuYouryou LIKE ?        -- 規格
      LIMIT 100
    `, name, name, spec) // ← name を2回渡すのを忘れずに

	if err != nil {
		log.Printf("▶ ProductSearch SQL error: %v", err)
		http.Error(w, "DB Query Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []ProductRec
	for rows.Next() {
		var p ProductRec
		if err := rows.Scan(
			&p.YJ,
			&p.Jan,
			&p.Name,
			&p.Spec,
			&p.PackQtyNumber,
			&p.PackQtyUnitCode,
			&p.PackTotal,
			&p.Coef,
			&p.UnitName,
			&p.UnitYaku,
		); err != nil {
			log.Printf("▶ ProductSearch scan error: %v", err)
			continue
		}
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("▶ ProductSearch rows error: %v", err)
		http.Error(w, "DB Rows Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(out)
}

// SaveIODHandler は /api/inout/save で明細を受け取り DB に登録します
func SaveIODHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディをデコード
	var recs []IODRecord
	if err := json.NewDecoder(r.Body).Decode(&recs); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Printf("SaveIOD payload: %+v\n", recs)

	tx, err := DB.Begin()
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 登録用ステートメント
	stmt, err := tx.Prepare(`
        INSERT OR REPLACE INTO iod (
          iodJan, iodDate, iodType,
          iodJanQuantity, iodJanUnit,
          iodQuantity, iodUnit,
          iodPackaging, iodUnitPrice, iodSubtotal,
          iodExpiryDate, iodLotNumber,
          iodOroshiCode, iodReceiptNumber, iodLineNumber
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		http.Error(w, "Prepare Error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, v := range recs {
		// JAN or 数量が空/0 のものはスキップ
		if v.IodJan == "" || v.IodQuantity == 0 {
			continue
		}

		// ← MA0 連携／MA2 登録ロジックを追加する箇所
		maRec, created, err0 := ma0.CheckOrCreateMA0(v.IodJan)
		if err0 != nil {
			log.Printf("[IOD] MA0 lookup error JAN=%s: %v", v.IodJan, err0)
		} else if created {
			log.Printf("[IOD] MA0 record created JAN=%s → YJ=%s", v.IodJan, maRec.MA009JC009YJCode)
		}
		// ここまで

		// 実際の iod テーブルへの登録
		if _, err := stmt.Exec(
			v.IodJan, v.IodDate, v.IodType,
			v.IodJanQuantity, v.IodJanUnit,
			v.IodQuantity, v.IodUnit,
			v.IodPackaging, v.IodUnitPrice, v.IodSubtotal,
			v.IodExpiryDate, v.IodLotNumber,
			v.IodOroshiCode, v.IodReceiptNumber, v.IodLineNumber,
		); err != nil {
			log.Println("iod insert error:", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
