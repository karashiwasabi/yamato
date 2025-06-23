// File: jchms/jcshms.go
package jcshms

import (
	"database/sql"
	"fmt"
	"reflect"
)

// JCFields holds 125 columns from the JCHMAS CSV.
// CSV取り込み時、テーブル jchmas の列は「JC000JanCode, JC001, JC002, …, JC124」として保存される前提です。
type JCFields struct {
	JC000JanCode                           string
	JC001JanCodeShikibetsuKubun            string
	JC002KyuuJanCode                       string
	JC003TouitsuShouhinCode                string
	JC004YakkaKijunShuusaiIyakuhinCode     string
	JC005KyuuYakkaKijunShuusaiIyakuhinCode string
	JC006HOTBangou                         string
	JC007ReseputoCode1                     string
	JC008ReseputoCode2                     string
	JC009YJCode                            string
	JC010YakkouBunruiCode                  string
	JC011YakkouBunruiMei                   string
	JC012ShiyouKubunCode                   string
	JC013ShiyouKubunMeishou                string
	JC014NihonHyoujunShouhinBunruiBangou   string
	JC015ZaikeiCode                        string
	JC016ZaikeiKigou                       string
	JC017ZaikeiMeishou                     string
	JC018ShouhinMei                        string
	JC019HankakuShouhinMei                 string
	JC020KikakuYouryou                     string
	JC021HankakuKikakuYouryou              string
	JC022ShouhinMeiKanaSortYou             string
	JC023ShouhinMeiKanpouYouKigou          string
	JC024IppanMeishou                      string
	JC025YakkaShuusaiMeishou               string
	JC026ReseYouIyakuhinMei                string
	JC027KikakuTaniMeishou                 string
	JC028KikakuTaniKigou                   string
	JC029HanbaiMotoCode                    string
	JC030HanbaiMotoMei                     string
	JC031HanbaiMotoMeiKana                 string
	JC032HanbaiMotoMeiRyakuMei             string
	JC033SeizouMotoYunyuuMotoCode          string
	JC034SeizouMotoYunyuuMotoMei           string
	JC035SeizouMotoYunyuuMotoMeiKana       string
	JC036SeizouMotoYunyuuMotoMeiRyakuMei   string
	JC037HousouKeitai                      string
	JC038HousouTaniSuuchi                  string
	JC039HousouTaniTani                    string
	JC040HousouSuuryouSuuchi               string
	JC041HousouSuuryouTani                 string
	JC042HousouIrisuuSuuchi                string
	JC043HousouIrisuuTani                  string
	JC044HousouSouryouSuuchi               string
	JC045HousouSouryouTani                 string
	JC046HousouYouryouSuuchi               string
	JC047HousouYouryouTani                 string
	JC048HousouYakkaKeisuu                 string
	JC049GenTaniYakka                      string
	JC050GenHousouYakka                    string
	JC051KyuuTaniYakka                     string
	JC052KyuuHousouYakka                   string
	JC053KokuchiTaniYakka                  string
	JC054KokuchiHousouYakka                string
	JC055YakkaKaiteiNengappi               string
	JC056YakkaShuusaiNengappi              string
	JC057HanbaiKaishiNengappi              string
	JC058KeikaSochiNengappi                string
	JC059HatsubaiChuushiNengappi           string
	JC060SeizouChuushiNengappi             string
	JC061Doyaku                            string
	JC062Gekiyaku                          string
	JC063Mayaku                            string
	JC064Kouseishinyaku                    string
	JC065Kakuseizai                        string
	JC066KakuseizaiGenryou                 string
	JC067ShuukanseiIyakuhin                string
	JC068ShiteiIyakuhinKyuuKiseiKubun      string
	JC069YoushijiIyakuhinKyuuKiseiKubun    string
	JC070KetsuekiSeizai                    string
	JC071NihonYakkyokuhou                  string
	JC072YuukouKikan                       string
	JC073ShiyouKigen                       string
	JC074SeibutsuYuraiSeihin               string
	JC075Kouhatsuhin                       string
	JC076YakkaKijunShuusaiKubun            string
	JC077KichouGimuKubun                   string
	JC078ShouhinKubun                      string
	JC079ShohousenIyakuhin                 string
	JC080ChuushiRiyuuKubun                 string
	JC081MishiyouKyuuRyuutsuuKanrihin      string
	JC082MentenanceKubun                   string
	JC083KouhatsuhinNoAruSenpatsuhinKubun  string
	JC084AuthorizedGeneric                 string
	JC085Biosimilar                        string
	JC086HighRiskYaku                      string
	JC087Kuuran1                           string
	JC088Kuuran2                           string
	JC089Shitsuon                          string
	JC090Reisho                            string
	JC091Reizou                            string
	JC092Reitou                            string
	JC093Ansho                             string
	JC094Shakou                            string
	JC095KimitsuYouki                      string
	JC096MippuuYouki                       string
	JC097Kikenbutsu                        string
	JC098OndoJougen                        string
	JC099OndoKagen                         string
	JC100SonotaHokanjouNoChui              string
	JC101KonpouJuuryouSizeJouhou           string
	JC102KonpouTateSizeJouhou              string
	JC103KonpouYokoSizeJouhou              string
	JC104KonpouTakasaSizeJouhou            string
	JC105KonpouTaiseiSizeJouhou            string
	JC106ChuubakoJuuryouSizeJouhou         string
	JC107ChuubakoTateSizeJouhou            string
	JC108ChuubakoYokoSizeJouhou            string
	JC109ChuubakoTakasaSizeJouhou          string
	JC110ChuubakoTaiseiSizeJouhou          string
	JC111KousouJuuryouSizeJouhou           string
	JC112KousouTateSizeJouhou              string
	JC113KousouYokoSizeJouhou              string
	JC114KousouTakasaSizeJouhou            string
	JC115KousouTaiseiSizeJouhou            string
	JC116KonpouTaniSizeJouhou              string
	JC117HacchuuTaniSizeJouhou             string
	JC118KoushinKubun                      string
	JC119TourokuNengappi                   string
	JC120KoushinNengappi                   string
	JC121ChouzaiHousouTaniCode             string
	JC122HanbaiHousouTaniCode              string
	JC123IppanMeiKana                      string
	JC124SaishouYakkaKansanKeisuu          string
}

