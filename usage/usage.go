// usage/usage.go
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

var taniMap map[string]string

// loadTaniMap は、TANI CSVファイルを読み込み、単位コードから単位名称へのマッピングを作成します。
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

// getOrganizedFlag は、指定されたJANコードについてJCShmsマスターに存在するかをチェックし、
// 存在すれば 1 (organized)、存在しなければ 0 (disorganized) を返します。
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

// ParseUsageFile は、Shift-JISでエンコードされたUSAGE CSVを読み込み、
// 各行を UsageRecord に変換します。各レコードに対して、単位名称の解決および
// JCShmsマスターによる整理状態の判定を実施します。
func ParseUsageFile(r io.Reader) ([]UsageRecord, error) {
	loadTaniMap()
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
			log.Printf("[USAGE] フィールド数不足の行をスキップ: %v", fields)
			continue
		}
		for i, f := range fields {
			fields[i] = strings.Trim(f, "\" ")
		}
		ur := UsageRecord{
			UsageDate:        strings.TrimSpace(fields[0]),
			UsageYjCode:      fields[1],
			UsageJanCode:     fields[2],
			UsageProductName: fields[3],
			UsageAmount:      fields[4],
			UsageUnit:        fields[5],
		}
		if name, ok := taniMap[ur.UsageUnit]; ok {
			ur.UsageUnitName = name
		} else {
			ur.UsageUnitName = ur.UsageUnit
		}
		flag, err := getOrganizedFlag(ur.UsageJanCode)
		if err != nil {
			log.Printf("[USAGE] Organized flag 確認エラー (JAN=%q): %v", ur.UsageJanCode, err)
			ur.OrganizedFlag = 0
		} else {
			ur.OrganizedFlag = flag
		}
		records = append(records, ur)
		// MA0連携処理（必要に応じて）
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
		log.Printf("[USAGE] Parsed record: %+v", ur)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// InsertUsageRecords は、パース済みUsageRecordのスライスをDBの"usagerecords"テーブルに登録します。
func InsertUsageRecords(db *sql.DB, recs []UsageRecord) error {
	stmt := `
		INSERT OR REPLACE INTO usagerecords (
			usageDate,
			usageYjCode,
			usageJanCode,
			usageProductName,
			usageAmount,
			usageUnit,
			usageUnitName,
			organizedFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	for _, r := range recs {
		_, err := db.Exec(stmt,
			strings.TrimSpace(r.UsageDate),
			r.UsageYjCode,
			r.UsageJanCode,
			r.UsageProductName,
			r.UsageAmount,
			r.UsageUnit,
			r.UsageUnitName,
			r.OrganizedFlag,
		)
		if err != nil {
			return fmt.Errorf("failed to insert USAGE record: %w", err)
		}
	}
	return nil
}

// ReplaceUsageRecordsWithPeriod は、ファイル内にある UsageDate の最小値から最大値までの期間のレコードを
// DBから削除した上で、新たなレコードを登録する一連の処理を実施します。
func ReplaceUsageRecordsWithPeriod(db *sql.DB, recs []UsageRecord) error {
	log.Printf("ReplaceUsageRecordsWithPeriod が呼ばれました。レコード件数: %d", len(recs))
	if len(recs) == 0 {
		log.Printf("レコードが存在しないため、処理を終了します。")
		return nil
	}

	// 対象期間を算出
	periodStart := strings.TrimSpace(recs[0].UsageDate)
	periodEnd := strings.TrimSpace(recs[0].UsageDate)
	for _, rec := range recs {
		rdate := strings.TrimSpace(rec.UsageDate)
		if rdate < periodStart {
			periodStart = rdate
		}
		if rdate > periodEnd {
			periodEnd = rdate
		}
	}
	log.Printf("削除対象期間: %s ～ %s", periodStart, periodEnd)

	// 対象期間の既存レコードを削除
	deleteStmt := `DELETE FROM usagerecords WHERE usageDate BETWEEN ? AND ?`
	res, err := db.Exec(deleteStmt, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to delete existing usage records for period %s-%s: %w", periodStart, periodEnd, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		log.Printf("RowsAffectedの取得エラー: %v", err)
	} else {
		log.Printf("対象期間内で削除されたレコード件数: %d", n)
	}

	// 新規レコードの挿入
	if err := InsertUsageRecords(db, recs); err != nil {
		return fmt.Errorf("failed to insert new usage records: %w", err)
	}
	log.Printf("新規レコードの挿入が完了しました。")
	return nil
}
