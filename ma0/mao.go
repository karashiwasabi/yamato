// File: ma0/ma0.go
package ma0

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"YAMATO/jancode"
	"YAMATO/jcshms"
)

type MA0Record struct {
	MA000JC000JanCode                           string
	MA001JC001JanCodeShikibetsuKubun            string
	MA002JC002KyuuJanCode                       string
	MA003JC003TouitsuShouhinCode                string
	MA004JC004YakkaKijunShuusaiIyakuhinCode     string
	MA005JC005KyuuYakkaKijunShuusaiIyakuhinCode string
	MA006JC006HOTBangou                         string
	MA007JC007ReseputoCode1                     string
	MA008JC008ReseputoCode2                     string
	MA009JC009YJCode                            string
	MA010JC010YakkouBunruiCode                  string
	MA011JC011YakkouBunruiMei                   string
	MA012JC012ShiyouKubunCode                   string
	MA013JC013ShiyouKubunMeishou                string
	MA014JC014NihonHyoujunShouhinBunruiBangou   string
	MA015JC015ZaikeiCode                        string
	MA016JC016ZaikeiKigou                       string
	MA017JC017ZaikeiMeishou                     string
	MA018JC018ShouhinMei                        string
	MA019JC019HankakuShouhinMei                 string
	MA020JC020KikakuYouryou                     string
	MA021JC021HankakuKikakuYouryou              string
	MA022JC022ShouhinMeiKanaSortYou             string
	MA023JC023ShouhinMeiKanpouYouKigou          string
	MA024JC024IppanMeishou                      string
	MA025JC025YakkaShuusaiMeishou               string
	MA026JC026ReseYouIyakuhinMei                string
	MA027JC027KikakuTaniMeishou                 string
	MA028JC028KikakuTaniKigou                   string
	MA029JC029HanbaiMotoCode                    string
	MA030JC030HanbaiMotoMei                     string
	MA031JC031HanbaiMotoMeiKana                 string
	MA032JC032HanbaiMotoMeiRyakuMei             string
	MA033JC033SeizouMotoYunyuuMotoCode          string
	MA034JC034SeizouMotoYunyuuMotoMei           string
	MA035JC035SeizouMotoYunyuuMotoMeiKana       string
	MA036JC036SeizouMotoYunyuuMotoMeiRyakuMei   string
	MA037JC037HousouKeitai                      string
	MA038JC038HousouTaniSuuchi                  string
	MA039JC039HousouTaniTani                    string
	MA040JC040HousouSuuryouSuuchi               string
	MA041JC041HousouSuuryouTani                 string
	MA042JC042HousouIrisuuSuuchi                string
	MA043JC043HousouIrisuuTani                  string
	MA044JC044HousouSouryouSuuchi               string
	MA045JC045HousouSouryouTani                 string
	MA046JC046HousouYouryouSuuchi               string
	MA047JC047HousouYouryouTani                 string
	MA048JC048HousouYakkaKeisuu                 string
	MA049JC049GenTaniYakka                      string
	MA050JC050GenHousouYakka                    string
	MA051JC051KyuuTaniYakka                     string
	MA052JC052KyuuHousouYakka                   string
	MA053JC053KokuchiTaniYakka                  string
	MA054JC054KokuchiHousouYakka                string
	MA055JC055YakkaKaiteiNengappi               string
	MA056JC056YakkaShuusaiNengappi              string
	MA057JC057HanbaiKaishiNengappi              string
	MA058JC058KeikaSochiNengappi                string
	MA059JC059HatsubaiChuushiNengappi           string
	MA060JC060SeizouChuushiNengappi             string
	MA061JC061Doyaku                            string
	MA062JC062Gekiyaku                          string
	MA063JC063Mayaku                            string
	MA064JC064Kouseishinyaku                    string
	MA065JC065Kakuseizai                        string
	MA066JC066KakuseizaiGenryou                 string
	MA067JC067ShuukanseiIyakuhin                string
	MA068JC068ShiteiIyakuhinKyuuKiseiKubun      string
	MA069JC069YoushijiIyakuhinKyuuKiseiKubun    string
	MA070JC070KetsuekiSeizai                    string
	MA071JC071NihonYakkyokuhou                  string
	MA072JC072YuukouKikan                       string
	MA073JC073ShiyouKigen                       string
	MA074JC074SeibutsuYuraiSeihin               string
	MA075JC075Kouhatsuhin                       string
	MA076JC076YakkaKijunShuusaiKubun            string
	MA077JC077KichouGimuKubun                   string
	MA078JC078ShouhinKubun                      string
	MA079JC079ShohousenIyakuhin                 string
	MA080JC080ChuushiRiyuuKubun                 string
	MA081JC081MishiyouKyuuRyuutsuuKanrihin      string
	MA082JC082MentenanceKubun                   string
	MA083JC083KouhatsuhinNoAruSenpatsuhinKubun  string
	MA084JC084AuthorizedGeneric                 string
	MA085JC085Biosimilar                        string
	MA086JC086HighRiskYaku                      string
	MA087JC087Kuuran1                           string
	MA088JC088Kuuran2                           string
	MA089JC089Shitsuon                          string
	MA090JC090Reisho                            string
	MA091JC091Reizou                            string
	MA092JC092Reitou                            string
	MA093JC093Ansho                             string
	MA094JC094Shakou                            string
	MA095JC095KimitsuYouki                      string
	MA096JC096MippuuYouki                       string
	MA097JC097Kikenbutsu                        string
	MA098JC098OndoJougen                        string
	MA099JC099OndoKagen                         string
	MA100JC100SonotaHokanjouNoChui              string
	MA101JC101KonpouJuuryouSizeJouhou           string
	MA102JC102KonpouTateSizeJouhou              string
	MA103JC103KonpouYokoSizeJouhou              string
	MA104JC104KonpouTakasaSizeJouhou            string
	MA105JC105KonpouTaiseiSizeJouhou            string
	MA106JC106ChuubakoJuuryouSizeJouhou         string
	MA107JC107ChuubakoTateSizeJouhou            string
	MA108JC108ChuubakoYokoSizeJouhou            string
	MA109JC109ChuubakoTakasaSizeJouhou          string
	MA110JC110ChuubakoTaiseiSizeJouhou          string
	MA111JC111KousouJuuryouSizeJouhou           string
	MA112JC112KousouTateSizeJouhou              string
	MA113JC113KousouYokoSizeJouhou              string
	MA114JC114KousouTakasaSizeJouhou            string
	MA115JC115KousouTaiseiSizeJouhou            string
	MA116JC116KonpouTaniSizeJouhou              string
	MA117JC117HacchuuTaniSizeJouhou             string
	MA118JC118KoushinKubun                      string
	MA119JC119TourokuNengappi                   string
	MA120JC120KoushinNengappi                   string
	MA121JC121ChouzaiHousouTaniCode             string
	MA122JC122HanbaiHousouTaniCode              string
	MA123JC123IppanMeiKana                      string
	MA124JC124SaishouYakkaKansanKeisuu          string
	MA125JA000                                  string
	MA126JA001JanCode                           string
	MA127JA002                                  string
	MA128JA003                                  string
	MA129JA004                                  string
	MA130JA005                                  string
	MA131JA006HousouSuuryouSuuchi               string
	MA132JA007HousouSuuryouTaniCode             string
	MA133JA008HousouSouryouSuuchi               string
	MA134JA009                                  string
	MA135JA010                                  string
	MA136JA011                                  string
	MA137JA012                                  string
	MA138JA013                                  string
	MA139JA014                                  string
	MA140JA015                                  string
	MA141JA016                                  string
	MA142JA017                                  string
	MA143JA018                                  string
	MA144JA019                                  string
	MA145JA020                                  string
	MA146JA021                                  string
	MA147JA022                                  string
	MA148JA023                                  string
	MA149JA024                                  string
	MA150JA025                                  string
	MA151JA026                                  string
	MA152JA027                                  string
	MA153JA028                                  string
	MA154JA029                                  string
}

