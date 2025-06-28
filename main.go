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
	"YAMATO/inventory"
	"YAMATO/jcshms"
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

// uploadInventoryHandler は棚卸 CSV を受け取り、inventory テーブルに UPSERT
// ———— JCSHMS に未登録の JAN だけ MA2 登録 ————
func uploadInventoryHandler(w http.ResponseWriter, r *http.Request) {
	// 1) POST チェック
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2) multipart/form-data からファイル取得
	file, _, err := r.FormFile("inventoryFile")
	if err != nil {
		http.Error(w, "ファイルが指定されていません", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3) CSV → InventoryRecord スライス
	recs, err := inventory.ParseInventoryCSV(file)
	if err != nil {
		http.Error(w, "CSV読み込みエラー: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 4) 各レコード処理
	for i := range recs {
		rec := &recs[i]

		// (A) MA0 に登録 or 取得
		maRec, _, err := ma0.CheckOrCreateMA0(rec.InvJanCode)
		if err != nil {
			log.Printf("[INVENTORY] MA0 lookup/create error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}

		// (B) JCSHMS に存在するかチェック
		cs, err := jcshms.QueryByJan(ma0.DB, rec.InvJanCode)
		if err != nil {
			log.Printf("[INVENTORY] JCShms lookup error JAN=%s: %v", rec.InvJanCode, err)
		}

		if len(cs) == 0 {
			// → 未登録 JAN のみ MA2 登録
			_, yjSeq, err := ma0.RegisterMA(ma0.DB, &ma0.MARecord{
				JanCode:                rec.InvJanCode,
				ProductName:            rec.InvProductName,
				HousouKeitai:           "",
				HousouTaniUnit:         rec.InvHousouTaniUnit,
				HousouSouryouNumber:    0,
				JanHousouSuuryouNumber: int(rec.InvJanHousouSuuryouNumber),
				JanHousouSuuryouUnit:   rec.JanHousouSuuryouUnit,
				JanHousouSouryouNumber: 0,
			})
			if err != nil {
				log.Printf("[INVENTORY] MA2 registration error JAN=%s: %v", rec.InvJanCode, err)
			} else {
				rec.InvYjCode = yjSeq
			}
		} else {
			// → 既存 JAN は ma0 から返ってきた YJ を使う
			rec.InvYjCode = maRec.MA009JC009YJCode
		}

		// (C) inventory テーブルに UPSERT
		prod := maRec.MA018JC018ShouhinMei
		if prod == "" {
			prod = rec.InvProductName
		}
		if _, err := ma0.DB.Exec(
			`INSERT OR REPLACE INTO inventory
         (invDate, invYjCode, invJanCode, invProductName,
          invJanHousouSuuryouNumber, qty,
          HousouTaniUnit, InvHousouTaniUnit,
          janqty, JanHousouSuuryouUnit, InvJanHousouSuuryouUnit)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			rec.InvDate,
			rec.InvYjCode,
			rec.InvJanCode,
			prod,
			rec.InvJanHousouSuuryouNumber,
			rec.Qty,
			rec.HousouTaniUnit,
			rec.InvHousouTaniUnit,
			rec.JanQty,
			rec.JanHousouSuuryouUnit,
			rec.InvJanHousouSuuryouUnit,
		); err != nil {
			log.Printf("[INVENTORY] upsert error JAN=%s: %v", rec.InvJanCode, err)
		}
	}

	// 5) 結果を JSON で返却
	resp := map[string]interface{}{
		"count":       len(recs),
		"inventories": recs,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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

	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")

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
