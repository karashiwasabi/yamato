// dat/dat.go
package dat

import (
	"YAMATO/jcshms"
	"YAMATO/ma0"
	"YAMATO/model"
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// getOrganizedFlag は指定された JAN コードについて、JCShms マスターに存在するかをチェックし、
// 存在すれば 1 (organized)、存在しなければ 0 (disorganized) を返します。
func getOrganizedFlag(jan string) (int, error) {
	records, err := jcshms.QueryByJan(ma0.DB, jan)
	if err != nil {
		return 0, fmt.Errorf("jcshms.QueryByJan error: %w", err)
	}
	if len(records) > 0 {
		return 1, nil
	}
	return 0, nil
}

// ProcessDATRecord は、DATファイルから抽出されたフィールド情報のスライスを用い、
// MA0 連携処理を実行します。data の 5 番目（インデックス4）が JAN コードである前提です。
// ※この呼び出しは MA0 連携の副作用として利用される（ログ出力など）。
func ProcessDATRecord(data []string) (bool, error) {
	if len(data) < 5 {
		return false, nil
	}
	jan := data[4]
	// MA0 連携処理を行い、新規作成の場合は created==true となりますが、
	// 整理状態は後述の JCShms チェックで判定するため、ここでは副作用のみ利用します。
	_, created, err := ma0.CheckOrCreateMA0(jan)
	if err != nil {
		return false, err
	}
	if created {
		log.Printf("[DAT] New MA0 record created for JAN: %q", jan)
	}
	return created, nil
}

// ParseDATFile は、io.Reader から DAT ファイルを読み込み、
// 固定長文字列フォーマットに従って各行を解析して model.DATRecord のスライスを返します。
// さらに、各レコードにつき MA0 連携のための処理を実施し、
// JCShms マスターによる整理状態（organizedFlag）を取得して datrecords テーブルへ INSERT します。
func ParseDATFile(r io.Reader) (records []model.DATRecord, totalCount, ma0CreatedCount, duplicateCount int, err error) {
	scanner := bufio.NewScanner(r)
	var currentOroshiCode string

	// getField は、固定長文字列から指定位置の部分文字列を返します。
	getField := func(s string, start, end int) string {
		if len(s) >= end {
			return s[start:end]
		} else if len(s) > start {
			return s[start:]
		}
		return ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		// 行長が短い場合はスキップ
		if len(line) < 3 {
			continue
		}

		// "S20" 行の場合、卸コード（位置3〜12）を取得して currentOroshiCode に保存
		if strings.HasPrefix(line, "S20") {
			currentOroshiCode = strings.TrimSpace(line[3:12])
			continue
		}

		// "D20" 行のみを対象とする
		if !strings.HasPrefix(line, "D20") {
			continue
		}
		totalCount++

		// 固定長フォーマットに従い各フィールドを抜き出し
		datDate := getField(line, 4, 12)
		datFlag := getField(line, 3, 4)
		datRecNo := getField(line, 12, 22)
		datLineNo := getField(line, 22, 24)
		datJan := getField(line, 25, 38)
		rawName := getField(line, 38, 78)
		// Shift‑JIS から UTF‑8 へ変換（エラー時は rawName をそのまま利用）
		name, _, convErr := transform.String(japanese.ShiftJIS.NewDecoder(), rawName)
		if convErr != nil {
			name = rawName
		}
		datQty := getField(line, 78, 83)
		datUnit := getField(line, 83, 92)
		datSub := getField(line, 92, 101)
		datPkg := getField(line, 101, 109)
		datExp := getField(line, 109, 115)
		datLot := getField(line, 115, 121)

		// model.DATRecord を生成
		rec := model.DATRecord{
			CurrentOroshiCode: currentOroshiCode,
			DatDate:           datDate,
			DatFlag:           datFlag,
			DatRecNo:          datRecNo,
			DatJan:            datJan,
			DatLineNo:         datLineNo,
			DatProductName:    name,
			DatQty:            datQty,
			DatUnit:           datUnit,
			DatSub:            datSub,
			DatPkg:            datPkg,
			DatExp:            datExp,
			DatLot:            datLot,
		}
		records = append(records, rec)

		// MA0 連携用データスライスを作成（フィールドの順序は既存の実装と同様）
		dataSlice := []string{
			currentOroshiCode,
			datDate,
			datFlag,
			datRecNo,
			datJan,
			datLineNo,
			name,
			datQty,
			datUnit,
			datSub,
			datPkg,
			datExp,
			datLot,
		}

		// MA0 連携の副作用として実行（結果はここでは利用せず）
		_, procErr := ProcessDATRecord(dataSlice)
		if procErr != nil {
			return records, totalCount, ma0CreatedCount, duplicateCount, procErr
		}

		// JCShms マスターを用いて整理状態フラグを取得（USAGE と同様の判定基準）
		flag, err := getOrganizedFlag(datJan)
		if err != nil {
			log.Printf("[DAT] Organized flag 確認エラー (JAN=%q): %v", datJan, err)
			flag = 0
		}
		organizedFlag := flag

		// ※ 統計用途としてカウントするなら、以下のようにカウントを更新（※名称は任意）
		if organizedFlag == 1 {
			ma0CreatedCount++
		} else {
			duplicateCount++
		}

		// DB 登録：InsertDATRecord は model.DATRecord と整理状態フラグを引数に取り、
		// datrecords テーブルへ INSERT を行います。
		if err := ma0.InsertDATRecord(ma0.DB, rec, organizedFlag); err != nil {
			log.Printf("Error inserting DATRecord: %v", err)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
	}
	return records, totalCount, ma0CreatedCount, duplicateCount, err
}
