package dat

import (
	"bufio"
	"io"
	"log"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"YAMATO/ma0" // モジュール名は "YAMATO"（または仕様に合わせ "wasabi" ）固定
)

// DATRecord は、DAT ファイルから抽出した各フィールドの値を保持する構造体です。
type DATRecord struct {
	DatOroshiCode         string // S 行で取得した卸コード（raw[3:14]）
	DatDate               string // D 行から raw[4:12]（YYYYMMDD形式）
	DatDeliveryFlag       string // D 行から raw[3:4]
	DatReceiptNumber      string // D 行から raw[12:22]
	DatLineNumber         string // D 行から raw[22:24]
	DatJanCode            string // D 行から raw[25:38] ※仕様書準拠の13バイト
	DatProductName        string // D 行から raw[38:78] → 必要部分のみ Shift‑JIS→UTF‑8 変換適用
	DatQuantity           string // D 行から raw[78:83]
	DatUnitPrice          string // D 行から raw[83:92]
	DatSubtotal           string // D 行から raw[92:101]
	DatPackagingDrugPrice string // D 行から raw[101:109]
	DatExpiryDate         string // D 行から raw[109:115]
	DatLotNumber          string // D 行から raw[115:121]
}

// ProcessDATRecord は、作成した 13 項目のスライスを MA0.go へ送るための関数です。
// ここでは、スライスの index 4 が JAN コードであると仮定し、
// その JAN コードを利用して MA0 の重複チェック／新規登録処理（CheckOrCreateMA0）を呼び出しています。
func ProcessDATRecord(data []string) error {
	if len(data) < 5 {
		return nil // 想定外のケースでは何もしない
	}
	jan := data[4] // index 4 が JAN コード
	_, created, err := ma0.CheckOrCreateMA0(jan)
	if err != nil {
		return err
	}
	if created {
		log.Printf("[DAT] MA0 新規レコード作成: JAN=%q, 全データ=%#v", jan, data)
	} else {
		log.Printf("[DAT] 既存の MA0 レコード利用: JAN=%q, 全データ=%#v", jan, data)
	}
	return nil
}

// ParseDATFile は、io.Reader から DAT ファイルの内容を読み込み、
// S 行で取得された卸コードを D 行の各レコードに適用しながら、
// 各 D 行の値を DATRecord に変換するとともに、
// 次工程へ送るための 13 項目のスライス（recordData）の生成と ProcessDATRecord の呼び出しを行います。
// また、総件数、MA0 の新規登録件数、重複件数をカウントして返します。
func ParseDATFile(r io.Reader) (records []DATRecord, totalCount, ma0CreatedCount, duplicateCount int, err error) {
	scanner := bufio.NewScanner(r)
	var currentOroshiCode string = ""
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		// S 行（識別子 "S20"）の場合、卸コードを抽出して保存する
		if strings.HasPrefix(line, "S20") {
			currentOroshiCode = strings.TrimSpace(line[3:12])
			continue
		}
		// D 行（識別子 "D20"）のみを対象とする
		if !strings.HasPrefix(line, "D20") {
			continue
		}
		totalCount++

		// 固定長フィールド抽出のヘルパー関数
		getField := func(s string, start, end int) string {
			if len(s) >= end {
				return s[start:end]
			}
			if len(s) > start {
				return s[start:]
			}
			return ""
		}

		// 各フィールドを仕様書に基づく位置から抽出する
		datDate := getField(line, 4, 12)
		datDeliveryFlag := getField(line, 3, 4)
		datReceiptNumber := getField(line, 12, 22)
		datLineNumber := getField(line, 22, 24)
		datJanCode := getField(line, 25, 38)
		// 商品名だけは Shift‑JIS→UTF‑8 変換を適用する
		rawProductName := getField(line, 38, 78)
		datProductName, _, errConv := transform.String(japanese.ShiftJIS.NewDecoder(), rawProductName)
		if errConv != nil {
			log.Printf("[DAT] 商品名変換エラー: %v", errConv)
			datProductName = rawProductName
		}
		datQuantity := getField(line, 78, 83)
		datUnitPrice := getField(line, 83, 92)
		datSubtotal := getField(line, 92, 101)
		datPackagingDrugPrice := getField(line, 101, 109)
		datExpiryDate := getField(line, 109, 115)
		datLotNumber := getField(line, 115, 121)

		// 作成した各フィールドを DATRecord 構造体にまとめる
		record := DATRecord{
			DatOroshiCode:         currentOroshiCode,
			DatDate:               datDate,
			DatDeliveryFlag:       datDeliveryFlag,
			DatReceiptNumber:      datReceiptNumber,
			DatLineNumber:         datLineNumber,
			DatJanCode:            datJanCode,
			DatProductName:        datProductName,
			DatQuantity:           datQuantity,
			DatUnitPrice:          datUnitPrice,
			DatSubtotal:           datSubtotal,
			DatPackagingDrugPrice: datPackagingDrugPrice,
			DatExpiryDate:         datExpiryDate,
			DatLotNumber:          datLotNumber,
		}
		records = append(records, record)

		// 新しいフィールド順に沿って、13項目のスライスを生成する
		// 順番は以下の通り：
		// index 0: DatOroshiCode
		// index 1: DatDate
		// index 2: DatDeliveryFlag
		// index 3: DatReceiptNumber
		// index 4: DatJanCode    ← JAN コード（SQL への送信でキーとなる）
		// index 5: DatLineNumber
		// index 6: DatProductName
		// index 7: DatQuantity
		// index 8: DatUnitPrice
		// index 9: DatSubtotal
		// index 10: DatPackagingDrugPrice
		// index 11: DatExpiryDate
		// index 12: DatLotNumber
		recordData := []string{
			currentOroshiCode,
			datDate,
			datDeliveryFlag,
			datReceiptNumber,
			datJanCode,
			datLineNumber,
			datProductName,
			datQuantity,
			datUnitPrice,
			datSubtotal,
			datPackagingDrugPrice,
			datExpiryDate,
			datLotNumber,
		}

		// 次工程（MA0.go 側）へ、このデータを送ります。
		errProc := ProcessDATRecord(recordData)
		if errProc != nil {
			log.Printf("[DAT] レコード処理エラー (JAN=%q): %v", datJanCode, errProc)
		}
		// ※ここで、ProcessDATRecord 内で MA0.go の関数 (CheckOrCreateMA0 等) を呼び出し、
		// MA0 の新規登録か重複かのログ出力を実施しています。
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
