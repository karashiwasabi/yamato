// File: YAMATO/ma0/service.go
package ma0

import (
	"database/sql"
	"fmt"
)

// MARecord は MA2 登録に必要な情報を保持します。
type MARecord struct {
	JanCode                string // 元の JAN コード
	ProductName            string // 商品名
	HousouKeitai           string // 包装形態
	HousouTaniUnit         string // 包装単位
	HousouSouryouNumber    int    // 包装総量
	JanHousouSuuryouNumber int    // JAN 包装数量
	JanHousouSuuryouUnit   string // JAN 包装数量単位
	JanHousouSouryouNumber int    // JAN 包装総量
}

// YAMATO/ma0/service.go の RegisterMA をこんな風に書き換えます
func RegisterMA(db *sql.DB, maRec *MARecord) (janSeq, yjSeq string, err error) {
	// ① 既存レコードがあれば、そこで使われたYJを返す（JANは maRec.JanCode そのまま）
	if maRec.JanCode != "" {
		var existingYJ string
		err := db.QueryRow(
			"SELECT MA2YjCode FROM ma2 WHERE MA2JanCode = ?",
			maRec.JanCode,
		).Scan(&existingYJ)
		if err == nil {
			return maRec.JanCode, existingYJ, nil
		} else if err != sql.ErrNoRows {
			return "", "", fmt.Errorf("MA2 lookup error: %w", err)
		}
	}

	// ② 新規ケース：JANが空ならシーケンス発番
	if maRec.JanCode == "" {
		seq, seqErr := NextSequence(db, "MA2J")
		if seqErr != nil {
			return "", "", fmt.Errorf("MA2J seq error: %w", seqErr)
		}
		janSeq = seq
	} else {
		janSeq = maRec.JanCode
	}

	// ③ YJは常にシーケンス発番（存在チェック済なので重複発番ナシ）
	yjSeq, seqErr := NextSequence(db, "MA2Y")
	if seqErr != nil {
		return "", "", fmt.Errorf("MA2Y seq error: %w", seqErr)
	}

	// ④ INSERT OR IGNORE
	_, execErr := db.Exec(
		`INSERT OR IGNORE INTO ma2 
           (MA2JanCode,MA2YjCode,Shouhinmei,HousouKeitai,
            HousouTaniUnit,HousouSouryouNumber,
            JanHousouSuuryouNumber,JanHousouSuuryouUnit,JanHousouSouryouNumber)
         VALUES(?,?,?,?,?,?,?,?,?)`,
		janSeq, yjSeq,
		maRec.ProductName, maRec.HousouKeitai,
		maRec.HousouTaniUnit, maRec.HousouSouryouNumber,
		maRec.JanHousouSuuryouNumber, maRec.JanHousouSuuryouUnit,
		maRec.JanHousouSouryouNumber,
	)
	if execErr != nil {
		return "", "", fmt.Errorf("MA2 insert error: %w", execErr)
	}

	return janSeq, yjSeq, nil
}
