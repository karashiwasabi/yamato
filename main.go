// File: main.go
package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"YAMATO/aggregate"
	"YAMATO/dat"
	"YAMATO/inout"
	"YAMATO/inventory"
	"YAMATO/ma0"
	"YAMATO/ma2"
	"YAMATO/model"
	"YAMATO/usage"

	_ "github.com/mattn/go-sqlite3"
)

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
	exec.Command(cmd, args...).Start()
}

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

func main() {
	// Register MIME types for static files
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")

	// Open SQLite database
	db, err := sql.Open("sqlite3", "yamato.db")
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	// Provide DB to other packages
	ma0.DB = db
	inout.DB = db
	aggregate.SetDB(db)
	usage.LoadTaniMap()

	// Apply schema.sql
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("read schema.sql error: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatalf("exec schema.sql error: %v", err)
	}

	// Load master CSVs
	if err := loadCSV(db, "SOU/JCSHMS.CSV", "jcshms", 125, false); err != nil {
		log.Fatalf("load JCSHMS failed: %v", err)
	}
	if err := loadCSV(db, "SOU/JANCODE.CSV", "jancode", 30, true); err != nil {
		log.Fatalf("load JANCODE failed: %v", err)
	}

	// Static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/api/productName", productNameHandler)

	// API endpoints
	http.HandleFunc("/uploadDat", uploadDatHandler)
	http.HandleFunc("/uploadUsage", usage.UploadUsageHandler)

	http.HandleFunc("/uploadInventory", inventory.UploadInventoryHandler)
	http.HandleFunc("/aggregate", aggregate.AggregateHandler)

	// Inout (出庫・入庫)
	http.HandleFunc("/api/inout", inout.Handler)
	http.HandleFunc("/api/inout/search", inout.ProductSearchHandler)
	http.HandleFunc("/api/inout/save", inout.SaveIODHandler)

	// MA2 endpoints
	http.HandleFunc("/api/ma2", listMa2Handler)
	http.HandleFunc("/api/ma2/upsert", ma2.UpsertHandler)

	// TANI map endpoint
	http.HandleFunc("/api/tani", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(usage.GetTaniMap())
	})

	// Auto-open browser
	go autoLaunchBrowser("http://localhost:8080")

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
