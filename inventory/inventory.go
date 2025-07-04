// C:/Dev/YAMATO/inventory/inventory.go
package inventory

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"YAMATO/jcshms"
	"YAMATO/ma0"
	"YAMATO/ma2"
	"YAMATO/tani"
	"YAMATO/usage"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// trimQS は前後のクォート・空白を削除
func trimQS(s string) string {
	return strings.Trim(s, `"' `)
}

// InventoryRecord は CSV の１行分
type InventoryRecord struct {
	InvDate                   string  // 棚卸日
	InvYjCode                 string  // YJコード
	InvJanCode                string  // JANコード
	InvProductName            string  // 商品名
	InvJanHousouSuuryouNumber float64 // JAN包装数量
	Qty                       float64 // 在庫数(包装単位)
	HousouTaniUnit            string  // 包装単位(名称)
	InvHousouTaniUnit         string  // 包装単位(コード)
	JanQty                    float64 // 在庫数(JAN包装単位)
	JanHousouSuuryouUnit      string  // JAN包装数量単位(名称)
	InvJanHousouSuuryouUnit   string  // JAN包装数量単位(コード)
}

// ParseInventoryCSV は Shift-JIS→UTF-8 変換しつつ CSV を読み込み、
// 第17列（包装単位）・第24列（JAN包装単位）の
// “'” を含む加工前後をログ出力します。
func ParseInventoryCSV(r io.Reader) ([]InventoryRecord, error) {
	rd := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	rd.LazyQuotes = true
	rd.FieldsPerRecord = -1

	// ヘッダ行取得
	hrow, err := rd.Read()
	if err != nil {
		return nil, fmt.Errorf("inventory: H行読み込みエラー: %w", err)
	}
	if len(hrow) <= 4 {
		return nil, fmt.Errorf("inventory: H行の列不足")
	}
	date := trimQS(hrow[4])

	var recs []InventoryRecord
	for {
		parts, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("inventory: CSV読み込みエラー: %w", err)
		}
		if len(parts) <= 45 {
			continue
		}

		// 元データ（クォート含む）
		origPackField := parts[16] // R17 包装単位
		origJanField := parts[23]  // R24 JAN包装単位
		log.Printf(
			"[ParseInventoryCSV] origPackField=%q origJanField=%q",
			origPackField, origJanField,
		)

		// トリム前／後比較
		rawPack := strings.ReplaceAll(origPackField, "　", "")
		rawPack = trimQS(rawPack)
		rawJan := strings.ReplaceAll(origJanField, "　", "")
		rawJan = trimQS(rawJan)
		log.Printf(
			"[ParseInventoryCSV] trimmed HousouTaniUnit=%q JanHousouSuuryouUnit=%q",
			rawPack, rawJan,
		)

		// 他フィールド読み取り（省略可）
		jps, _ := strconv.ParseFloat(trimQS(parts[17]), 64)
		baseQty, _ := strconv.ParseFloat(trimQS(parts[21]), 64)
		qty := baseQty * jps
		janQty, _ := strconv.ParseFloat(trimQS(parts[21]), 64)
		yj := trimQS(parts[42])
		jan := trimQS(parts[45])
		name := trimQS(parts[12])

		recs = append(recs, InventoryRecord{
			InvDate:                   date,
			InvYjCode:                 yj,
			InvJanCode:                jan,
			InvProductName:            name,
			InvJanHousouSuuryouNumber: jps,
			Qty:                       qty,
			HousouTaniUnit:            rawPack,
			InvHousouTaniUnit:         rawPack,
			JanQty:                    janQty,
			JanHousouSuuryouUnit:      rawJan,
			InvJanHousouSuuryouUnit:   rawJan,
		})
	}
	return recs, nil
}

