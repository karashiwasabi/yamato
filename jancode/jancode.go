// File: jancode/jancode.go
package jancode

import (
	"database/sql"
	"fmt"
)

type JANCODERecord struct {
	JA000                      string
	JA001JanCode               string
	JA002                      string
	JA003                      string
	JA004                      string
	JA005                      string
	JA006HousouSuuryouSuuchi   string
	JA007HousouSuuryouTaniCode string
	JA008HousouSouryouSuuchi   string
	JA009                      string
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
}

func QueryJANCODERecordsByJan(db *sql.DB, jan string) ([]JANCODERecord, error) {
	query := `
        SELECT 
JA001JanCode,
JA006HousouSuuryouSuuchi
        FROM jancode
        WHERE JA001JanCode = ?
    `
	rows, err := db.Query(query, jan)
	if err != nil {
		return nil, fmt.Errorf("jancode query error: %v", err)
	}
	defer rows.Close()

	var records []JANCODERecord
	for rows.Next() {
		var rec JANCODERecord
		if err := rows.Scan(&rec.JA001JanCode, &rec.JA006HousouSuuryouSuuchi); err != nil {
			return nil, fmt.Errorf("jancode row scan error: %v", err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jancode rows error: %v", err)
	}
	return records, nil
}
