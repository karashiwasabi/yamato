package jcshms

import (
	"database/sql"
	"fmt"
)

// JCFields は jcshms テーブルの125フィールドを表す構造体です
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

// JCSHMSRecord は QueryJCSHMSRecordsByJan の返却型
type JCSHMSRecord struct {
	JC000JanCode string
	JC           JCFields
}

// QueryJCSHMSRecordsByJan は JAN コードを受けて jcshms テーブルを検索し、
// JCSHMSRecord スライスを返します
// QueryJCSHMSRecordsByJan は JAN コードを受け、該当する JCSHMSRecord を返します
func QueryJCSHMSRecordsByJan(db *sql.DB, jan string) ([]JCSHMSRecord, error) {
	const query = `


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
    FROM jcshms
    WHERE JC000JanCode = ?
  `

	rows, err := db.Query(query, jan)
	if err != nil {
		return nil, fmt.Errorf("jcshms query error: %w", err)
	}
	defer rows.Close()

	var out []JCSHMSRecord
	for rows.Next() {
		var rec JCSHMSRecord
		// Scan の順序は上記 SELECT とまったく同じ順番で書くこと
		if err := rows.Scan(
			&rec.JC.JC000JanCode,
			&rec.JC.JC001JanCodeShikibetsuKubun,
			&rec.JC.JC002KyuuJanCode,
			&rec.JC.JC003TouitsuShouhinCode,
			&rec.JC.JC004YakkaKijunShuusaiIyakuhinCode,
			&rec.JC.JC005KyuuYakkaKijunShuusaiIyakuhinCode,
			&rec.JC.JC006HOTBangou,
			&rec.JC.JC007ReseputoCode1,
			&rec.JC.JC008ReseputoCode2,
			&rec.JC.JC009YJCode,
			&rec.JC.JC010YakkouBunruiCode,
			&rec.JC.JC011YakkouBunruiMei,
			&rec.JC.JC012ShiyouKubunCode,
			&rec.JC.JC013ShiyouKubunMeishou,
			&rec.JC.JC014NihonHyoujunShouhinBunruiBangou,
			&rec.JC.JC015ZaikeiCode,
			&rec.JC.JC016ZaikeiKigou,
			&rec.JC.JC017ZaikeiMeishou,
			&rec.JC.JC018ShouhinMei,
			&rec.JC.JC019HankakuShouhinMei,
			&rec.JC.JC020KikakuYouryou,
			&rec.JC.JC021HankakuKikakuYouryou,
			&rec.JC.JC022ShouhinMeiKanaSortYou,
			&rec.JC.JC023ShouhinMeiKanpouYouKigou,
			&rec.JC.JC024IppanMeishou,
			&rec.JC.JC025YakkaShuusaiMeishou,
			&rec.JC.JC026ReseYouIyakuhinMei,
			&rec.JC.JC027KikakuTaniMeishou,
			&rec.JC.JC028KikakuTaniKigou,
			&rec.JC.JC029HanbaiMotoCode,
			&rec.JC.JC030HanbaiMotoMei,
			&rec.JC.JC031HanbaiMotoMeiKana,
			&rec.JC.JC032HanbaiMotoMeiRyakuMei,
			&rec.JC.JC033SeizouMotoYunyuuMotoCode,
			&rec.JC.JC034SeizouMotoYunyuuMotoMei,
			&rec.JC.JC035SeizouMotoYunyuuMotoMeiKana,
			&rec.JC.JC036SeizouMotoYunyuuMotoMeiRyakuMei,
			&rec.JC.JC037HousouKeitai,
			&rec.JC.JC038HousouTaniSuuchi,
			&rec.JC.JC039HousouTaniTani,
			&rec.JC.JC040HousouSuuryouSuuchi,
			&rec.JC.JC041HousouSuuryouTani,
			&rec.JC.JC042HousouIrisuuSuuchi,
			&rec.JC.JC043HousouIrisuuTani,
			&rec.JC.JC044HousouSouryouSuuchi,
			&rec.JC.JC045HousouSouryouTani,
			&rec.JC.JC046HousouYouryouSuuchi,
			&rec.JC.JC047HousouYouryouTani,
			&rec.JC.JC048HousouYakkaKeisuu,
			&rec.JC.JC049GenTaniYakka,
			&rec.JC.JC050GenHousouYakka,
			&rec.JC.JC051KyuuTaniYakka,
			&rec.JC.JC052KyuuHousouYakka,
			&rec.JC.JC053KokuchiTaniYakka,
			&rec.JC.JC054KokuchiHousouYakka,
			&rec.JC.JC055YakkaKaiteiNengappi,
			&rec.JC.JC056YakkaShuusaiNengappi,
			&rec.JC.JC057HanbaiKaishiNengappi,
			&rec.JC.JC058KeikaSochiNengappi,
			&rec.JC.JC059HatsubaiChuushiNengappi,
			&rec.JC.JC060SeizouChuushiNengappi,
			&rec.JC.JC061Doyaku,
			&rec.JC.JC062Gekiyaku,
			&rec.JC.JC063Mayaku,
			&rec.JC.JC064Kouseishinyaku,
			&rec.JC.JC065Kakuseizai,
			&rec.JC.JC066KakuseizaiGenryou,
			&rec.JC.JC067ShuukanseiIyakuhin,
			&rec.JC.JC068ShiteiIyakuhinKyuuKiseiKubun,
			&rec.JC.JC069YoushijiIyakuhinKyuuKiseiKubun,
			&rec.JC.JC070KetsuekiSeizai,
			&rec.JC.JC071NihonYakkyokuhou,
			&rec.JC.JC072YuukouKikan,
			&rec.JC.JC073ShiyouKigen,
			&rec.JC.JC074SeibutsuYuraiSeihin,
			&rec.JC.JC075Kouhatsuhin,
			&rec.JC.JC076YakkaKijunShuusaiKubun,
			&rec.JC.JC077KichouGimuKubun,
			&rec.JC.JC078ShouhinKubun,
			&rec.JC.JC079ShohousenIyakuhin,
			&rec.JC.JC080ChuushiRiyuuKubun,
			&rec.JC.JC081MishiyouKyuuRyuutsuuKanrihin,
			&rec.JC.JC082MentenanceKubun,
			&rec.JC.JC083KouhatsuhinNoAruSenpatsuhinKubun,
			&rec.JC.JC084AuthorizedGeneric,
			&rec.JC.JC085Biosimilar,
			&rec.JC.JC086HighRiskYaku,
			&rec.JC.JC087Kuuran1,
			&rec.JC.JC088Kuuran2,
			&rec.JC.JC089Shitsuon,
			&rec.JC.JC090Reisho,
			&rec.JC.JC091Reizou,
			&rec.JC.JC092Reitou,
			&rec.JC.JC093Ansho,
			&rec.JC.JC094Shakou,
			&rec.JC.JC095KimitsuYouki,
			&rec.JC.JC096MippuuYouki,
			&rec.JC.JC097Kikenbutsu,
			&rec.JC.JC098OndoJougen,
			&rec.JC.JC099OndoKagen,
			&rec.JC.JC100SonotaHokanjouNoChui,
			&rec.JC.JC101KonpouJuuryouSizeJouhou,
			&rec.JC.JC102KonpouTateSizeJouhou,
			&rec.JC.JC103KonpouYokoSizeJouhou,
			&rec.JC.JC104KonpouTakasaSizeJouhou,
			&rec.JC.JC105KonpouTaiseiSizeJouhou,
			&rec.JC.JC106ChuubakoJuuryouSizeJouhou,
			&rec.JC.JC107ChuubakoTateSizeJouhou,
			&rec.JC.JC108ChuubakoYokoSizeJouhou,
			&rec.JC.JC109ChuubakoTakasaSizeJouhou,
			&rec.JC.JC110ChuubakoTaiseiSizeJouhou,
			&rec.JC.JC111KousouJuuryouSizeJouhou,
			&rec.JC.JC112KousouTateSizeJouhou,
			&rec.JC.JC113KousouYokoSizeJouhou,
			&rec.JC.JC114KousouTakasaSizeJouhou,
			&rec.JC.JC115KousouTaiseiSizeJouhou,
			&rec.JC.JC116KonpouTaniSizeJouhou,
			&rec.JC.JC117HacchuuTaniSizeJouhou,
			&rec.JC.JC118KoushinKubun,
			&rec.JC.JC119TourokuNengappi,
			&rec.JC.JC120KoushinNengappi,
			&rec.JC.JC121ChouzaiHousouTaniCode,
			&rec.JC.JC122HanbaiHousouTaniCode,
			&rec.JC.JC123IppanMeiKana,
			&rec.JC.JC124SaishouYakkaKansanKeisuu,
		); err != nil {
			return nil, fmt.Errorf("jcshms scan error: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jcshms rows error: %w", err)
	}
	return out, nil
}

func QueryByJan(db *sql.DB, jan string) ([]JCFields, error) {
	recs, err := QueryJCSHMSRecordsByJan(db, jan)
	if err != nil {
		return nil, err
	}
	// JCSHMSRecord の中身（.JC）を抜き出す
	out := make([]JCFields, len(recs))
	for i, r := range recs {
		f := r.JC
		// 主キーであるJC000JanCodeも JCFields に含めたい場合はここで設定
		f.JC000JanCode = r.JC000JanCode
		out[i] = f
	}
	return out, nil
}
