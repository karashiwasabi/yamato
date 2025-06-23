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

// UsageRecord は USAGE CSV の各レコード情報を保持します
type UsageRecord struct {
	UsageDate        string `json:"usageDate"`
	UsageYjCode      string `json:"usageYjCode"`
	UsageJanCode     string `json:"usageJanCode"`
	UsageProductName string `json:"usageProductName"`
	UsageAmount      string `json:"usageAmount"`
	UsageUnit        string `json:"usageUnit"`
	UsageUnitName    string `json:"usageUnitName"`
}

var taniMap map[string]string

// loadTaniMap は所定のパスの TANI.CSV を読み込み、taniMap にセットします
func loadTaniMap() {
	if taniMap != nil {
		return
	}
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

// ParseUsageFile は、USAGE CSV を Shift‑JIS から UTF‑8 に変換しながらパースして UsageRecord のスライスを返します。
// 各レコードの全データを抽出し、同時に MA0 への連携（全データをそのまま送る）を行います。
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

		// カンマ区切りで各フィールドを取得
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}
		// 各フィールドの前後の引用符や空白を除去
		for i, f := range fields {
			fields[i] = strings.Trim(f, "\" ")
		}

		ur := UsageRecord{
			UsageDate:        fields[0],
			UsageYjCode:      fields[1],
			UsageJanCode:     fields[2],
			UsageProductName: fields[3],
			UsageAmount:      fields[4],
			UsageUnit:        fields[5],
		}
		// TANI マスターより単位コードに対応する単位名称を取得
		if name, ok := taniMap[ur.UsageUnit]; ok {
			ur.UsageUnitName = name
		} else {
			ur.UsageUnitName = ur.UsageUnit
		}

		records = append(records, ur)

		// ここで UsageRecord の全データをスライスにまとめ、ma0 へそのまま送ります
		recordData := []string{
			ur.UsageDate,        // 0: 使用日
			ur.UsageYjCode,      // 1: YJコード
			ur.UsageJanCode,     // 2: JANコード ← MA0 のキーとなります
			ur.UsageProductName, // 3: 商品名
			ur.UsageAmount,      // 4: 数量／金額
			ur.UsageUnit,        // 5: 単位コード
			ur.UsageUnitName,    // 6: 単位名称
		}
		if err := ma0.ProcessMA0Record(recordData); err != nil {
			log.Printf("[USAGE] MA0照合エラー (JAN=%q): %v", ur.UsageJanCode, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
