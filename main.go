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

	// 自動ブラウザ起動
	go autoLaunchBrowser("http://localhost:8080")

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
