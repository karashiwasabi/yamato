// File: YAMATO/ma2/ma2.go
package ma2

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"YAMATO/ma0"
	"YAMATO/tani"
	"YAMATO/usage"
)

// Record は MA2 登録／更新に必要なフィールド群です。
// CSVインポート／JSONデコード後にこの型へ詰め替えて使います。
type Record struct {
	JanCode                  string `json:"janCode"`
	YjCode                   string `json:"yjCode"`
	Shouhinmei               string `json:"shouhinmei"`
	HousouKeitai             string `json:"housouKeitai"`
	HousouTaniUnitName       string `json:"housouTaniUnit"`
	HousouSouryouNumber      int    `json:"housouSouryouNumber"`
	JanHousouSuuryouNumber   int    `json:"janHousouSuuryouNumber"`
	JanHousouSuuryouUnitName string `json:"janHousouSuuryouUnit"`
	JanHousouSouryouNumber   int    `json:"janHousouSouryouNumber"`
}

// Upsert は MA2 テーブルへの登録／更新を行います。
// - 「名称→コード」マッピング
// - MA2JanCode／MA2YjCode の発番 or 取得
// - INSERT OR REPLACE INTO ma2
func Upsert(db *sql.DB, rec *Record) error {
	// 1) 単位名称→コードマップ
	nameToCode := tani.BuildNameToCodeMap(usage.GetTaniMap())

	// 2) 名称→コードへの置換
	if code, ok := nameToCode[strings.Trim(rec.HousouTaniUnitName, `"' `)]; ok {
		rec.HousouTaniUnitName = code
	}
	if code, ok := nameToCode[strings.Trim(rec.JanHousouSuuryouUnitName, `"' `)]; ok {
		rec.JanHousouSuuryouUnitName = code
	}

	// 3) MA0.RegisterMA で MA2JanCode／MA2YjCode を発番または取得
	jaSeq, yjSeq, err := ma0.RegisterMA(db, &ma0.MARecord{
		JanCode:                rec.JanCode,
		ProductName:            rec.Shouhinmei,
		HousouKeitai:           rec.HousouKeitai,
		HousouTaniUnit:         rec.HousouTaniUnitName,
		HousouSouryouNumber:    rec.HousouSouryouNumber,
		JanHousouSuuryouNumber: rec.JanHousouSuuryouNumber,
		JanHousouSuuryouUnit:   rec.JanHousouSuuryouUnitName,
		JanHousouSouryouNumber: rec.JanHousouSouryouNumber,
	})
	if err != nil {
		return fmt.Errorf("MA2 シーケンス発番エラー: %w", err)
	}
	// 更新時は jaSeq が空のままなので既存値をそのまま利用
	if rec.JanCode == "" {
		rec.JanCode = jaSeq
	}
	rec.YjCode = yjSeq

	// 4) ma2 テーブルへの UPSERT
	stmt := `
    INSERT OR REPLACE INTO ma2
      (MA2JanCode, MA2YjCode, Shouhinmei,
       HousouKeitai, HousouTaniUnit, HousouSouryouNumber,
       JanHousouSuuryouNumber, JanHousouSuuryouUnit, JanHousouSouryouNumber)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `
	if _, err := db.Exec(
		stmt,
		rec.JanCode, rec.YjCode, rec.Shouhinmei,
		rec.HousouKeitai, rec.HousouTaniUnitName, rec.HousouSouryouNumber,
		rec.JanHousouSuuryouNumber, rec.JanHousouSuuryouUnitName, rec.JanHousouSouryouNumber,
	); err != nil {
		return fmt.Errorf("ma2 UPSERT エラー: %w", err)
	}
	return nil
}

// UpsertHandler は /api/ma2/upsert 向け HTTP ハンドラです。
// JSON デコード → Upsert → 結果返却 を行います。
func UpsertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec Record
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := Upsert(ma0.DB, &rec); err != nil {
		http.Error(w, "upsert error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(rec)
}
