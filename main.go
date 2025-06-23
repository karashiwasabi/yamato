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

	"YAMATO/dat"
	"YAMATO/ma0"
	"YAMATO/usage"

	_ "github.com/mattn/go-sqlite3"
)

// loadCSV は Shift-JIS → UTF-8 変換しつつ CSV を INSERT OR REPLACE します。
// table: テーブル名, cols: カラム数, skipHeader: ヘッダ行をスキップするか
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

	// VALUES(?,?,…)
	ph := make([]string, cols)
	for i := range ph {
		ph[i] = "?"
	}

	stmt, err := tx.Prepare(
		"INSERT OR REPLACE INTO " + table +
			" VALUES(" + strings.Join(ph, ",") + ")",
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

// uploadDatHandler は /uploadDat エンドポイント
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

	var all []dat.DATRecord
	total, created, dup := 0, 0, 0

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Println("open DAT error:", err)
			continue
		}
		defer file.Close()

		recs, tc, mc, dc, err := dat.ParseDATFile(file)
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

// uploadUsageHandler は /uploadUsage エンドポイント
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

	var all []usage.UsageRecord
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			log.Println("open USAGE error:", err)
			continue
		}
		defer file.Close()

		recs, err := usage.ParseUsageFile(file)
		if err != nil {
			log.Println("parse USAGE error:", err)
			continue
		}
		all = append(all, recs...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"USAGERecords": all,
		"TotalRecords": len(all),
	})
}

// autoLaunchBrowser は起動時にブラウザを自動オープン
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
	// SQLite DB オープン
	db, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	// ma0 パッケージに DB をセット（DAT/Usage 連携用）
	ma0.DB = db

	// スキーマ実行
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("read schema.sql error: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatalf("exec schema.sql error: %v", err)
	}

	// マスター CSV のロード
	if err := loadCSV(db, "SOU/JCSHMS.CSV", "jcshms", 125, false); err != nil {
		log.Fatalf("load JCSHMS failed: %v", err)
	}
	if err := loadCSV(db, "SOU/JANCODE.CSV", "jancode", 30, true); err != nil {
		log.Fatalf("load JANCODE failed: %v", err)
	}

	// ルーティング設定
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/uploadDat", uploadDatHandler)
	http.HandleFunc("/uploadUsage", uploadUsageHandler)

	// 自動ブラウザ起動
	go autoLaunchBrowser("http://localhost:8080")

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
