package usage

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"YAMATO/ma0"
	"YAMATO/tani"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// UsageRecord は USAGE CSV の各レコード情報を保持します。
type UsageRecord struct {
	UsageDate        string `json:"usageDate"`        // 使用日
	UsageYjCode      string `json:"usageYjCode"`      // YJコード（以前の UsageCode）
	UsageJanCode     string `json:"usageJanCode"`     // JANコード（以前の JANCode）
	UsageProductName string `json:"usageProductName"` // 商品名（以前の ProductName）
	UsageAmount      string `json:"usageAmount"`      // 数量／金額（以前の QuantityOrAmount）
	UsageUnit        string `json:"usageUnit"`        // 単位コード（以前の Unit）
	UsageUnitName    string `json:"usageUnitName"`    // 単位名称（以前の UnitName → TANI マスターから取得）
}

// taniMap は TANI マスターのデータを保持するグローバル変数です。
var taniMap map[string]string

// loadTaniMap は、所定のパスから TANI.CSV を読み込み、 taniMap にセットします。
func loadTaniMap() {
	if taniMap != nil {
		return
	}
	// TANI.CSV のパス。各環境に合わせて修正してください。
	f, err := os.Open("C:\\Dev\\YAMATO\\SOU\\TANI.CSV")
	if err != nil {
		log.Printf("TANIファイルオープンエラー: %v", err)
		taniMap = make(map[string]string)
		return
	}
	defer f.Close()
	tMap, err := tani.ParseTANI(f)
	if err != nil {
		log.Printf("TANIパース失敗: %v", err)
		taniMap = make(map[string]string)
		return
	}
	taniMap = tMap
}

// ParseUsageFile は、USAGE CSV をパースして動作情報を UsageRecord として返します。
// USAGE CSV は Shift‑JIS でエンコードされているため変換を適用し、最初のヘッダー行をスキップして各行を処理します。
func ParseUsageFile(r io.Reader) ([]UsageRecord, error) {
	loadTaniMap()

	var records []UsageRecord
	scanner := bufio.NewScanner(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	headerSkipped := false
	for scanner.Scan() {
		line := scanner.Text()
		// ヘッダー行（"UsageDate" を含む場合）をスキップ
		if !headerSkipped {
			if strings.Contains(line, "UsageDate") {
				headerSkipped = true
				continue
			}
			headerSkipped = true
		}
		// カンマ区切りで分割
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}
		// 各フィールドの前後の引用符や空白を除去
		for i, f := range fields {
			fields[i] = strings.Trim(f, "\" ")
		}
		// 仮のフィールド配置（実際は仕様に合わせて調整してください）：
		// fields[0]: 使用日, [1]: YJコード, [2]: JANコード, [3]: 商品名, [4]: 数量／金額, [5]: 単位コード
		ur := UsageRecord{
			UsageDate:        fields[0],
			UsageYjCode:      fields[1],
			UsageJanCode:     fields[2],
			UsageProductName: fields[3],
			UsageAmount:      fields[4],
			UsageUnit:        fields[5],
		}
		// TANI マスターから、単位コードに対応する単位名称を取得
		if name, ok := taniMap[ur.UsageUnit]; ok {
			ur.UsageUnitName = name
		} else {
			ur.UsageUnitName = ur.UsageUnit
		}
		records = append(records, ur)

		// MA0 照合処理：UsageRecord の JANコード（UsageJanCode）をキーとしてチェック
		_, created, err := ma0.CheckOrCreateMA0(ur.UsageJanCode)
		if err != nil {
			log.Printf("[USAGE] MA0照合エラー (JAN=%q): %v", ur.UsageJanCode, err)
		} else if created {
			log.Printf("[USAGE] 新規MA0登録: %q", ur.UsageJanCode)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
