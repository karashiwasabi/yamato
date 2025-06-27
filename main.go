// File: main.go
package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"YAMATO/aggregate"
	"YAMATO/dat"
	"YAMATO/inventory"
	"YAMATO/ma0"
	"YAMATO/model"
	"YAMATO/usage"

	_ "github.com/mattn/go-sqlite3"
)

// loadCSV は Shift-JIS エンコードの CSV を指定テーブルにロードするユーティリティ
func loadCSV(db *sql.DB, filePath, table string, cols int, skipHeader bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rd := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
	rd.LazyQuotes = true
	rd.FieldsPerRecord = -1
	if skipHeader {
		if _, err := rd.Read(); err != nil {
			return err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	ph := make([]string, cols)
	for i := range ph {
		ph[i] = "?"
	}
	stmt, err := tx.Prepare(
		"INSERT OR REPLACE INTO " + table + " VALUES(" + strings.Join(ph, ",") + ")",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		rec, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		args := make([]interface{}, len(rec))
		for i, v := range rec {
			args[i] = v
		}
		if _, err := stmt.Exec(args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// uploadDatHandler は DAT ファイルのアップロードを処理
func uploadDatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["datFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No DAT file uploaded", http.StatusBadRequest)
		return
	}

	var all []model.DATRecord
	var total, created, dup int
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Println("open DAT error:", err)
			continue
		}
		recs, tc, mc, dc, err := dat.ParseDATFile(file)
		file.Close()
		if err != nil {
			log.Println("parse DAT error:", err)
			continue
		}
		total += tc
		created += mc
		dup += dc
		all = append(all, recs...)
	}

	resp := map[string]interface{}{
		"DATReadCount":    total,
		"MA0CreatedCount": created,
		"DuplicateCount":  dup,
		"DATRecords":      all,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// uploadUsageHandler は USAGE ファイルのアップロードを処理
func uploadUsageHandler(w http.ResponseWriter, r *http.Request) {
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

	var allRecords []usage.UsageRecord
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Printf("Error opening usage file %s: %v", fh.Filename, err)
			continue
		}
		recs, err := usage.ParseUsageFile(file)
		file.Close()
		if err != nil {
			log.Printf("Error parsing usage file %s: %v", fh.Filename, err)
			continue
		}
		allRecords = append(allRecords, recs...)
	}

	if err := usage.ReplaceUsageRecordsWithPeriod(ma0.DB, allRecords); err != nil {
		log.Printf("Failed to replace USAGE records: %v", err)
		http.Error(w, "Failed to update USAGE records", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"TotalRecords": len(allRecords),
		"USAGERecords": allRecords,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// autoLaunchBrowser は起動時にデフォルトブラウザでアプリを開く
func autoLaunchBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("browser start failed: %v", err)
	}
}

func uploadInventoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("inventoryFile")
	if err != nil {
		http.Error(w, "ファイルが指定されていません", http.StatusBadRequest)
		return
	}
	defer file.Close()

	recs, err := inventory.ParseInventoryCSV(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// DB に UPSERT するときの product_name を決める
	for _, rec := range recs {
		// MA0 名があればそれを、なければ CSV 名を使う
		prodName := rec.MA0Name
		if prodName == "" {
			prodName = rec.CSVName
		}

		_, err := ma0.DB.Exec(
			`INSERT OR REPLACE INTO inventory
               (inv_date, jan_code, product_name, qty, unit)
             VALUES (?, ?, ?, ?, ?)`,
			rec.Date, rec.JAN, prodName, rec.Qty, rec.Unit,
		)
		if err != nil {
			log.Printf("inventory upsert error: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":       len(recs),
		"inventories": recs, // rec.MA0Name／rec.CSVName が JSON に含まれます
	})
}

// listMa2Handler は MA2全件を空配列保証で返却
func listMa2Handler(w http.ResponseWriter, r *http.Request) {
	type rec struct {
		JanCode                string `json:"janCode"`
		YjCode                 string `json:"yjCode"`
		Shouhinmei             string `json:"shouhinmei"`
		HousouKeitai           string `json:"housouKeitai"`
		HousouTaniUnit         string `json:"housouTaniUnit"`
		HousouSouryouNumber    int    `json:"housouSouryouNumber"`
		JanHousouSuuryouNumber int    `json:"janHousouSuuryouNumber"`
		JanHousouSuuryouUnit   string `json:"janHousouSuuryouUnit"`
		JanHousouSouryouNumber int    `json:"janHousouSouryouNumber"`
	}
	out := make([]rec, 0) // nil→[] になる
	rows, err := ma0.DB.Query(`
      SELECT MA2JanCode, MA2YjCode, Shouhinmei,
             HousouKeitai, HousouTaniUnit, HousouSouryouNumber,
             JanHousouSuuryouNumber, JanHousouSuuryouUnit, JanHousouSouryouNumber
        FROM ma2
       ORDER BY MA2JanCode
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r rec
		if err := rows.Scan(
			&r.JanCode, &r.YjCode, &r.Shouhinmei,
			&r.HousouKeitai, &r.HousouTaniUnit, &r.HousouSouryouNumber,
			&r.JanHousouSuuryouNumber, &r.JanHousouSuuryouUnit, &r.JanHousouSouryouNumber,
		); err != nil {
			continue
		}
		out = append(out, r)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(out)
}

// upsertMa2Handler は INSERT OR REPLACE で upsert
func upsertMa2Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec struct {
		JanCode                string `json:"janCode"`
		YjCode                 string `json:"yjCode"`
		Shouhinmei             string `json:"shouhinmei"`
		HousouKeitai           string `json:"housouKeitai"`
		HousouTaniUnit         string `json:"housouTaniUnit"`
		HousouSouryouNumber    int    `json:"housouSouryouNumber"`
		JanHousouSuuryouNumber int    `json:"janHousouSuuryouNumber"`
		JanHousouSuuryouUnit   string `json:"janHousouSuuryouUnit"`
		JanHousouSouryouNumber int    `json:"janHousouSouryouNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err := ma0.DB.Exec(`
      INSERT OR REPLACE INTO ma2
       (MA2JanCode, MA2YjCode, Shouhinmei,
        HousouKeitai, HousouTaniUnit, HousouSouryouNumber,
        JanHousouSuuryouNumber, JanHousouSuuryouUnit, JanHousouSouryouNumber)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		rec.JanCode, rec.YjCode, rec.Shouhinmei,
		rec.HousouKeitai, rec.HousouTaniUnit, rec.HousouSouryouNumber,
		rec.JanHousouSuuryouNumber, rec.JanHousouSuuryouUnit, rec.JanHousouSouryouNumber,
	)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// SQLite DB を開く
	db, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	// global DB をセット
	ma0.DB = db

	// TANI マップを先にロードしておく
	usage.LoadTaniMap()

	aggregate.SetDB(db)

	// スキーマ読み込み
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("read schema.sql error: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatalf("exec schema.sql error: %v", err)
	}

	// マスター CSV をロード
	if err := loadCSV(db, "SOU/JCSHMS.CSV", "jcshms", 125, false); err != nil {
		log.Fatalf("load JCSHMS failed: %v", err)
	}
	if err := loadCSV(db, "SOU/JANCODE.CSV", "jancode", 30, true); err != nil {
		log.Fatalf("load JANCODE failed: %v", err)
	}

	// 静的ファイル配信
	staticFS := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFS))
	http.Handle("/", staticFS)

	// API エンドポイント
	http.HandleFunc("/uploadDat", uploadDatHandler)
	http.HandleFunc("/uploadUsage", uploadUsageHandler)
	http.HandleFunc("/aggregate", aggregate.AggregateHandler)
	http.HandleFunc("/productName", productNameHandler)
	http.HandleFunc("/uploadInventory", uploadInventoryHandler)

	// --- MA2編集用API 追加 ---
	http.HandleFunc("/api/ma2", listMa2Handler)
	http.HandleFunc("/api/ma2/upsert", upsertMa2Handler)
	http.HandleFunc("/api/tani", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(usage.GetTaniMap())
	})

	// 自動ブラウザ起動
	go autoLaunchBrowser("http://localhost:8080")

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
