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

// Record は MA2 登録／更新の DTO です
type Record struct {
	JanCode                  string `json:"janCode"` // MA2JanCode
	YjCode                   string `json:"yjCode"`  // MA2YjCode
	Shouhinmei               string `json:"shouhinmei"`
	HousouKeitai             string `json:"housouKeitai"`
	HousouTaniUnitName       string `json:"housouTaniUnit"` // 入力は名称
	HousouSouryouNumber      int    `json:"housouSouryouNumber"`
	JanHousouSuuryouNumber   int    `json:"janHousouSuuryouNumber"`
	JanHousouSuuryouUnitName string `json:"janHousouSuuryouUnit"` // 入力は名称
	JanHousouSouryouNumber   int    `json:"janHousouSouryouNumber"`
}

// Upsert は MA2 テーブルへの登録／更新を行います。
// - rec.JanCode/YjCode が空ならシーケンス発番 (新規)
// - 空でなければそのまま (更新)
func Upsert(db *sql.DB, rec *Record) error {
	// (1) 単位名称→コード変換マップを取得
	nameToCode := tani.BuildNameToCodeMap(usage.GetTaniMap())

	// (2) フィールド名→コードに置き換え
	if code, ok := nameToCode[strings.Trim(rec.HousouTaniUnitName, `"' `)]; ok {
		rec.HousouTaniUnitName = code
	}
	if code, ok := nameToCode[strings.Trim(rec.JanHousouSuuryouUnitName, `"' `)]; ok {
		rec.JanHousouSuuryouUnitName = code
	}

	// (3) 新規か更新か判定してシーケンス発番
	var jaSeq, yjSeq string
	var err error

	if rec.JanCode != "" && rec.YjCode != "" {
		// 更新: 入力されたコードをそのまま使う
		jaSeq = rec.JanCode
		yjSeq = rec.YjCode
	} else {
		// 新規: MA0.RegisterMA で発番
		jaSeq, yjSeq, err = ma0.RegisterMA(db, &ma0.MARecord{
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
		rec.JanCode = jaSeq
		rec.YjCode = yjSeq
	}

	// (4) UPSERT INTO ma2
	stmt := `
INSERT OR REPLACE INTO ma2
  (MA2JanCode, MA2YjCode, Shouhinmei,
   HousouKeitai, HousouTaniUnit, HousouSouryouNumber,
   JanHousouSuuryouNumber, JanHousouSuuryouUnit, JanHousouSouryouNumber)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
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

// UpsertHandler は /api/ma2/upsert の HTTP ハンドラです
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