// UploadInventoryHandler は棚卸CSVのアップロードを受け取り、
// 単位マッピング前後をログ出力しつつDBにUPSERT、JSONを返します。
func UploadInventoryHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[UploadInventoryHandler] start")

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

	// 1) CSV→構造体
	recs, err := ParseInventoryCSV(file)
	if err != nil {
		http.Error(w, "CSV読み込みエラー: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("[UploadInventoryHandler] parsed %d records", len(recs))

	// 2) 名称→コードマップ取得
	nameToCode := tani.BuildNameToCodeMap(usage.GetTaniMap())
	// 2a) マップキー一覧ログ
	var keys []string
	for k := range nameToCode {
		keys = append(keys, k)
	}
	log.Printf("[UploadInventoryHandler] nameToCode keys: %v", keys)

	// 3) レコードごとにマッピング前後をログ出力
	for i := range recs {
		rec := &recs[i]
		log.Printf(
			"[UploadInventoryHandler] #%d before mapping: HousouTaniUnit=%q JanHousouSuuryouUnit=%q",
			i, rec.HousouTaniUnit, rec.JanHousouSuuryouUnit,
		)

		// 包装単位→コード
		rawPack := strings.Trim(rec.HousouTaniUnit, `"' `)
		if code, ok := nameToCode[rawPack]; ok {
			rec.InvHousouTaniUnit = code
			log.Printf("[UploadInventoryHandler] #%d mapped pack: %q → %q", i, rawPack, code)
		} else {
			rec.InvHousouTaniUnit = ""
			log.Printf("[UploadInventoryHandler] #%d no map for pack %q", i, rawPack)
		}

		// JAN包装単位→コード
		rawJan := strings.Trim(rec.JanHousouSuuryouUnit, `"' `)
		if code, ok := nameToCode[rawJan]; ok {
			rec.InvJanHousouSuuryouUnit = code
			log.Printf("[UploadInventoryHandler] #%d mapped jan unit: %q → %q", i, rawJan, code)
		} else {
			rec.InvJanHousouSuuryouUnit = ""
			log.Printf("[UploadInventoryHandler] #%d no map for jan unit %q", i, rawJan)
		}

		// 以下、MA0登録・DB UPSERT・MA2登録は既存ロジック
		maRec, _, err := ma0.CheckOrCreateMA0(rec.InvJanCode, rec.InvProductName)
		if err != nil {
			log.Printf("[UploadInventoryHandler] MA0 error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}
		rec.InvYjCode = maRec.MA009JC009YJCode

		prod := maRec.MA018JC018ShouhinMei
		if prod == "" {
			prod = rec.InvProductName
		}
		_, err = ma0.DB.Exec(
			`INSERT OR REPLACE INTO inventory
              (invDate, invYjCode, invJanCode, invProductName,
               invJanHousouSuuryouNumber, qty,
               HousouTaniUnit, InvHousouTaniUnit,
               janqty, JanHousouSuuryouUnit, InvJanHousouSuuryouUnit)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			rec.InvDate, rec.InvYjCode, rec.InvJanCode, prod,
			rec.InvJanHousouSuuryouNumber, rec.Qty,
			rec.HousouTaniUnit, rec.InvHousouTaniUnit,
			rec.JanQty, rec.JanHousouSuuryouUnit, rec.InvJanHousouSuuryouUnit,
		)
		if err != nil {
			log.Printf("[UploadInventoryHandler] upsert error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}

		cs, err := jcshms.QueryByJan(ma0.DB, rec.InvJanCode)
		if err != nil {
			log.Printf("[UploadInventoryHandler] JCShms error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}
		if len(cs) == 0 {
			m2 := &ma2.Record{
				JanCode:                  rec.InvJanCode,
				Shouhinmei:               rec.InvProductName,
				HousouKeitai:             "",
				HousouTaniUnitName:       rec.HousouTaniUnit,
				HousouSouryouNumber:      0,
				JanHousouSuuryouNumber:   int(rec.InvJanHousouSuuryouNumber),
				JanHousouSuuryouUnitName: rec.JanHousouSuuryouUnit,
				JanHousouSouryouNumber:   0,
			}
			if err := ma2.Upsert(ma0.DB, m2); err != nil {
				log.Printf("[UploadInventoryHandler] MA2 Upsert error JAN=%s: %v", rec.InvJanCode, err)
			}
		}
	}

	// 4) レスポンス直前ログ
	log.Printf("[UploadInventoryHandler] returning %d records", len(recs))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":       len(recs),
		"inventories": recs,
	})
}
