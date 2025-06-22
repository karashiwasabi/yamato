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

	_ "github.com/mattn/go-sqlite3"

	"YAMATO/dat"
	"YAMATO/ma0"
	"YAMATO/usage"
)

var db *sql.DB

// uploadDatHandler は既存のまま
func uploadDatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["datFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No DAT file uploaded", http.StatusBadRequest)
		return
	}

	var allRecords []dat.DATRecord
	totalCount, ma0CreatedCount, duplicateCount := 0, 0, 0
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Error opening DAT file:", err)
			continue
		}
		defer file.Close()

		records, tc, mc, dc, err := dat.ParseDATFile(file)
		if err != nil {
			log.Println("Error parsing DAT file:", err)
			continue
		}
		totalCount += tc
		ma0CreatedCount += mc
		duplicateCount += dc
		allRecords = append(allRecords, records...)
	}

	resp := map[string]interface{}{
		"DATReadCount":    totalCount,
		"MA0CreatedCount": ma0CreatedCount,
		"DuplicateCount":  duplicateCount,
		"DATRecords":      allRecords,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// uploadUsageHandler は既存のまま
func uploadUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["usageFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No USAGE file uploaded", http.StatusBadRequest)
		return
	}
	var allRecords []usage.UsageRecord
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Error opening USAGE file:", err)
			continue
		}
		defer file.Close()
		records, err := usage.ParseUsageFile(file)
		if err != nil {
			log.Println("Error parsing USAGE file:", err)
			continue
		}
		allRecords = append(allRecords, records...)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"USAGERecords": allRecords,
		"TotalRecords": len(allRecords),
	})
}

// viewMA0Handler は既存のまま
func viewMA0Handler(w http.ResponseWriter, r *http.Request) {
	ma0.ViewMA0Handler(w, r)
}

// autoLaunchBrowser は既存のまま
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
		log.Printf("Browser auto-launch failed: %v", err)
	}
}

// loadCSV は起動時の JCHMAS/JANCODE 一括取り込み用
func loadCSV(db *sql.DB, filePath, table string, cols int) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	// ヘッダー行をスキップ
	if _, err := r.Read(); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// プレースホルダーを作成
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
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		args := make([]interface{}, len(rec))
		for i, v := range rec {
			args[i] = v
		}
		if _, err := stmt.Exec(args...); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func main() {
	var err error

	// ① DB を開く
	db, err = sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	// ② schema.sql を読み込んでテーブル定義を反映（IF NOT EXISTS を schema.sql に記述）
	b, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}
	if _, err := db.Exec(string(b)); err != nil {
		log.Fatalf("Failed to exec schema.sql: %v", err)
	}

	// ③ SOUフォルダの JCHMAS/JANCODE CSV を毎回取り込み
	jchmasPath := `C:\Dev\YAMATO\SOU\JCSHMS.CSV`
	jancodePath := `C:\Dev\YAMATO\SOU\JANCODE.CSV`
	if err := loadCSV(db, jchmasPath, "jchmas", 125); err != nil {
		log.Fatalf("failed to load %s: %v", jchmasPath, err)
	}
	if err := loadCSV(db, jancodePath, "jancode", 30); err != nil {
		log.Fatalf("failed to load %s: %v", jancodePath, err)
	}

	// ④ HTTP サーバーの既存処理
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/uploadDat", uploadDatHandler)
	http.HandleFunc("/uploadUsage", uploadUsageHandler)
	http.HandleFunc("/viewMA0", viewMA0Handler)

	go autoLaunchBrowser("http://localhost:8080")
	log.Println("Server starting on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
