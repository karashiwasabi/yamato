package inventory

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"YAMATO/ma0"
	"YAMATO/tani"
	"YAMATO/usage"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// trimQS は前後のダブルクォート、シングルクォート、空白を削ります。
func trimQS(s string) string {
	return strings.Trim(s, `"' `)
}

var (
	// codeToName はコード→名称マップ
	codeToName = usage.GetTaniMap()
	// nameToCode は名称→コード逆マップ
	nameToCode = tani.BuildNameToCodeMap(codeToName)
)

// InventoryRecord は CSV の１行分を表し、schema.sql の inventory 11列に対応します。
type InventoryRecord struct {
	InvDate                   string  // 棚卸日
	InvYjCode                 string  // YJコード
	InvJanCode                string  // JANコード
	InvProductName            string  // 商品名
	InvJanHousouSuuryouNumber float64 // JAN包装数量（数字）
	Qty                       float64 // 在庫数（包装単位）
	HousouTaniUnit            string  // 包装単位（名称）
	InvHousouTaniUnit         string  // 包装単位（コード）
	JanQty                    float64 // 在庫数（JAN包装単位）
	JanHousouSuuryouUnit      string  // JAN包装数量単位（名称）
	InvJanHousouSuuryouUnit   string  // JAN包装数量単位（コード）
}

func ParseInventoryCSV(r io.Reader) ([]InventoryRecord, error) {
	rd := csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	rd.LazyQuotes = true
	rd.FieldsPerRecord = -1

	// ── 1行目(H行)読み込み: parts[4] に棚卸日が入っている ──
	hrow, err := rd.Read()
	if err != nil {
		return nil, fmt.Errorf("inventory: H行読込エラー: %w", err)
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
			return nil, fmt.Errorf("inventory: csv read error: %w", err)
		}
		// 必要列数チェック
		if len(parts) <= 45 {
			continue
		}

		// ── 基本情報 ──
		yj := trimQS(parts[42])   // R1行 43列目
		jan := trimQS(parts[45])  // R1行 46列目
		name := trimQS(parts[12]) // R1行 13列目

		// ── 数量情報 ──
		// JAN包装数量: 18列目
		jps, _ := strconv.ParseFloat(trimQS(parts[17]), 64)
		// 基準数量: 15列目
		baseQty, _ := strconv.ParseFloat(trimQS(parts[21]), 64)
		// 在庫数(包装単位) = 基準数量 × JAN包装数量
		qty := baseQty * jps
		// 在庫数(JAN包装単位): 22列目
		jq, _ := strconv.ParseFloat(trimQS(parts[21]), 64)

		// ── 包装単位情報 ──
		rawPack := trimQS(parts[16]) // 17列目
		packUnit := rawPack
		if packUnit == "" {
			maRec, _, _ := ma0.CheckOrCreateMA0(jan)
			code := maRec.MA038JC038HousouTaniSuuchi
			if nm := usage.GetTaniName(code); nm != "" {
				packUnit = nm
			} else {
				packUnit = code
			}
		}
		packCode := nameToCode[packUnit]
		if packCode == "" {
			packCode = packUnit
		}

		// ── JAN包装数量単位情報 ──
		rawJanUnit := trimQS(parts[23]) // 24列目
		janUnit := rawJanUnit
		if janUnit == "" {
			maRec, _, _ := ma0.CheckOrCreateMA0(jan)
			code := maRec.MA132JA007HousouSuuryouTaniCode
			if nm := usage.GetTaniName(code); nm != "" {
				janUnit = nm
			} else {
				janUnit = code
			}
		}
		janCodeMap := nameToCode[janUnit]
		if janCodeMap == "" {
			janCodeMap = janUnit
		}

		recs = append(recs, InventoryRecord{
			InvDate:                   date,
			InvYjCode:                 yj,
			InvJanCode:                jan,
			InvProductName:            name,
			InvJanHousouSuuryouNumber: jps,
			Qty:                       qty,
			HousouTaniUnit:            packUnit,
			InvHousouTaniUnit:         packCode,
			JanQty:                    jq,
			JanHousouSuuryouUnit:      janUnit,
			InvJanHousouSuuryouUnit:   janCodeMap,
		})
	}

	return recs, nil
}
