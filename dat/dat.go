// File: dat/dat.go
package dat

import (
	"bufio"
	"io"
	"log"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"YAMATO/ma0"
)

// DATRecord は DAT ファイルの1行をパースした結果を保持します。
type DATRecord struct {
	DatOroshiCode         string // S行の卸コード
	DatDate               string // D行の日付
	DatDeliveryFlag       string // D行のデリバリフラグ
	DatReceiptNumber      string // D行の受領番号
	DatLineNumber         string // D行の行番号
	DatJanCode            string // D行のJANコード
	DatProductName        string // D行の商品名
	DatQuantity           string // D行の数量
	DatUnitPrice          string // D行の単価
	DatSubtotal           string // D行の小計
	DatPackagingDrugPrice string // D行の包装医薬品価格
	DatExpiryDate         string // D行の賞味期限
	DatLotNumber          string // D行のロット番号
}

// ProcessDATRecord は、スライス化したレコードデータで MA0 を作成／参照します。
func ProcessDATRecord(data []string) (bool, error) {
	if len(data) < 5 {
		return false, nil
	}
	jan := data[4]
	_, created, err := ma0.CheckOrCreateMA0(jan)
	if err != nil {
		return false, err
	}
	if created {
		log.Printf("[DAT] New MA0 record: JAN=%q", jan)
	}
	return created, nil
}

// ParseDATFile は io.Reader から DAT を読み込み、各レコードを DATRecord にして返します。
// 同時に ProcessDATRecord で MA0 の登録も行い、件数も集計します。
func ParseDATFile(r io.Reader) (records []DATRecord, totalCount, ma0CreatedCount, duplicateCount int, err error) {
	scanner := bufio.NewScanner(r)
	var currentOroshiCode string

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		if strings.HasPrefix(line, "S20") {
			currentOroshiCode = strings.TrimSpace(line[3:14])
			continue
		}
		if !strings.HasPrefix(line, "D20") {
			continue
		}
		totalCount++

		// フィールド抽出
		get := func(s string, start, end int) string {
			if len(s) >= end {
				return s[start:end]
			} else if len(s) > start {
				return s[start:]
			}
			return ""
		}
		datDate := get(line, 4, 12)
		datFlag := get(line, 3, 4)
		datRecNo := get(line, 12, 22)
		datLineNo := get(line, 22, 24)
		datJan := get(line, 25, 38)
		rawName := get(line, 38, 78)
		name, _, convErr := transform.String(japanese.ShiftJIS.NewDecoder(), rawName)
		if convErr != nil {
			name = rawName
		}
		datQty := get(line, 78, 83)
		datUnit := get(line, 83, 92)
		datSub := get(line, 92, 101)
		datPkg := get(line, 101, 109)
		datExp := get(line, 109, 115)
		datLot := get(line, 115, 121)

		rec := DATRecord{
			DatOroshiCode:         currentOroshiCode,
			DatDate:               datDate,
			DatDeliveryFlag:       datFlag,
			DatReceiptNumber:      datRecNo,
			DatLineNumber:         datLineNo,
			DatJanCode:            datJan,
			DatProductName:        name,
			DatQuantity:           datQty,
			DatUnitPrice:          datUnit,
			DatSubtotal:           datSub,
			DatPackagingDrugPrice: datPkg,
			DatExpiryDate:         datExp,
			DatLotNumber:          datLot,
		}
		records = append(records, rec)

		// MA0連携
		created, procErr := ProcessDATRecord([]string{
			currentOroshiCode,
			datDate,
			datFlag,
			datRecNo,
			datJan,
			datLineNo,
			name,
			datQty,
			datUnit,
			datSub,
			datPkg,
			datExp,
			datLot,
		})
		if procErr != nil {
			return records, totalCount, ma0CreatedCount, duplicateCount, procErr
		}
		if created {
			ma0CreatedCount++
		} else {
			duplicateCount++
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
	}
	return
}
