// File: ma0/ma0.go
package ma0

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"YAMATO/jancode"
	"YAMATO/jchms"
)

type MA0Record struct {
	MA000JC000JanCode             string `json:"mA000JC000JanCode"`
	MA009JC009YJCode              string `json:"mA009JC009YJCode"`
	MA131JA007HousouSuuryouSuuchi string `json:"mA131JA007HousouSuuryouSuuchi"`
}

var (
	DB     *sql.DB
	cache  = make(map[string]MA0Record)
	cacheM sync.Mutex
)

// CheckOrCreateMA0 はキャッシュ→DB→マスター照会→INSERT を行います。
func CheckOrCreateMA0(jan string) (MA0Record, bool, error) {
	cacheM.Lock()
	defer cacheM.Unlock()

	if rec, ok := cache[jan]; ok {
		return rec, false, nil
	}
	// DB検索
	var rec MA0Record
	err := DB.QueryRow(`
        SELECT
            MA000JC000JanCode,
            MA009JC009YJCode,
            MA131JA007HousouSuuryouSuuchi
        FROM ma0
        WHERE MA000JC000JanCode = ?
    `, jan).Scan(
		&rec.MA000JC000JanCode,
		&rec.MA009JC009YJCode,
		&rec.MA131JA007HousouSuuryouSuuchi,
	)
	if err == nil {
		cache[jan] = rec
		return rec, false, nil
	}
	if err != sql.ErrNoRows {
		return MA0Record{}, false, err
	}

	// マスター照会
	jcRecs, _ := jchms.QueryJCHMASRecordsByJan(DB, jan)
	jaRecs, _ := jancode.QueryJANCODERecordsByJan(DB, jan)
	var yj, t string
	if len(jcRecs) > 0 {
		yj = jcRecs[0].JC.JC009YJCode
	}
	if len(jaRecs) > 0 {
		t = jaRecs[1].JA007HousouSuuryouSuuchi
	}
	newRec := MA0Record{jan, yj, t}

	if _, err := DB.Exec(`
        INSERT INTO ma0 (
            MA000JC000JanCode,
            MA009JC009YJCode,
            MA131JA007HousouSuuryouSuuchi
        ) VALUES (?, ?, ?)
    `, jan, yj, t); err != nil {
		return MA0Record{}, false, err
	}
	cache[jan] = newRec
	return newRec, true, nil
}

// ProcessMA0Record は DAT レコードを受け取り、Create時だけ出力します。
func ProcessMA0Record(data []string) error {
	if len(data) < 3 {
		return fmt.Errorf("insufficient data")
	}
	rec, created, err := CheckOrCreateMA0(data[2])
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("New MA0 created: %+v\n", rec)
	}
	return nil
}

// ViewMA0Handler はキャッシュ全件を JSON で返します。
func ViewMA0Handler(w http.ResponseWriter, r *http.Request) {
	cacheM.Lock()
	defer cacheM.Unlock()
	list := make([]MA0Record, 0, len(cache))
	for _, v := range cache {
		list = append(list, v)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(list)
}

func CountMA0() (int, error) {
	var cnt int
	err := DB.QueryRow("SELECT COUNT(*) FROM ma0").Scan(&cnt)
	return cnt, err
}
