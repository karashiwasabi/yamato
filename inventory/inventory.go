package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"YAMATO/ma0"
	"YAMATO/usage"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// InventoryRecord は棚卸１件分の情報です。
type InventoryRecord struct {
	Date    string `json:"Date"`    // YYYYMMDD
	JAN     string `json:"JAN"`     // コロン等除去済JANコード
	MA0Name string `json:"MA0Name"` // MA0 から取得した商品名
	CSVName string `json:"CSVName"` // CSV 上の生商品名(rec[12])
	Qty     int    `json:"Qty"`     // rawQty × MA131 換算後在庫数
	Unit    string `json:"Unit"`    // MA038 を名称に変換した単位

	// 以下、CSVから取得した包装関連
	PackagingUnit    string  `json:"packagingUnit"`    // CSV[16] 包装単位名称
	JanPackagingQty  float64 `json:"janPackagingQty"`  // CSV[17] JAN包装数量
	JanPackagingUnit string  `json:"janPackagingUnit"` // CSV[23] JAN包装単位名称
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
// InventoryRecord スライスを返します。
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

	// ヘッダー１行目(インデックス0)の5列目(インデックス4)から日付取得
	dateRaw := rows[0][4]
	date := sanitizeDigits(dateRaw)

	var out []InventoryRecord
	for i, rec := range rows[1:] {
		rowNum := i + 2
		if rec[0] != "R1" && len(rec) <= 45 {
			log.Printf("行%dスキップ: 列数不足(len=%d)", rowNum, len(rec))
			continue
		}

		// rawQty (21列目) を ParseFloat → int
		rawF, err := strconv.ParseFloat(strings.TrimSpace(rec[21]), 64)
		if err != nil {
			if rec[0] == "R1" {
				log.Printf("行%d: R1行 rawQtyパース失敗 %q → 0とみなす", rowNum, rec[21])
				rawF = 0
			} else {
				log.Printf("行%dスキップ: rawQtyパース失敗 %q", rowNum, rec[21])
				continue
			}
		}
		rawQty := int(rawF)

		// CSV上の商品名(rec[12])
		csvName := strings.TrimSpace(rec[12])

		// JANコード(45列目)→数字のみ抽出
		jan := sanitizeDigits(rec[45])

		// MA0 連携して在庫換算と基本単位取得
		maRec, _, err := ma0.CheckOrCreateMA0(jan)
		if err != nil {
			log.Printf("行%d: MA0連携エラー JAN=%s: %v", rowNum, jan, err)
		}
		pkgQty, err := strconv.Atoi(maRec.MA131JA006HousouSuuryouSuuchi)
		if err != nil || pkgQty <= 0 {
			pkgQty = 1
		}
		qty := rawQty * pkgQty

		basicUnitCode := maRec.MA038JC038HousouTaniSuuchi
		unitName := usage.GetTaniName(basicUnitCode)
		if unitName == "" {
			unitName = basicUnitCode
		}

		// MA0 から取得した商品名
		ma0Name := maRec.MA018JC018ShouhinMei

		// CSV の 16,17,23 列から包装情報を取得
		pu, jpUnit := "", ""
		jpQty := 0.0
		if len(rec) > 23 {
			pu = strings.Trim(rec[16], "：:'\"")
			jpQty, _ = strconv.ParseFloat(strings.TrimSpace(rec[17]), 64)
			jpUnit = strings.Trim(rec[23], "：:'\"")
		}

		out = append(out, InventoryRecord{
			Date:             date,
			JAN:              jan,
			MA0Name:          ma0Name,
			CSVName:          csvName,
			Qty:              qty,
			Unit:             unitName,
			PackagingUnit:    pu,
			JanPackagingQty:  jpQty,
			JanPackagingUnit: jpUnit,
		})
	}
	return out, nil
}
