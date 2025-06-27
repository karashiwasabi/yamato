package tani

import (
	"encoding/csv"
	"io"
	"log"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ParseTANI は、TANI CSV ファイルを Shift‑JIS から UTF‑8 に変換しながら読み込み、
// 各行のフィールド0（単位コード）をキー、フィールド1（単位名称）を値とするマップを返します。
func ParseTANI(r io.Reader) (map[string]string, error) {
	// Shift‑JIS → UTF‑8 変換を適用
	decoder := transform.NewReader(r, japanese.ShiftJIS.NewDecoder())
	reader := csv.NewReader(decoder)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	taniMap := make(map[string]string)
	for _, row := range records {
		if len(row) < 2 {
			log.Printf("TANI行のフィールド不足: %v", row)
			continue
		}
		code := row[0]
		unit := row[1]
		taniMap[code] = unit
	}
	return taniMap, nil
}

func BuildNameToCodeMap(codeToName map[string]string) map[string]string {
	nameToCode := make(map[string]string, len(codeToName))
	for code, name := range codeToName {
		nameToCode[name] = code
	}
	return nameToCode
}
