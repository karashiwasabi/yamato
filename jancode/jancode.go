package jancode

import (
	"encoding/csv"
	"io"
	"log"
)

// JAFields は、JANCODE CSVから取り出す JA領域の各フィールドを保持する構造体です。
// 例として、JA000、JA001、JA029 を定義していますが、必要に応じて中間のフィールドも追加してください。
type JAFields struct {
	JA000 string
	JA001 string
	JA002 string
	JA003 string
	JA004 string
	JA005 string
	JA006 string
	JA007 string
	JA008 string
	JA009 string
	JA010 string
	JA011 string
	JA012 string
	JA013 string
	JA014 string
	JA015 string
	JA016 string
	JA017 string
	JA018 string
	JA019 string
	JA020 string
	JA021 string
	JA022 string
	JA023 string
	JA024 string
	JA025 string
	JA026 string
	JA027 string
	JA028 string
	JA029 string
}

// JANCODERecord は、JANCODE CSVの1行分のデータを表します。
// キーは、フィールド1の値（JANコード）として使用し、JA領域の情報を保持します。
type JANCODERecord struct {
	JANCode string   // CSVのインデックス1 (JANコード)
	JA      JAFields // JA領域の情報
}

// ParseJANCODE は、入力の io.Reader から CSV データを読み込み、
// ヘッダー行をスキップした上で各行を JANCODERecord のスライスに変換して返します。
// CSVは各行が少なくとも30フィールド（インデックス0～29）を持つことを前提としています。
func ParseJANCODE(r io.Reader) ([]JANCODERecord, error) {
	reader := csv.NewReader(r)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var records []JANCODERecord
	for i, row := range rows {
		// 1行目はヘッダーのためスキップ
		if i == 0 {
			continue
		}

		if len(row) < 30 {
			log.Printf("Row %d: insufficient fields (%d)", i+1, len(row))
			continue
		}

		ja := JAFields{
			JA000: row[0],
			JA001: row[1],
			JA002: row[2],
			JA003: row[3],
			JA004: row[4],
			JA005: row[5],
			JA006: row[6],
			JA007: row[7],
			JA008: row[8],
			JA009: row[9],
			JA010: row[10],
			JA011: row[11],
			JA012: row[12],
			JA013: row[13],
			JA014: row[14],
			JA015: row[15],
			JA016: row[16],
			JA017: row[17],
			JA018: row[18],
			JA019: row[19],
			JA020: row[20],
			JA021: row[21],
			JA022: row[22],
			JA023: row[23],
			JA024: row[24],
			JA025: row[25],
			JA026: row[26],
			JA027: row[27],
			JA028: row[28],
			JA029: row[29],
		}

		record := JANCODERecord{
			JANCode: row[1], // キーとしてフィールド1の値を使用
			JA:      ja,
		}

		records = append(records, record)
	}
	return records, nil
}
