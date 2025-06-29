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

// ParseInventoryCSV は Shift-JIS から UTF-8 変換しつつ CSV を読み込む
func ParseInventoryCSV(r io.Reader) ([]InventoryRecord, error) {
	rd := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	rd.LazyQuotes = true
	rd.FieldsPerRecord = -1

	// 1行目: H行から日付を取得
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

		yj := trimQS(parts[42])   // R43
		jan := trimQS(parts[45])  // R46
		name := trimQS(parts[12]) // R13

		jps, _ := strconv.ParseFloat(trimQS(parts[17]), 64)
		baseQty, _ := strconv.ParseFloat(trimQS(parts[21]), 64)
		qty := baseQty * jps
		janQty, _ := strconv.ParseFloat(trimQS(parts[21]), 64)

		rawPack := trimQS(parts[16])    // R17
		rawJanUnit := trimQS(parts[23]) // R24

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
			JanHousouSuuryouUnit:      rawJanUnit,
			InvJanHousouSuuryouUnit:   rawJanUnit,
		})
	}
	return recs, nil
}

// UploadInventoryHandler は棚卸CSVを受け取って
//  1. inventory テーブルにUPSERT
//  2. JCSHMS未登録のみMA2にUpsert
//
// を実行します。
func UploadInventoryHandler(w http.ResponseWriter, r *http.Request) {
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

	// CSVパース
	recs, err := ParseInventoryCSV(file)
	if err != nil {
		http.Error(w, "CSV読み込みエラー: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 名称→コードマップ
	nameToCode := tani.BuildNameToCodeMap(usage.GetTaniMap())

	for i := range recs {
		rec := &recs[i]

		// MA0 登録／取得（YJコード取得用）
		maRec, _, err := ma0.CheckOrCreateMA0(rec.InvJanCode)
		if err != nil {
			log.Printf("[INVENTORY] MA0 error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}
		rec.InvYjCode = maRec.MA009JC009YJCode

		// 在庫テーブル用：名称→コード
		rawPack := strings.Trim(rec.HousouTaniUnit, `"' `)
		if code, ok := nameToCode[rawPack]; ok {
			rec.InvHousouTaniUnit = code
		}
		rawJan := strings.Trim(rec.JanHousouSuuryouUnit, `"' `)
		if code, ok := nameToCode[rawJan]; ok {
			rec.InvJanHousouSuuryouUnit = code
		}

		// inventory UPSERT
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
			continue
		}

		// JCSHMS に未登録なら MA2 Upsert
		cs, err := jcshms.QueryByJan(ma0.DB, rec.InvJanCode)
		if err != nil {
			log.Printf("[INVENTORY] JCShms error JAN=%s: %v", rec.InvJanCode, err)
			continue
		}
		if len(cs) == 0 {
			// MA2 登録用レコード組立
			m2 := &ma2.Record{
				JanCode:                  rec.InvJanCode,
				Shouhinmei:               rec.InvProductName,
				HousouKeitai:             "", // CSVに無ければ空
				HousouTaniUnitName:       rec.HousouTaniUnit,
				HousouSouryouNumber:      0,
				JanHousouSuuryouNumber:   int(rec.InvJanHousouSuuryouNumber),
				JanHousouSuuryouUnitName: rec.JanHousouSuuryouUnit,
				JanHousouSouryouNumber:   0,
			}
			if err := ma2.Upsert(ma0.DB, m2); err != nil {
				log.Printf("[INVENTORY] MA2 Upsert error JAN=%s: %v", rec.InvJanCode, err)
			} else {
				// 結果のYJコードを在庫レコードにも反映
				rec.InvYjCode = m2.YjCode
			}
		}
	}

	// JSON応答
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":       len(recs),
		"inventories": recs,
	})
}
