// File: YAMATO/ma0/service.go
package ma0

import (
	"database/sql"
	"fmt"
)

// MARecord は MA2 登録に必要な情報を保持します。
type MARecord struct {
	JanCode                string // ここに元のJANコードをセットできる
	ProductName            string // 商品名
	HousouKeitai           string // 包装形態
	HousouTaniUnit         string // 包装単位
	HousouSouryouNumber    int    // 包装数量
	JanHousouSuuryouNumber int    // JAN包装数量
	JanHousouSuuryouUnit   string // JAN包装単位
	JanHousouSouryouNumber int    // JAN包装形態別数量
}

// RegisterMA は MA0マスターにヒットしなかったレコードを MA2 に登録します。
// JanCode が空でなければそれをそのまま MA2JanCode に使い、重複を防止します。
func RegisterMA(db *sql.DB, maRec *MARecord) error {
	// MA2JanCode を決定
	var janSeq string
	if maRec.JanCode != "" {
		janSeq = maRec.JanCode
	} else {
		s, err := NextSequence(db, "MA2J")
		if err != nil {
			return fmt.Errorf("failed to get MA2J sequence: %w", err)
		}
		janSeq = s
	}

	// YJ は常にシーケンス発行
	yjSeq, err := NextSequence(db, "MA2Y")
	if err != nil {
		return fmt.Errorf("failed to get MA2Y sequence: %w", err)
	}

	// INSERT OR IGNORE なら重複レコードはスキップ
	_, err = db.Exec(
		`INSERT OR IGNORE INTO ma2
      (MA2JanCode, MA2YjCode, Shouhinmei, HousouKeitai,
       HousouTaniUnit, HousouSouryouNumber,
       JanHousouSuuryouNumber, JanHousouSuuryouUnit,
       JanHousouSouryouNumber)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		janSeq, yjSeq,
		maRec.ProductName,
		maRec.HousouKeitai,
		maRec.HousouTaniUnit,
		maRec.HousouSouryouNumber,
		maRec.JanHousouSuuryouNumber,
		maRec.JanHousouSuuryouUnit,
		maRec.JanHousouSouryouNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into ma2: %w", err)
	}
	return nil
}
