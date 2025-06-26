package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"YAMATO/ma0"
	"YAMATO/usage"
)

// InventoryRecord は棚卸１件分の情報です。
type InventoryRecord struct {
	Date    string `json:"Date"`    // YYYYMMDD
	JAN     string `json:"JAN"`     // コロン等除去済JANコード
	MA0Name string `json:"MA0Name"` // MA0 から取得した商品名
	CSVName string `json:"CSVName"` // CSV 上の生商品名(rec[12])
	Qty     int    `json:"Qty"`     // rawQty × MA131 換算後在庫数
	Unit    string `json:"Unit"`    // MA038 を名称に変換した単位
}

// sanitizeDigits は文字列から数字以外をすべて削除します。
func sanitizeDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// ParseInventoryCSV は Shift-JIS 形式の棚卸 CSV を読み込み、
// ・日付/JAN を sanitizeDigits で整形
// ・R1 行は問答無用に読み込み
// ・rec[12] を生 CSV 名として保持
// ・Rec[21] を ParseFloat→int で rawQty として扱う
// ・ma0.CheckOrCreateMA0 で JCSHMS/JANCODE 参照＆ma0 連携
// ・MA131 で在庫換算、MA038 で基本単位取得
// ・usage.GetTaniName で単位名称解決
// を行い、[]InventoryRecord を返します。
func ParseInventoryCSV(r io.Reader) ([]InventoryRecord, error) {
	rd := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	rd.FieldsPerRecord = -1

	rows, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV読み込みエラー: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("CSVにデータがありません")
	}

	// ヘッダー1行目の5列目(インデックス4)から日付取得、数字以外除去
	dateRaw := rows[0][4]
	date := sanitizeDigits(dateRaw)

	var out []InventoryRecord
	for i, rec := range rows[1:] {
		rowNum := i + 2 // ヘッダー行を含めた実際の行番号

		// R1行以外は列数チェック
		if rec[0] != "R1" && len(rec) <= 45 {
			log.Printf("行%dスキップ: 列数不足 (len=%d)", rowNum, len(rec))
			continue
		}

		// 生在庫数(21列目)を ParseFloat→int
		rawQ := strings.TrimSpace(rec[21])
		rawF, err := strconv.ParseFloat(rawQ, 64)
		if err != nil {
			if rec[0] == "R1" {
				log.Printf("行%d: R1行 rawQtyパース失敗 %q → 0 とみなす", rowNum, rawQ)
				rawF = 0
			} else {
				log.Printf("行%dスキップ: rawQtyパース失敗 %q", rowNum, rawQ)
				continue
			}
		}
		rawQty := int(rawF)

		// CSV上の商品名(rec[12])
		csvName := strings.TrimSpace(rec[12])

		// JAN(45列目)→数字のみ抽出
		janRaw := rec[45]
		jan := sanitizeDigits(janRaw)

		// MA0 連携(JCSHMS/JANCODE 参照＆ma0 テーブル埋め)
		maRec, _, err := ma0.CheckOrCreateMA0(jan)
		if err != nil {
			log.Printf("行%d: MA0連携エラー JAN=%s: %v", rowNum, jan, err)
		}

		// 包装基準数量 MA131 → pkgQty
		pkgQty, err := strconv.Atoi(maRec.MA131JA006HousouSuuryouSuuchi)
		if err != nil || pkgQty <= 0 {
			pkgQty = 1
		}

		// 換算後在庫数
		qty := rawQty * pkgQty

		// 基本単位コード MA038 → 単位名称
		basicUnitCode := maRec.MA038JC038HousouTaniSuuchi
		unitName := usage.GetTaniName(basicUnitCode)
		if unitName == "" {
			unitName = basicUnitCode
		}

		// MA0からの商品名
		ma0Name := maRec.MA018JC018ShouhinMei

		out = append(out, InventoryRecord{
			Date:    date,
			JAN:     jan,
			MA0Name: ma0Name,
			CSVName: csvName,
			Qty:     qty,
			Unit:    unitName,
		})
	}

	return out, nil
}
