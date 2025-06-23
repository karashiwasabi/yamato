// File: ma0/ma0.go
package ma0

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"YAMATO/jancode"
	"YAMATO/jcshms"
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

// CheckOrCreateMA0 はキャッシュ→DB→マスター照会→INSERT の順で動作します。
func CheckOrCreateMA0(jan string) (MA0Record, bool, error) {
	cacheM.Lock()
	defer cacheM.Unlock()

	log.Printf("[ma0] ▶ CheckOrCreateMA0 start: JAN=%s", jan)

	// キャッシュチェック
	if rec, ok := cache[jan]; ok {
		log.Printf("[ma0] ⇨ cache hit: %+v", rec)
		logCount()
		return rec, false, nil
	}

	// 永続テーブル検索
	var rec MA0Record
	err := DB.QueryRow(`
        SELECT MA000JC000JanCode,
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
		log.Printf("[ma0] ⇨ found in DB: %+v", rec)
		cache[jan] = rec
		logCount()
		return rec, false, nil
	}
	if err != sql.ErrNoRows {
		return MA0Record{}, false, fmt.Errorf("DB query error: %v", err)
	}
	log.Printf("[ma0] ⇨ not found in ma0, querying masters")

	// JCSHMSマスター照会
	csRecs, err := jcshms.QueryJCSHMSRecordsByJan(DB, jan)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("jcshms query error: %v", err)
	}
	// JANCODEマスター照会
	jaRecs, err := jancode.QueryJANCODERecordsByJan(DB, jan)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("jancode query error: %v", err)
	}
	log.Printf("[ma0] ⇨ master counts: JCSHMS=%d rows, JANCODE=%d rows", len(csRecs), len(jaRecs))

	// 値の組み立て
	var yj, t string
	if len(csRecs) > 0 {
		yj = csRecs[0].JC.JC009YJCode
	}
	if len(jaRecs) > 1 {
		t = jaRecs[1].JA007HousouSuuryouSuuchi
	} else if len(jaRecs) > 0 {
		t = jaRecs[0].JA007HousouSuuryouSuuchi
	}
	log.Printf("[ma0] ⇨ about to insert: jan=%q, yj=%q, t=%q", jan, yj, t)

	// INSERT
	res, err := DB.Exec(`
        INSERT INTO ma0 (
            MA000JC000JanCode,
            MA009JC009YJCode,
            MA131JA007HousouSuuryouSuuchi
        ) VALUES (?, ?, ?)
    `, jan, yj, t)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("insert error: %v", err)
	}
	cnt, _ := res.RowsAffected()
	log.Printf("[ma0] ⇨ inserted rows: %d", cnt)

	newRec := MA0Record{jan, yj, t}
	cache[jan] = newRec
	logCount()
	return newRec, true, nil
}

// ProcessMA0Record は DAT レコードごとに呼ばれます。
func ProcessMA0Record(data []string) error {
	if len(data) < 3 {
		return fmt.Errorf("insufficient DAT data: %v", data)
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

// ViewMA0Handler はキャッシュの内容を JSON で返却します。
func ViewMA0Handler(w http.ResponseWriter, r *http.Request) {
	cacheM.Lock()
	defer cacheM.Unlock()

	list := make([]MA0Record, 0, len(cache))
	for _, rec := range cache {
		list = append(list, rec)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(list)
}

// CountMA0 は ma0 テーブルの件数を返します。
func CountMA0() (int, error) {
	var cnt int
	err := DB.QueryRow("SELECT COUNT(*) FROM ma0").Scan(&cnt)
	return cnt, err
}

func logCount() {
	cnt, err := CountMA0()
	if err != nil {
		log.Printf("[ma0] count error: %v", err)
		return
	}
	log.Printf("[ma0] current ma0 count: %d", cnt)
}