// JCHMASRecord represents one record in table jchmas.
type JCSHMSRecord struct {
	// JANCode は、SELECT 文で「JC000JanCode AS JC000」により取得された値です。
	JC000JanCode string
	// JC は、CSVの125フィールドを保持します。
	JC JCFields
}

// QueryJCHMASRecordsByJan queries the jchmas table for records matching the JAN code.
// SELECT 句では、列「JC000JanCode」をエイリアス「JC000」として取得します。
func QueryJCSHMSRecordsByJan(db *sql.DB, jan string) ([]JCSHMSRecord, error) {
	query := `
        SELECT 
JC000JanCode,
JC001JanCodeShikibetsuKubun,
JC002KyuuJanCode,
JC003TouitsuShouhinCode,
JC004YakkaKijunShuusaiIyakuhinCode,
JC005KyuuYakkaKijunShuusaiIyakuhinCode,
JC006HOTBangou,
JC007ReseputoCode1,
JC008ReseputoCode2,
JC009YJCode,
JC010YakkouBunruiCode,
JC011YakkouBunruiMei,
JC012ShiyouKubunCode,
JC013ShiyouKubunMeishou,
JC014NihonHyoujunShouhinBunruiBangou,
JC015ZaikeiCode,
JC016ZaikeiKigou,
JC017ZaikeiMeishou,
JC018ShouhinMei,
JC019HankakuShouhinMei,
JC020KikakuYouryou,
JC021HankakuKikakuYouryou,
JC022ShouhinMeiKanaSortYou,
JC023ShouhinMeiKanpouYouKigou,
JC024IppanMeishou,
JC025YakkaShuusaiMeishou,
JC026ReseYouIyakuhinMei,
JC027KikakuTaniMeishou,
JC028KikakuTaniKigou,
JC029HanbaiMotoCode,
JC030HanbaiMotoMei,
JC031HanbaiMotoMeiKana,
JC032HanbaiMotoMeiRyakuMei,
JC033SeizouMotoYunyuuMotoCode,
JC034SeizouMotoYunyuuMotoMei,
JC035SeizouMotoYunyuuMotoMeiKana,
JC036SeizouMotoYunyuuMotoMeiRyakuMei,
JC037HousouKeitai,
JC038HousouTaniSuuchi,
JC039HousouTaniTani,
JC040HousouSuuryouSuuchi,
JC041HousouSuuryouTani,
JC042HousouIrisuuSuuchi,
JC043HousouIrisuuTani,
JC044HousouSouryouSuuchi,
JC045HousouSouryouTani,
JC046HousouYouryouSuuchi,
JC047HousouYouryouTani,
JC048HousouYakkaKeisuu,
JC049GenTaniYakka,
JC050GenHousouYakka,
JC051KyuuTaniYakka,
JC052KyuuHousouYakka,
JC053KokuchiTaniYakka,
JC054KokuchiHousouYakka,
JC055YakkaKaiteiNengappi,
JC056YakkaShuusaiNengappi,
JC057HanbaiKaishiNengappi,
JC058KeikaSochiNengappi,
JC059HatsubaiChuushiNengappi,
JC060SeizouChuushiNengappi,
JC061Doyaku,
JC062Gekiyaku,
JC063Mayaku,
JC064Kouseishinyaku,
JC065Kakuseizai,
JC066KakuseizaiGenryou,
JC067ShuukanseiIyakuhin,
JC068ShiteiIyakuhinKyuuKiseiKubun,
JC069YoushijiIyakuhinKyuuKiseiKubun,
JC070KetsuekiSeizai,
JC071NihonYakkyokuhou,
JC072YuukouKikan,
JC073ShiyouKigen,
JC074SeibutsuYuraiSeihin,
JC075Kouhatsuhin,
JC076YakkaKijunShuusaiKubun,
JC077KichouGimuKubun,
JC078ShouhinKubun,
JC079ShohousenIyakuhin,
JC080ChuushiRiyuuKubun,
JC081MishiyouKyuuRyuutsuuKanrihin,
JC082MentenanceKubun,
JC083KouhatsuhinNoAruSenpatsuhinKubun,
JC084AuthorizedGeneric,
JC085Biosimilar,
JC086HighRiskYaku,
JC087Kuuran1,
JC088Kuuran2,
JC089Shitsuon,
JC090Reisho,
JC091Reizou,
JC092Reitou,
JC093Ansho,
JC094Shakou,
JC095KimitsuYouki,
JC096MippuuYouki,
JC097Kikenbutsu,
JC098OndoJougen,
JC099OndoKagen,
JC100SonotaHokanjouNoChui,
JC101KonpouJuuryouSizeJouhou,
JC102KonpouTateSizeJouhou,
JC103KonpouYokoSizeJouhou,
JC104KonpouTakasaSizeJouhou,
JC105KonpouTaiseiSizeJouhou,
JC106ChuubakoJuuryouSizeJouhou,
JC107ChuubakoTateSizeJouhou,
JC108ChuubakoYokoSizeJouhou,
JC109ChuubakoTakasaSizeJouhou,
JC110ChuubakoTaiseiSizeJouhou,
JC111KousouJuuryouSizeJouhou,
JC112KousouTateSizeJouhou,
JC113KousouYokoSizeJouhou,
JC114KousouTakasaSizeJouhou,
JC115KousouTaiseiSizeJouhou,
JC116KonpouTaniSizeJouhou,
JC117HacchuuTaniSizeJouhou,
JC118KoushinKubun,
JC119TourokuNengappi,
JC120KoushinNengappi,
JC121ChouzaiHousouTaniCode,
JC122HanbaiHousouTaniCode,
JC123IppanMeiKana,
JC124SaishouYakkaKansanKeisuu
        FROM jchmas
        WHERE JC000JanCode = ?
    `
	rows, err := db.Query(query, jan)
	if err != nil {
		return nil, fmt.Errorf("jcshms query error: %v", err)
	}
	defer rows.Close()

	const colsCount = 125
	var records []JCSHMSRecord
	for rows.Next() {
		columns := make([]interface{}, colsCount)
		columnPtrs := make([]interface{}, colsCount)
		for i := 0; i < colsCount; i++ {
			columnPtrs[i] = &columns[i]
		}
		if err := rows.Scan(columnPtrs...); err != nil {
			return nil, fmt.Errorf("jcshms scan error: %v", err)
		}

		var rec JCSHMSRecord
		// 最初のカラム（エイリアス済みの JC000）を JANCode として取得
		if b, ok := columns[0].([]byte); ok {
			rec.JC000JanCode = string(b)
		} else if columns[0] != nil {
			rec.JC000JanCode = columns[0].(string)
		}

		var jf JCFields
		jfVal := reflect.ValueOf(&jf).Elem()
		for i := 0; i < colsCount; i++ {
			var colStr string
			if b, ok := columns[i].([]byte); ok {
				colStr = string(b)
			} else if columns[i] != nil {
				colStr = columns[i].(string)
			}
			if i < jfVal.NumField() && jfVal.Field(i).CanSet() {
				jfVal.Field(i).SetString(colStr)
			} else {
				return nil, fmt.Errorf("failed to set JCFields field index %d", i)
			}
		}
		rec.JC = jf
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jchmas rows error: %v", err)
	}
	return records, nil
}
