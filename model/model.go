// model/model.go
package model

// DATRecord は、DAT ファイルの1行分の情報を表す共通データモデルです。
// 注意: DAT ファイルから生の文字列として読み込む場合は、初期値として
// "Name" フィールドに読み込まれますが、Shift‑JIS から UTF‑8 への変換後は
// この値を "DatProductName" にセットします。
type DATRecord struct {
	CurrentOroshiCode string `json:"DatOroshiCode"`         // 卸コード
	DatDate           string `json:"DatDate"`               // 日付
	DatFlag           string `json:"DatDeliveryFlag"`       // 納品／返品フラグ
	DatRecNo          string `json:"DatReceiptNumber"`      // 伝票番号
	DatJan            string `json:"DatJanCode"`            // JANコード
	DatLineNo         string `json:"DatLineNumber"`         // 行番号
	DatProductName    string `json:"DatProductName"`        // 商品名（変換後の値）
	DatQty            string `json:"DatQuantity"`           // 数量
	DatUnit           string `json:"DatUnitPrice"`          // 単価または単位
	DatSub            string `json:"DatSubtotal"`           // 小計
	DatPkg            string `json:"DatPackagingDrugPrice"` // 包装薬価
	DatExp            string `json:"DatExpiryDate"`         // 賞味期限
	DatLot            string `json:"DatLotNumber"`          // ロット番号
}

// USAGERecord は、USAGE CSV の1行分の情報を表します。
type USAGERecord struct {
	UsageDate        string // 使用日
	UsageYjCode      string // YJコード
	UsageJanCode     string // JANコード（MA0のキーとして利用）
	UsageProductName string // 商品名
	UsageAmount      string // 数量または金額
	UsageUnit        string // 単位コード
	UsageUnitName    string // 単位名称（TANI マップ経由で解決）
	OrganizedFlag    int    // 1: organized, 0: disorganized
}