var (
	DB     *sql.DB
	cache  = make(map[string]MA0Record)
	cacheM sync.Mutex
)

// CheckOrCreateMA0 はキャッシュ→DB→マスター照会→INSERT の順で動作します。
func CheckOrCreateMA0(jan string) (MA0Record, bool, error) {
	cacheM.Lock()
	defer cacheM.Unlock()

	log.Printf("[ma0] ▶ CheckOrCreateMA0 start: JAN=%s", jan)

	// キャッシュチェック
	if rec, ok := cache[jan]; ok {
		log.Printf("[ma0] ⇨ cache hit: %+v", rec)
		logCount()
		return rec, false, nil
	}

	// 永続テーブル検索
	var rec MA0Record
	err := DB.QueryRow(`
        SELECT 
MA000JC000JanCode,
MA001JC001JanCodeShikibetsuKubun,
MA002JC002KyuuJanCode,
MA003JC003TouitsuShouhinCode,
MA004JC004YakkaKijunShuusaiIyakuhinCode,
MA005JC005KyuuYakkaKijunShuusaiIyakuhinCode,
MA006JC006HOTBangou,
MA007JC007ReseputoCode1,
MA008JC008ReseputoCode2,
MA009JC009YJCode,
MA010JC010YakkouBunruiCode,
MA011JC011YakkouBunruiMei,
MA012JC012ShiyouKubunCode,
MA013JC013ShiyouKubunMeishou,
MA014JC014NihonHyoujunShouhinBunruiBangou,
MA015JC015ZaikeiCode,
MA016JC016ZaikeiKigou,
MA017JC017ZaikeiMeishou,
MA018JC018ShouhinMei,
MA019JC019HankakuShouhinMei,
MA020JC020KikakuYouryou,
MA021JC021HankakuKikakuYouryou,
MA022JC022ShouhinMeiKanaSortYou,
MA023JC023ShouhinMeiKanpouYouKigou,
MA024JC024IppanMeishou,
MA025JC025YakkaShuusaiMeishou,
MA026JC026ReseYouIyakuhinMei,
MA027JC027KikakuTaniMeishou,
MA028JC028KikakuTaniKigou,
MA029JC029HanbaiMotoCode,
MA030JC030HanbaiMotoMei,
MA031JC031HanbaiMotoMeiKana,
MA032JC032HanbaiMotoMeiRyakuMei,
MA033JC033SeizouMotoYunyuuMotoCode,
MA034JC034SeizouMotoYunyuuMotoMei,
MA035JC035SeizouMotoYunyuuMotoMeiKana,
MA036JC036SeizouMotoYunyuuMotoMeiRyakuMei,
MA037JC037HousouKeitai,
MA038JC038HousouTaniSuuchi,
MA039JC039HousouTaniTani,
MA040JC040HousouSuuryouSuuchi,
MA041JC041HousouSuuryouTani,
MA042JC042HousouIrisuuSuuchi,
MA043JC043HousouIrisuuTani,
MA044JC044HousouSouryouSuuchi,
MA045JC045HousouSouryouTani,
MA046JC046HousouYouryouSuuchi,
MA047JC047HousouYouryouTani,
MA048JC048HousouYakkaKeisuu,
MA049JC049GenTaniYakka,
MA050JC050GenHousouYakka,
MA051JC051KyuuTaniYakka,
MA052JC052KyuuHousouYakka,
MA053JC053KokuchiTaniYakka,
MA054JC054KokuchiHousouYakka,
MA055JC055YakkaKaiteiNengappi,
MA056JC056YakkaShuusaiNengappi,
MA057JC057HanbaiKaishiNengappi,
MA058JC058KeikaSochiNengappi,
MA059JC059HatsubaiChuushiNengappi,
MA060JC060SeizouChuushiNengappi,
MA061JC061Doyaku,
MA062JC062Gekiyaku,
MA063JC063Mayaku,
MA064JC064Kouseishinyaku,
MA065JC065Kakuseizai,
MA066JC066KakuseizaiGenryou,
MA067JC067ShuukanseiIyakuhin,
MA068JC068ShiteiIyakuhinKyuuKiseiKubun,
MA069JC069YoushijiIyakuhinKyuuKiseiKubun,
MA070JC070KetsuekiSeizai,
MA071JC071NihonYakkyokuhou,
MA072JC072YuukouKikan,
MA073JC073ShiyouKigen,
MA074JC074SeibutsuYuraiSeihin,
MA075JC075Kouhatsuhin,
MA076JC076YakkaKijunShuusaiKubun,
MA077JC077KichouGimuKubun,
MA078JC078ShouhinKubun,
MA079JC079ShohousenIyakuhin,
MA080JC080ChuushiRiyuuKubun,
MA081JC081MishiyouKyuuRyuutsuuKanrihin,
MA082JC082MentenanceKubun,
MA083JC083KouhatsuhinNoAruSenpatsuhinKubun,
MA084JC084AuthorizedGeneric,
MA085JC085Biosimilar,
MA086JC086HighRiskYaku,
MA087JC087Kuuran1,
MA088JC088Kuuran2,
MA089JC089Shitsuon,
MA090JC090Reisho,
MA091JC091Reizou,
MA092JC092Reitou,
MA093JC093Ansho,
MA094JC094Shakou,
MA095JC095KimitsuYouki,
MA096JC096MippuuYouki,
MA097JC097Kikenbutsu,
MA098JC098OndoJougen,
MA099JC099OndoKagen,
MA100JC100SonotaHokanjouNoChui,
MA101JC101KonpouJuuryouSizeJouhou,
MA102JC102KonpouTateSizeJouhou,
MA103JC103KonpouYokoSizeJouhou,
MA104JC104KonpouTakasaSizeJouhou,
MA105JC105KonpouTaiseiSizeJouhou,
MA106JC106ChuubakoJuuryouSizeJouhou,
MA107JC107ChuubakoTateSizeJouhou,
MA108JC108ChuubakoYokoSizeJouhou,
MA109JC109ChuubakoTakasaSizeJouhou,
MA110JC110ChuubakoTaiseiSizeJouhou,
MA111JC111KousouJuuryouSizeJouhou,
MA112JC112KousouTateSizeJouhou,
MA113JC113KousouYokoSizeJouhou,
MA114JC114KousouTakasaSizeJouhou,
MA115JC115KousouTaiseiSizeJouhou,
MA116JC116KonpouTaniSizeJouhou,
MA117JC117HacchuuTaniSizeJouhou,
MA118JC118KoushinKubun,
MA119JC119TourokuNengappi,
MA120JC120KoushinNengappi,
MA121JC121ChouzaiHousouTaniCode,
MA122JC122HanbaiHousouTaniCode,
MA123JC123IppanMeiKana,
MA124JC124SaishouYakkaKansanKeisuu,
MA125JA000,
MA126JA001JanCode,
MA127JA002,
MA128JA003,
MA129JA004,
MA130JA005,
MA131JA006HousouSuuryouSuuchi,
MA132JA007HousouSuuryouTaniCode,
MA133JA008HousouSouryouSuuchi,
MA134JA009,
MA135JA010,
MA136JA011,
MA137JA012,
MA138JA013,
MA139JA014,
MA140JA015,
MA141JA016,
MA142JA017,
MA143JA018,
MA144JA019,
MA145JA020,
MA146JA021,
MA147JA022,
MA148JA023,
MA149JA024,
MA150JA025,
MA151JA026,
MA152JA027,
MA153JA028,
MA154JA029
          FROM ma0
         WHERE MA000JC000JanCode = ?
    `, jan).Scan(
		&rec.MA000JC000JanCode,
		&rec.MA001JC001JanCodeShikibetsuKubun,
		&rec.MA002JC002KyuuJanCode,
		&rec.MA003JC003TouitsuShouhinCode,
		&rec.MA004JC004YakkaKijunShuusaiIyakuhinCode,
		&rec.MA005JC005KyuuYakkaKijunShuusaiIyakuhinCode,
		&rec.MA006JC006HOTBangou,
		&rec.MA007JC007ReseputoCode1,
		&rec.MA008JC008ReseputoCode2,
		&rec.MA009JC009YJCode,
		&rec.MA010JC010YakkouBunruiCode,
		&rec.MA011JC011YakkouBunruiMei,
		&rec.MA012JC012ShiyouKubunCode,
		&rec.MA013JC013ShiyouKubunMeishou,
		&rec.MA014JC014NihonHyoujunShouhinBunruiBangou,
		&rec.MA015JC015ZaikeiCode,
		&rec.MA016JC016ZaikeiKigou,
		&rec.MA017JC017ZaikeiMeishou,
		&rec.MA018JC018ShouhinMei,
		&rec.MA019JC019HankakuShouhinMei,
		&rec.MA020JC020KikakuYouryou,
		&rec.MA021JC021HankakuKikakuYouryou,
		&rec.MA022JC022ShouhinMeiKanaSortYou,
		&rec.MA023JC023ShouhinMeiKanpouYouKigou,
		&rec.MA024JC024IppanMeishou,
		&rec.MA025JC025YakkaShuusaiMeishou,
		&rec.MA026JC026ReseYouIyakuhinMei,
		&rec.MA027JC027KikakuTaniMeishou,
		&rec.MA028JC028KikakuTaniKigou,
		&rec.MA029JC029HanbaiMotoCode,
		&rec.MA030JC030HanbaiMotoMei,
		&rec.MA031JC031HanbaiMotoMeiKana,
		&rec.MA032JC032HanbaiMotoMeiRyakuMei,
		&rec.MA033JC033SeizouMotoYunyuuMotoCode,
		&rec.MA034JC034SeizouMotoYunyuuMotoMei,
		&rec.MA035JC035SeizouMotoYunyuuMotoMeiKana,
		&rec.MA036JC036SeizouMotoYunyuuMotoMeiRyakuMei,
		&rec.MA037JC037HousouKeitai,
		&rec.MA038JC038HousouTaniSuuchi,
		&rec.MA039JC039HousouTaniTani,
		&rec.MA040JC040HousouSuuryouSuuchi,
		&rec.MA041JC041HousouSuuryouTani,
		&rec.MA042JC042HousouIrisuuSuuchi,
		&rec.MA043JC043HousouIrisuuTani,
		&rec.MA044JC044HousouSouryouSuuchi,
		&rec.MA045JC045HousouSouryouTani,
		&rec.MA046JC046HousouYouryouSuuchi,
		&rec.MA047JC047HousouYouryouTani,
		&rec.MA048JC048HousouYakkaKeisuu,
		&rec.MA049JC049GenTaniYakka,
		&rec.MA050JC050GenHousouYakka,
		&rec.MA051JC051KyuuTaniYakka,
		&rec.MA052JC052KyuuHousouYakka,
		&rec.MA053JC053KokuchiTaniYakka,
		&rec.MA054JC054KokuchiHousouYakka,
		&rec.MA055JC055YakkaKaiteiNengappi,
		&rec.MA056JC056YakkaShuusaiNengappi,
		&rec.MA057JC057HanbaiKaishiNengappi,
		&rec.MA058JC058KeikaSochiNengappi,
		&rec.MA059JC059HatsubaiChuushiNengappi,
		&rec.MA060JC060SeizouChuushiNengappi,
		&rec.MA061JC061Doyaku,
		&rec.MA062JC062Gekiyaku,
		&rec.MA063JC063Mayaku,
		&rec.MA064JC064Kouseishinyaku,
		&rec.MA065JC065Kakuseizai,
		&rec.MA066JC066KakuseizaiGenryou,
		&rec.MA067JC067ShuukanseiIyakuhin,
		&rec.MA068JC068ShiteiIyakuhinKyuuKiseiKubun,
		&rec.MA069JC069YoushijiIyakuhinKyuuKiseiKubun,
		&rec.MA070JC070KetsuekiSeizai,
		&rec.MA071JC071NihonYakkyokuhou,
		&rec.MA072JC072YuukouKikan,
		&rec.MA073JC073ShiyouKigen,
		&rec.MA074JC074SeibutsuYuraiSeihin,
		&rec.MA075JC075Kouhatsuhin,
		&rec.MA076JC076YakkaKijunShuusaiKubun,
		&rec.MA077JC077KichouGimuKubun,
		&rec.MA078JC078ShouhinKubun,
		&rec.MA079JC079ShohousenIyakuhin,
		&rec.MA080JC080ChuushiRiyuuKubun,
		&rec.MA081JC081MishiyouKyuuRyuutsuuKanrihin,
		&rec.MA082JC082MentenanceKubun,
		&rec.MA083JC083KouhatsuhinNoAruSenpatsuhinKubun,
		&rec.MA084JC084AuthorizedGeneric,
		&rec.MA085JC085Biosimilar,
		&rec.MA086JC086HighRiskYaku,
		&rec.MA087JC087Kuuran1,
		&rec.MA088JC088Kuuran2,
		&rec.MA089JC089Shitsuon,
		&rec.MA090JC090Reisho,
		&rec.MA091JC091Reizou,
		&rec.MA092JC092Reitou,
		&rec.MA093JC093Ansho,
		&rec.MA094JC094Shakou,
		&rec.MA095JC095KimitsuYouki,
		&rec.MA096JC096MippuuYouki,
		&rec.MA097JC097Kikenbutsu,
		&rec.MA098JC098OndoJougen,
		&rec.MA099JC099OndoKagen,
		&rec.MA100JC100SonotaHokanjouNoChui,
		&rec.MA101JC101KonpouJuuryouSizeJouhou,
		&rec.MA102JC102KonpouTateSizeJouhou,
		&rec.MA103JC103KonpouYokoSizeJouhou,
		&rec.MA104JC104KonpouTakasaSizeJouhou,
		&rec.MA105JC105KonpouTaiseiSizeJouhou,
		&rec.MA106JC106ChuubakoJuuryouSizeJouhou,
		&rec.MA107JC107ChuubakoTateSizeJouhou,
		&rec.MA108JC108ChuubakoYokoSizeJouhou,
		&rec.MA109JC109ChuubakoTakasaSizeJouhou,
		&rec.MA110JC110ChuubakoTaiseiSizeJouhou,
		&rec.MA111JC111KousouJuuryouSizeJouhou,
		&rec.MA112JC112KousouTateSizeJouhou,
		&rec.MA113JC113KousouYokoSizeJouhou,
		&rec.MA114JC114KousouTakasaSizeJouhou,
		&rec.MA115JC115KousouTaiseiSizeJouhou,
		&rec.MA116JC116KonpouTaniSizeJouhou,
		&rec.MA117JC117HacchuuTaniSizeJouhou,
		&rec.MA118JC118KoushinKubun,
		&rec.MA119JC119TourokuNengappi,
		&rec.MA120JC120KoushinNengappi,
		&rec.MA121JC121ChouzaiHousouTaniCode,
		&rec.MA122JC122HanbaiHousouTaniCode,
		&rec.MA123JC123IppanMeiKana,
		&rec.MA124JC124SaishouYakkaKansanKeisuu,
		&rec.MA125JA000,
		&rec.MA126JA001JanCode,
		&rec.MA127JA002,
		&rec.MA128JA003,
		&rec.MA129JA004,
		&rec.MA130JA005,
		&rec.MA131JA006HousouSuuryouSuuchi,
		&rec.MA132JA007HousouSuuryouTaniCode,
		&rec.MA133JA008HousouSouryouSuuchi,
		&rec.MA134JA009,
		&rec.MA135JA010,
		&rec.MA136JA011,
		&rec.MA137JA012,
		&rec.MA138JA013,
		&rec.MA139JA014,
		&rec.MA140JA015,
		&rec.MA141JA016,
		&rec.MA142JA017,
		&rec.MA143JA018,
		&rec.MA144JA019,
		&rec.MA145JA020,
		&rec.MA146JA021,
		&rec.MA147JA022,
		&rec.MA148JA023,
		&rec.MA149JA024,
		&rec.MA150JA025,
		&rec.MA151JA026,
		&rec.MA152JA027,
		&rec.MA153JA028,
		&rec.MA154JA029,
	)
	if err == nil {
		log.Printf("[ma0] ⇨ found in DB: %+v", rec)
		cache[jan] = rec
		logCount()
		return rec, false, nil
	}
	if err != sql.ErrNoRows {
		return MA0Record{}, false, fmt.Errorf("DB query error: %v", err)
	}
	log.Printf("[ma0] ⇨ not found in ma0, querying masters")

	// JCSHMSマスター照会
	csRecs, err := jcshms.QueryJCSHMSRecordsByJan(DB, jan)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("jcshms query error: %v", err)
	}
	// JANCODEマスター照会
	jaRecs, err := jancode.QueryJANCODERecordsByJan(DB, jan)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("jancode query error: %v", err)
	}
	log.Printf("[ma0] ⇨ master counts: JCSHMS=%d rows, JANCODE=%d rows", len(csRecs), len(jaRecs))

	// 値の組み立て
	var yj, t string
	if len(csRecs) > 0 {
		yj = csRecs[0].JC.JC009YJCode
	}
	if len(jaRecs) > 1 {
		t = jaRecs[1].JA006HousouSuuryouSuuchi
	} else if len(jaRecs) > 0 {
		t = jaRecs[0].JA006HousouSuuryouSuuchi
	}
	log.Printf("[ma0] ⇨ about to insert: jan=%q, yj=%q, t=%q", jan, yj, t)

	// INSERT
	res, err := DB.Exec(`
        INSERT INTO ma0 (
            MA000JC000JanCode,
            MA009JC009YJCode,
            MA131JA006HousouSuuryouSuuchi
        ) VALUES (?, ?, ?)
    `, jan, yj, t)
	if err != nil {
		return MA0Record{}, false, fmt.Errorf("insert error: %v", err)
	}
	cnt, _ := res.RowsAffected()
	log.Printf("[ma0] ⇨ inserted rows: %d", cnt)

	newRec := MA0Record{
		MA000JC000JanCode:             jan,
		MA009JC009YJCode:              yj,
		MA131JA006HousouSuuryouSuuchi: t,
	}

	cache[jan] = newRec
	logCount()
	return newRec, true, nil
}

// ProcessMA0Record は DAT レコードごとに呼ばれます。
func ProcessMA0Record(data []string) error {
	if len(data) < 3 {
		return fmt.Errorf("insufficient DAT data: %v", data)
	}
	rec, created, err := CheckOrCreateMA0(data[2])
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("New MA0 created: %+v\n", rec)
	}
	return nil
}

// ViewMA0Handler はキャッシュの内容を JSON で返却します。
func ViewMA0Handler(w http.ResponseWriter, r *http.Request) {
	cacheM.Lock()
	defer cacheM.Unlock()

	list := make([]MA0Record, 0, len(cache))
	for _, rec := range cache {
		list = append(list, rec)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(list)
}

// CountMA0 は ma0 テーブルの件数を返します。
func CountMA0() (int, error) {
	var cnt int
	err := DB.QueryRow("SELECT COUNT(*) FROM ma0").Scan(&cnt)
	return cnt, err
}

func logCount() {
	cnt, err := CountMA0()
	if err != nil {
		log.Printf("[ma0] count error: %v", err)
		return
	}
	log.Printf("[ma0] current ma0 count: %d", cnt)
}
