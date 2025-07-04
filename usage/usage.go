// File: YAMATO/usage/usage.go
package usage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"YAMATO/jcshms"
	"YAMATO/ma0"
	"YAMATO/tani"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// UsageRecord は USAGE CSV の１行分を表します。
type UsageRecord struct {
	UsageDate        string `json:"usageDate"`
	UsageYjCode      string `json:"usageYjCode"`
	UsageJanCode     string `json:"usageJanCode"`
	UsageProductName string `json:"usageProductName"`
	UsageAmount      string `json:"usageAmount"`
	UsageUnit        string `json:"usageUnit"`
	UsageUnitName    string `json:"usageUnitName"`
	OrganizedFlag    int    `json:"organizedFlag"`
}

var taniMap map[string]string

// loadTaniMap は内部用：TANI.CSV を読み込んで taniMap を初期化します。
func loadTaniMap() {
	if taniMap != nil {
		return
	}
	f, err := os.Open("C:\\Dev\\YAMATO\\SOU\\TANI.CSV")
	if err != nil {
		log.Printf("TANI file open error: %v", err)
		taniMap = make(map[string]string)
		return
	}
	defer f.Close()

	m, err := tani.ParseTANI(f)
	if err != nil {
		log.Printf("TANI parse error: %v", err)
		taniMap = make(map[string]string)
		return
	}
	taniMap = m
}

// GetTaniName は単位コードから単位名称を返します。
func GetTaniName(code string) string {
	if taniMap == nil {
		loadTaniMap()
	}
	if name, ok := taniMap[code]; ok {
		return name
	}
	return ""
}

// getOrganizedFlag は JCShms マスターに JAN があれば1、なければ0を返します。
func getOrganizedFlag(jan string) int {
	recs, err := jcshms.QueryByJan(ma0.DB, jan)
	if err != nil {
		log.Printf("[USAGE] OrganizedFlag error JAN=%q: %v", jan, err)
		return 0
	}
	if len(recs) > 0 {
		return 1
	}
	return 0
}

// ParseUsageFile は SHIFT-JIS USAGE CSV を読み込み、UsageRecord スライスを返します。
// MA0 未登録品は MA2 テーブルに登録します。
func ParseUsageFile(r io.Reader) ([]UsageRecord, error) {
	loadTaniMap()
	scanner := bufio.NewScanner(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))

	var records []UsageRecord
	headerSkipped := false

	for scanner.Scan() {
		line := scanner.Text()
		if !headerSkipped {
			headerSkipped = true
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			log.Printf("[USAGE] skip short row: %v", fields)
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
		// 単位名称を解決
		if nm := GetTaniName(ur.UsageUnit); nm != "" {
			ur.UsageUnitName = nm
		} else {
			ur.UsageUnitName = ur.UsageUnit
		}
		// organizedFlag
		ur.OrganizedFlag = getOrganizedFlag(ur.UsageJanCode)

		// MA0 連携／MA2 登録
		ma0Rec, created, err0 := ma0.CheckOrCreateMA0(ur.UsageJanCode, ur.UsageProductName)
		if err0 != nil {
			log.Printf("[USAGE] MA0 lookup error JAN=%s: %v", ur.UsageJanCode, err0)
		}

		// マスター名未設定の既存品のみ MA2 登録
		if !created && ma0Rec.MA018JC018ShouhinMei == "" {
			hs, _ := strconv.Atoi(ma0Rec.MA044JC044HousouSouryouSuuchi)
			jsn, _ := strconv.Atoi(ma0Rec.MA131JA006HousouSuuryouSuuchi)
			jssn, _ := strconv.Atoi(ma0Rec.MA133JA008HousouSouryouSuuchi)

			mrec := &ma0.MARecord{
				JanCode:                ur.UsageJanCode,
				ProductName:            ur.UsageProductName,
				HousouKeitai:           ma0Rec.MA037JC037HousouKeitai,
				HousouTaniUnit:         ur.UsageUnit,
				HousouSouryouNumber:    hs,
				JanHousouSuuryouNumber: jsn,
				JanHousouSuuryouUnit:   ma0Rec.MA132JA007HousouSuuryouTaniCode,
				JanHousouSouryouNumber: jssn,
			}
			// シーケンスを受け取りつつ登録（戻り値は破棄）
			_, _, err2 := ma0.RegisterMA(ma0.DB, mrec)
			if err2 != nil {
				log.Printf("[USAGE] MA2 registration error JAN=%s: %v", ur.UsageJanCode, err2)
			}
		}

		records = append(records, ur)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("USAGE scan error: %w", err)
	}
	return records, nil
}

// LoadTaniMap は main.go から呼ばれる公開版です。
func LoadTaniMap() {
	loadTaniMap()
}

// ReplaceUsageRecordsWithPeriod は main.go から呼ばれる公開版です。
// 指定期間の USAGE レコードを削除し、再挿入します。
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
	if _, err := db.Exec(
		`DELETE FROM usagerecords WHERE usageDate BETWEEN ? AND ?`,
		start, end,
	); err != nil {
		return fmt.Errorf("delete existing USAGE error: %w", err)
	}
	stmt := `
        INSERT OR REPLACE INTO usagerecords (
          usageDate, usageYjCode, usageJanCode,
          usageProductName, usageAmount, usageUnit,
          usageUnitName, organizedFlag
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	for _, r := range recs {
		if _, err := db.Exec(stmt,
			r.UsageDate, r.UsageYjCode, r.UsageJanCode,
			r.UsageProductName, r.UsageAmount, r.UsageUnit,
			r.UsageUnitName, r.OrganizedFlag,
		); err != nil {
			return fmt.Errorf("insert USAGE record error: %w", err)
		}
	}
	return nil
}

// GetTaniMap は main.go から呼ばれる公開版です。
func GetTaniMap() map[string]string {
	loadTaniMap()
	return taniMap
}

// UploadUsageHandler は USAGE CSV アップロードを処理します。
func UploadUsageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[UploadUsageHandler] start")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["usageFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No USAGE file uploaded", http.StatusBadRequest)
		return
	}

	var allRecords []UsageRecord
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Printf("[UploadUsageHandler] open error: %v", err)
			continue
		}
		recs, err := ParseUsageFile(file)
		file.Close()
		if err != nil {
			log.Printf("[UploadUsageHandler] parse error: %v", err)
			continue
		}
		allRecords = append(allRecords, recs...)
	}

	if err := ReplaceUsageRecordsWithPeriod(ma0.DB, allRecords); err != nil {
		log.Printf("[UploadUsageHandler] replace error: %v", err)
		http.Error(w, "Failed to update USAGE records", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"TotalRecords": len(allRecords),
		"USAGERecords": allRecords,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}
