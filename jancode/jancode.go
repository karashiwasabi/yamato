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

// QueryByJan は JAN コードを受け取り、該当レコードを返します
func QueryJANCODERecordsByJan(db *sql.DB, jan string) ([]JANCODERecord, error) {
	const sqlQuery = `
    SELECT


JA000,
JA001JanCode,
JA002,
JA003,
JA004,
JA005,
JA006HousouSuuryouSuuchi,
JA007HousouSuuryouTaniCode,
JA008HousouSouryouSuuchi,
JA009,
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
JA029
        FROM jancode
        WHERE JA001JanCode = ?
 `
	rows, err := db.Query(sqlQuery, jan)
	if err != nil {
		return nil, fmt.Errorf("jancode query error: %w", err)
	}
	defer rows.Close()

	var results []JANCODERecord
	for rows.Next() {
		var rec JANCODERecord
		if err := rows.Scan(
			&rec.JA000,
			&rec.JA001JanCode,
			&rec.JA002,
			&rec.JA003,
			&rec.JA004,
			&rec.JA005,
			&rec.JA006HousouSuuryouSuuchi,
			&rec.JA007HousouSuuryouTaniCode,
			&rec.JA008HousouSouryouSuuchi,
			&rec.JA009,
			&rec.JA010,
			&rec.JA011,
			&rec.JA012,
			&rec.JA013,
			&rec.JA014,
			&rec.JA015,
			&rec.JA016,
			&rec.JA017,
			&rec.JA018,
			&rec.JA019,
			&rec.JA020,
			&rec.JA021,
			&rec.JA022,
			&rec.JA023,
			&rec.JA024,
			&rec.JA025,
			&rec.JA026,
			&rec.JA027,
			&rec.JA028,
			&rec.JA029,
		); err != nil {
			return nil, fmt.Errorf("jancode scan error: %w", err)
		}
		results = append(results, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jancode rows error: %w", err)
	}
	return results, nil
}

// QueryByJan は QueryJANCODERecordsByJan の alias
func QueryByJan(db *sql.DB, jan string) ([]JANCODERecord, error) {
	return QueryJANCODERecordsByJan(db, jan)
}
