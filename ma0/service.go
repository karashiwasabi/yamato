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

// RegisterMA は MA0 マスターにヒットしなかったレコードを MA2 に登録します。
// JanCode が空でなければそれを MA2JanCode に、常にシーケンスで MA2YjCode を発行します。
// 返り値として発番済みの janSeq, yjSeq を返却します。
func RegisterMA(db *sql.DB, maRec *MARecord) (janSeq, yjSeq string, err error) {
	// MA2JanCode を決定
	if maRec.JanCode != "" {
		janSeq = maRec.JanCode
	} else {
		seq, seqErr := NextSequence(db, "MA2J")
		if seqErr != nil {
			err = fmt.Errorf("failed to get MA2J sequence: %w", seqErr)
			return
		}
		janSeq = seq
	}

	// YJ は常にシーケンス発行
	seqYJ, seqErr := NextSequence(db, "MA2Y")
	if seqErr != nil {
		err = fmt.Errorf("failed to get MA2Y sequence: %w", seqErr)
		return
	}
	yjSeq = seqYJ

	// INSERT OR IGNORE で重複をスキップ
	_, execErr := db.Exec(
		`INSERT OR IGNORE INTO ma2
           (MA2JanCode, MA2YjCode, Shouhinmei, HousouKeitai,
            HousouTaniUnit, HousouSouryouNumber,
            JanHousouSuuryouNumber, JanHousouSuuryouUnit,
            JanHousouSouryouNumber)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		janSeq,
		yjSeq,
		maRec.ProductName,
		maRec.HousouKeitai,
		maRec.HousouTaniUnit,
		maRec.HousouSouryouNumber,
		maRec.JanHousouSuuryouNumber,
		maRec.JanHousouSuuryouUnit,
		maRec.JanHousouSouryouNumber,
	)
	if execErr != nil {
		err = fmt.Errorf("failed to insert into ma2: %w", execErr)
		return
	}

	return
}
