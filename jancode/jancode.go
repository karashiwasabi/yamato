// File: jancode/jancode.go
package jancode

import (
	"database/sql"
	"fmt"
)

type JANCODERecord struct {
	JA001                      string
	JA002JanCode               string
	JA003                      string
	JA004                      string
	JA005                      string
	JA006                      string
	JA007HousouSuuryouSuuchi   string
	JA008HousouSuuryouTaniCode string
	JA009HousouSouryouSuuchi   string
	JA010                      string
	JA011                      string
	JA012                      string
	JA013                      string
	JA014                      string
	JA015                      string
	JA016                      string
	JA017                      string
	JA018                      string
	JA019                      string
	JA020                      string
	JA021                      string
	JA022                      string
	JA023                      string
	JA024                      string
	JA025                      string
	JA026                      string
	JA027                      string
	JA028                      string
	JA029                      string
	JA030                      string
}

func QueryJANCODERecordsByJan(db *sql.DB, jan string) ([]JANCODERecord, error) {
	query := `
        SELECT 
JA001,
JA002JanCode,
JA003,
JA004,
JA005,
JA006,
JA007HousouSuuryouSuuchi,
JA008HousouSuuryouTaniCode,
JA009HousouSouryouSuuchi,
JA010,
JA011,
JA012,
JA013,
JA014,
JA015,
JA016,
JA017,
JA018,
JA019,
JA020,
JA021,
JA022,
JA023,
JA024,
JA025,
JA026,
JA027,
JA028,
JA029,
JA030
        FROM jancode
        WHERE JA002JanCode = ?
    `
	rows, err := db.Query(query, jan)
	if err != nil {
		return nil, fmt.Errorf("jancode query error: %v", err)
	}
	defer rows.Close()

	var records []JANCODERecord
	for rows.Next() {
		var rec JANCODERecord
		if err := rows.Scan(&rec.JA002JanCode, &rec.JA007HousouSuuryouSuuchi); err != nil {
			return nil, fmt.Errorf("jancode row scan error: %v", err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jancode rows error: %v", err)
	}
	return records, nil
}
