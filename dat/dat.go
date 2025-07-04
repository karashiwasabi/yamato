// File: YAMATO/dat/dat.go
package dat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"YAMATO/jcshms"
	"YAMATO/ma0"
	"YAMATO/model"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// getOrganizedFlag は JAN が JCShms マスターにあれば1、なければ0を返します。
func getOrganizedFlag(jan string) (int, error) {
	recs, err := jcshms.QueryByJan(ma0.DB, jan)
	if err != nil {
		return 0, fmt.Errorf("jcshms.QueryByJan error: %w", err)
	}
	if len(recs) > 0 {
		return 1, nil
	}
	return 0, nil
}

// ParseDATFile は DAT ファイルを読み込み、
// model.DATRecord スライスと統計値を返します。
// MA0 未登録品はすべて MA2 テーブルに登録します。
func ParseDATFile(
	r io.Reader,
) (
	records []model.DATRecord,
	totalCount, ma0CreatedCount, duplicateCount int,
	err error,
) {
	scanner := bufio.NewScanner(r)
	var currentOroshiCode string

	// 固定長フィールド取得ヘルパー
	getField := func(s string, start, end int) string {
		if len(s) >= end {
			return s[start:end]
		} else if len(s) > start {
			return s[start:]
		}
		return ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		if strings.HasPrefix(line, "S20") {
			currentOroshiCode = strings.TrimSpace(getField(line, 3, 12))
			continue
		}
		if !strings.HasPrefix(line, "D20") {
			continue
		}
		totalCount++

		// DATRecord 組み立て
		datDate := getField(line, 4, 12)
		datFlag := getField(line, 3, 4)
		datRecNo := getField(line, 12, 22)
		datLineNo := getField(line, 22, 24)
		datJan := getField(line, 25, 38)
		rawName := getField(line, 38, 78)
		name, _, convErr := transform.String(japanese.ShiftJIS.NewDecoder(), rawName)
		if convErr != nil {
			name = rawName
		}
		datQty := getField(line, 78, 83)
		datUnit := getField(line, 83, 92)
		datSub := getField(line, 92, 101)
		datPkg := getField(line, 101, 109)
		datExp := getField(line, 109, 115)
		datLot := getField(line, 115, 121)

		rec := model.DATRecord{
			CurrentOroshiCode: currentOroshiCode,
			DatDate:           datDate,
			DatFlag:           datFlag,
			DatRecNo:          datRecNo,
			DatLineNo:         datLineNo,
			DatJan:            datJan,
			DatProductName:    name,
			DatQty:            datQty,
			DatUnit:           datUnit,
			DatSub:            datSub,
			DatPkg:            datPkg,
			DatExp:            datExp,
			DatLot:            datLot,
		}
		records = append(records, rec)

		// datrecords テーブル挿入＋organizedFlag 集計
		flag, fgErr := getOrganizedFlag(datJan)
		if fgErr != nil {
			log.Printf("[DAT] OrganizedFlag error JAN=%q: %v", datJan, fgErr)
			flag = 0
		}
		if err := ma0.InsertDATRecord(ma0.DB, rec, flag); err != nil {
			log.Printf("[DAT] InsertDATRecord error: %v", err)
		}
		if flag == 1 {
			// organized
		} else {
			duplicateCount++
		}

		// MA0 連携／MA2 登録
		ma0Rec, created, err0 := ma0.CheckOrCreateMA0(datJan, name)

		if err0 != nil {
			log.Printf("[DAT] MA0 lookup error JAN=%s: %v", datJan, err0)
		}
		if created {
			ma0CreatedCount++
		}
		// マスター未登録品は MA2 に登録
		if !created && ma0Rec.MA018JC018ShouhinMei == "" {
			hs, _ := strconv.Atoi(ma0Rec.MA044JC044HousouSouryouSuuchi)
			jsn, _ := strconv.Atoi(ma0Rec.MA131JA006HousouSuuryouSuuchi)
			jssn, _ := strconv.Atoi(ma0Rec.MA133JA008HousouSouryouSuuchi)
			mrec := &ma0.MARecord{
				JanCode:                datJan,
				ProductName:            name,
				HousouKeitai:           ma0Rec.MA037JC037HousouKeitai,
				HousouTaniUnit:         ma0Rec.MA038JC038HousouTaniSuuchi,
				HousouSouryouNumber:    hs,
				JanHousouSuuryouNumber: jsn,
				JanHousouSuuryouUnit:   ma0Rec.MA132JA007HousouSuuryouTaniCode,
				JanHousouSouryouNumber: jssn,
			}
			_, _, err2 := ma0.RegisterMA(ma0.DB, mrec)
			if err2 != nil {
				log.Printf("[DAT] MA2 registration error JAN=%s: %v", datJan, err2)
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
	}
	return
}
