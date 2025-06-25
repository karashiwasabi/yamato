// File: usage/usage.go
package usage

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"YAMATO/jcshms"
	"YAMATO/ma0"
	"YAMATO/tani"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// UsageRecord はUSAGE CSVの1行分のデータを表します。
type UsageRecord struct {
	UsageDate        string `json:"usageDate"`
	UsageYjCode      string `json:"usageYjCode"`
	UsageJanCode     string `json:"usageJanCode"`
	UsageProductName string `json:"usageProductName"`
	UsageAmount      string `json:"usageAmount"`
	UsageUnit        string `json:"usageUnit"`
	UsageUnitName    string `json:"usageUnitName"`
	OrganizedFlag    int    `json:"organizedFlag"` // 1: organized, 0: disorganized
}

// taniMap はコード→単位名称を保持します（Shift-JIS CSVからロード）。
var taniMap map[string]string

// loadTaniMap は内部用。TANI CSVを読み込み、taniMapを初期化します。
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

	m, err := tani.ParseTANI(f)
	if err != nil {
		log.Printf("TANIパース失敗: %v", err)
		taniMap = make(map[string]string)
		return
	}
	taniMap = m
}

// LoadTaniMap は外部からTANIマップを初期化するための公開関数です。
func LoadTaniMap() {
	loadTaniMap()
}

// GetTaniName はコードをキーに単位名称を返します。
// マップ未初期化時は自動でロードします。
func GetTaniName(code string) string {
	if taniMap == nil {
		loadTaniMap()
	}
	if name, ok := taniMap[code]; ok {
		return name
	}
	return ""
}

// getOrganizedFlag はJCShmsマスターにJANが存在すれば1、なければ0を返します。
func getOrganizedFlag(jan string) (int, error) {
	records, err := jcshms.QueryByJan(ma0.DB, jan)
	if err != nil {
		return 0, fmt.Errorf("jcshms.QueryByJan error: %w", err)
	}
	if len(records) > 0 {
		return 1, nil
	}
	return 0, nil
}

// ParseUsageFile は Shift-JIS の USAGE CSVをパースし、UsageRecordスライスを返します。
// 各レコードで単位名称を解決し、整理フラグをセットします。
func ParseUsageFile(r io.Reader) ([]UsageRecord, error) {
	loadTaniMap() // 念のため
	decoder := japanese.ShiftJIS.NewDecoder()
	scanner := bufio.NewScanner(transform.NewReader(r, decoder))

	var records []UsageRecord
	headerSkipped := false

	for scanner.Scan() {
		line := scanner.Text()
		// ヘッダー行をスキップ
		if !headerSkipped {
			if strings.Contains(line, "UsageDate") {
				headerSkipped = true
				continue
			}
			headerSkipped = true
		}
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			log.Printf("[USAGE] フィールド数不足をスキップ: %v", fields)
			continue
		}
		for i := range fields {
			fields[i] = strings.Trim(fields[i], "\" ")
		}

		ur := UsageRecord{
			UsageDate:        fields[0],
			UsageYjCode:      fields[1],
			UsageJanCode:     fields[2],
			UsageProductName: fields[3],
			UsageAmount:      fields[4],
			UsageUnit:        fields[5],
		}

		// 単位名称解決
		if name := GetTaniName(ur.UsageUnit); name != "" {
			ur.UsageUnitName = name
		} else {
			ur.UsageUnitName = ur.UsageUnit
		}

		// 整理フラグ判定
		flag, err := getOrganizedFlag(ur.UsageJanCode)
		if err != nil {
			log.Printf("[USAGE] Organized flag確認エラー (JAN=%q): %v", ur.UsageJanCode, err)
			ur.OrganizedFlag = 0
		} else {
			ur.OrganizedFlag = flag
		}

		records = append(records, ur)

		// MA0 連携
		dataSlice := []string{
			ur.UsageDate,
			ur.UsageYjCode,
			ur.UsageJanCode,
			ur.UsageProductName,
			ur.UsageAmount,
			ur.UsageUnit,
			ur.UsageUnitName,
		}
		if err := ma0.ProcessMA0Record(dataSlice); err != nil {
			log.Printf("[USAGE] MA0照合エラー (JAN=%q): %v", ur.UsageJanCode, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// ReplaceUsageRecordsWithPeriod は指定期間の既存レコードを削除後、新規挿入します。
func ReplaceUsageRecordsWithPeriod(db *sql.DB, recs []UsageRecord) error {
	if len(recs) == 0 {
		return nil
	}
	start, end := recs[0].UsageDate, recs[0].UsageDate
	for _, r := range recs {
		if r.UsageDate < start {
			start = r.UsageDate
		}
		if r.UsageDate > end {
			end = r.UsageDate
		}
	}

	_, err := db.Exec(`DELETE FROM usagerecords WHERE usageDate BETWEEN ? AND ?`, start, end)
	if err != nil {
		return fmt.Errorf("failed to delete existing usage records: %w", err)
	}

	// 挿入
	stmt := `
INSERT OR REPLACE INTO usagerecords (
  usageDate, usageYjCode, usageJanCode,
  usageProductName, usageAmount, usageUnit, usageUnitName, organizedFlag
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`
	for _, r := range recs {
		if _, err := db.Exec(stmt,
			r.UsageDate, r.UsageYjCode, r.UsageJanCode,
			r.UsageProductName, r.UsageAmount, r.UsageUnit, r.UsageUnitName, r.OrganizedFlag,
		); err != nil {
			return fmt.Errorf("failed to insert USAGE record: %w", err)
		}
	}
	return nil
}
