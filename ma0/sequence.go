package ma0

import (
	"database/sql"
	"fmt"
)

// NextSequence は prefix（"MA1Y"|"MA2Y"|"MA2J"）ごとに
// 8桁ゼロパディング連番を発行します。
func NextSequence(db *sql.DB, prefix string) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var lastNo int
	if err = tx.QueryRow(
		`SELECT last_no FROM code_sequences WHERE name = ?`,
		prefix,
	).Scan(&lastNo); err != nil {
		return "", fmt.Errorf("select last_no: %w", err)
	}

	lastNo++
	if _, err = tx.Exec(
		`UPDATE code_sequences SET last_no = ? WHERE name = ?`,
		lastNo, prefix,
	); err != nil {
		return "", fmt.Errorf("update last_no: %w", err)
	}

	seq := fmt.Sprintf("%s%08d", prefix, lastNo)
	return seq, nil
}
