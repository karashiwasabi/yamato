package ma0

import (
	"encoding/json"
	"net/http"
	"sync"
)

// MA0Record は MA0 の各レコード情報を保持する構造体です。
type MA0Record struct {
	JANCode          string `json:"janCode"`
	PackagingUnit    string `json:"packagingUnit"`
	ConversionFactor string `json:"conversionFactor"`
}

var (
	// ma0Data は in‑memory のマスター情報を保持します。
	ma0Data = make(map[string]MA0Record)
	// mutex により同時アクセスの排他制御を行います。
	mutex sync.Mutex
)

// CheckOrCreateMA0 は、指定された JAN コードについて、
// 既に登録済みかをチェックし、存在すればそのレコードと false を返し、
// 存在しなければダミー値（PackagingUnit: "錠", ConversionFactor: "100"）を用いて
// 新規に登録し、登録したレコードと true を返します。エラーは基本的に発生しません。
func CheckOrCreateMA0(jan string) (MA0Record, bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if rec, exists := ma0Data[jan]; exists {
		return rec, false, nil
	}

	newRec := MA0Record{
		JANCode:          jan,
		PackagingUnit:    "錠",   // 仮の値。実際はマスターCSVから取得するなどの処理を追加可能
		ConversionFactor: "100", // 仮の値
	}
	ma0Data[jan] = newRec
	return newRec, true, nil
}

// ViewMA0Handler は、現在の MA0 の中身（登録された MA0Record の一覧）を JSON 形式で返却します。
func ViewMA0Handler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	// ma0Data をスライスに変換して返す。
	var records []MA0Record
	for _, rec := range ma0Data {
		records = append(records, rec)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"MA0Records": records,
	})
}
