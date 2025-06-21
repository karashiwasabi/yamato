package jchms

import (
	"encoding/csv"
	"io"
	"log"
)

// JCFields は、JCHMAS CSVから取り出す JC 領域の各フィールドを保持する構造体です。
// ここでは例として JC000, JC001, JC124 を定義していますが、必要に応じて他のフィールドも追加してください。
type JCFields struct {
	JC000 string
	JC001 string
	JC002 string
	JC003 string
	JC004 string
	JC005 string
	JC006 string
	JC007 string
	JC008 string
	JC009 string
	JC010 string
	JC011 string
	JC012 string
	JC013 string
	JC014 string
	JC015 string
	JC016 string
	JC017 string
	JC018 string
	JC019 string
	JC020 string
	JC021 string
	JC022 string
	JC023 string
	JC024 string
	JC025 string
	JC026 string
	JC027 string
	JC028 string
	JC029 string
	JC030 string
	JC031 string
	JC032 string
	JC033 string
	JC034 string
	JC035 string
	JC036 string
	JC037 string
	JC038 string
	JC039 string
	JC040 string
	JC041 string
	JC042 string
	JC043 string
	JC044 string
	JC045 string
	JC046 string
	JC047 string
	JC048 string
	JC049 string
	JC050 string
	JC051 string
	JC052 string
	JC053 string
	JC054 string
	JC055 string
	JC056 string
	JC057 string
	JC058 string
	JC059 string
	JC060 string
	JC061 string
	JC062 string
	JC063 string
	JC064 string
	JC065 string
	JC066 string
	JC067 string
	JC068 string
	JC069 string
	JC070 string
	JC071 string
	JC072 string
	JC073 string
	JC074 string
	JC075 string
	JC076 string
	JC077 string
	JC078 string
	JC079 string
	JC080 string
	JC081 string
	JC082 string
	JC083 string
	JC084 string
	JC085 string
	JC086 string
	JC087 string
	JC088 string
	JC089 string
	JC090 string
	JC091 string
	JC092 string
	JC093 string
	JC094 string
	JC095 string
	JC096 string
	JC097 string
	JC098 string
	JC099 string
	JC100 string
	JC101 string
	JC102 string
	JC103 string
	JC104 string
	JC105 string
	JC106 string
	JC107 string
	JC108 string
	JC109 string
	JC110 string
	JC111 string
	JC112 string
	JC113 string
	JC114 string
	JC115 string
	JC116 string
	JC117 string
	JC118 string
	JC119 string
	JC120 string
	JC121 string
	JC122 string
	JC123 string
	JC124 string
}

// JCHMASRecord は、JCHMAS（旧JCHMS）CSVの1行分を表します。
// キーはフィールド0（JANコード）とし、JC領域の情報を構造体として保持します。
type JCHMASRecord struct {
	JANCode string   // CSV のインデックス0 (JANコード)
	JC      JCFields // JC 領域の情報
}

// ParseJCHMAS は、入力の io.Reader から CSV データを読み込み、
// 各行を JCHMASRecord のスライスに変換して返します。
// CSV はヘッダーなしで、各行は少なくとも 125 フィールド（0〜124）を持つことが前提です。
func ParseJCHMAS(r io.Reader) ([]JCHMASRecord, error) {
	reader := csv.NewReader(r)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var records []JCHMASRecord
	for i, row := range rows {
		// 行のフィールド数が125未満の場合は、警告を出してその行をスキップ
		if len(row) < 125 {
			log.Printf("Row %d: insufficient fields (%d)", i+1, len(row))
			continue
		}

		// JC領域のデータを構造体に格納
		jc := JCFields{
			JC000: row[0],
			JC001: row[1],
			JC002: row[2],
			JC003: row[3],
			JC004: row[4],
			JC005: row[5],
			JC006: row[6],
			JC007: row[7],
			JC008: row[8],
			JC009: row[9],
			JC010: row[10],
			JC011: row[11],
			JC012: row[12],
			JC013: row[13],
			JC014: row[14],
			JC015: row[15],
			JC016: row[16],
			JC017: row[17],
			JC018: row[18],
			JC019: row[19],
			JC020: row[20],
			JC021: row[21],
			JC022: row[22],
			JC023: row[23],
			JC024: row[24],
			JC025: row[25],
			JC026: row[26],
			JC027: row[27],
			JC028: row[28],
			JC029: row[29],
			JC030: row[30],
			JC031: row[31],
			JC032: row[32],
			JC033: row[33],
			JC034: row[34],
			JC035: row[35],
			JC036: row[36],
			JC037: row[37],
			JC038: row[38],
			JC039: row[39],
			JC040: row[40],
			JC041: row[41],
			JC042: row[42],
			JC043: row[43],
			JC044: row[44],
			JC045: row[45],
			JC046: row[46],
			JC047: row[47],
			JC048: row[48],
			JC049: row[49],
			JC050: row[50],
			JC051: row[51],
			JC052: row[52],
			JC053: row[53],
			JC054: row[54],
			JC055: row[55],
			JC056: row[56],
			JC057: row[57],
			JC058: row[58],
			JC059: row[59],
			JC060: row[60],
			JC061: row[61],
			JC062: row[62],
			JC063: row[63],
			JC064: row[64],
			JC065: row[65],
			JC066: row[66],
			JC067: row[67],
			JC068: row[68],
			JC069: row[69],
			JC070: row[70],
			JC071: row[71],
			JC072: row[72],
			JC073: row[73],
			JC074: row[74],
			JC075: row[75],
			JC076: row[76],
			JC077: row[77],
			JC078: row[78],
			JC079: row[79],
			JC080: row[80],
			JC081: row[81],
			JC082: row[82],
			JC083: row[83],
			JC084: row[84],
			JC085: row[85],
			JC086: row[86],
			JC087: row[87],
			JC088: row[88],
			JC089: row[89],
			JC090: row[90],
			JC091: row[91],
			JC092: row[92],
			JC093: row[93],
			JC094: row[94],
			JC095: row[95],
			JC096: row[96],
			JC097: row[97],
			JC098: row[98],
			JC099: row[99],
			JC100: row[100],
			JC101: row[101],
			JC102: row[102],
			JC103: row[103],
			JC104: row[104],
			JC105: row[105],
			JC106: row[106],
			JC107: row[107],
			JC108: row[108],
			JC109: row[109],
			JC110: row[110],
			JC111: row[111],
			JC112: row[112],
			JC113: row[113],
			JC114: row[114],
			JC115: row[115],
			JC116: row[116],
			JC117: row[117],
			JC118: row[118],
			JC119: row[119],
			JC120: row[120],
			JC121: row[121],
			JC122: row[122],
			JC123: row[123],
			JC124: row[124],
		}

		// レコードを生成（キーは row[0] の JANコード）
		record := JCHMASRecord{
			JANCode: row[0],
			JC:      jc,
		}
		records = append(records, record)
	}
	return records, nil
}
