// ma0/mao.go
package ma0

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"YAMATO/jancode"
	"YAMATO/jcshms"
	"YAMATO/model"
)

// MA0Record は、マスター連携用の全155フィールドを保持する構造体です。
// ※フィールド名は "MAxxxJCyyy"（JC マスター連携用）と "MAxxxJAyyy"（JANコード連携用）に分かれています。
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

// DB は、ma0 連携用に参照するグローバルなデータベース接続です。
var DB *sql.DB

// Migrate は、MA0Record の全フィールドを TEXT 型として、
// 最初のフィールドを PRIMARY KEY としたテーブル "ma0" を作成します。
func Migrate(db *sql.DB) error {
	t := reflect.TypeOf(MA0Record{})
	cols := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if i == 0 {
			cols[i] = fmt.Sprintf("%s TEXT PRIMARY KEY", name)
		} else {
			cols[i] = fmt.Sprintf("%s TEXT", name)
		}
	}
	ddl := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS ma0 (\n  %s\n);",
		strings.Join(cols, ",\n  "),
	)
	_, err := db.Exec(ddl)
	return err
}

// columns は MA0Record の各フィールド名をスライスとして返します。
func columns() []string {
	t := reflect.TypeOf(MA0Record{})
	out := make([]string, t.NumField())
	for i := range out {
		out[i] = t.Field(i).Name
	}
	return out
}

// values は与えられた MA0Record のフィールド値の一覧を []interface{} として返します。
func values(rec MA0Record) []interface{} {
	v := reflect.ValueOf(rec)
	out := make([]interface{}, v.NumField())
	for i := range out {
		out[i] = v.Field(i).Interface()
	}
	return out
}

// InsertIgnore は、複数の MA0Record を一括で INSERT OR IGNORE します。
// PRIMARY KEY 制約により重複が自動的に防がれます。
func InsertIgnore(db *sql.DB, recs []MA0Record) error {
	cols := columns()
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	stmt := fmt.Sprintf(
		"INSERT OR IGNORE INTO ma0 (%s) VALUES (%s)",
		strings.Join(cols, ","),
		strings.Join(placeholders, ","),
	)
	prep, err := db.Prepare(stmt)
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, rec := range recs {
		if _, err := prep.Exec(values(rec)...); err != nil {
			return err
		}
	}
	return nil
}

// CheckOrCreateMA0 は、指定された JAN コードで ma0 テーブルを検索します。
// 既存ならそのレコードを返し、created=false とします。
// 見つからなければ、jcshms および jancode からマスター照会を行い、
// 新規レコードを INSERT OR IGNORE して created=true として返します。
func CheckOrCreateMA0(jan string) (MA0Record, bool, error) {
	// 1) ma0 に既にレコードが存在するかチェック
	var rec MA0Record
	cols := columns()
	addrs := make([]interface{}, len(cols))
	recVal := reflect.ValueOf(&rec).Elem()
	for i := range addrs {
		addrs[i] = recVal.Field(i).Addr().Interface()
	}
	query := fmt.Sprintf("SELECT %s FROM ma0 WHERE MA000JC000JanCode = ?", strings.Join(cols, ","))
	err := DB.QueryRow(query, jan).Scan(addrs...)
	if err == nil {
		// 既存レコードが見つかった場合
		return rec, false, nil
	}
	if err != sql.ErrNoRows {
		return MA0Record{}, false, fmt.Errorf("ma0 select error: %v", err)
	}

	// 2) マスター照会（jcshms および jancode から）およびフィールドのコピー
	cs, _ := jcshms.QueryByJan(DB, jan)
	ja, _ := jancode.QueryByJan(DB, jan)

	// 両方のマスターにヒットがなければ、登録せずに終了する
	if len(cs) == 0 && len(ja) == 0 {
		return MA0Record{}, false, nil
	}

	// 反射を用いて、jcshms からの項目を MA0Record にコピー
	if len(cs) > 0 {
		jcVal := reflect.ValueOf(cs[0])
		for i := 0; i < recVal.NumField(); i++ {
			field := recVal.Type().Field(i)
			// MAレコードで "JC" を含むフィールドは、jcshms の対応フィールドへマッピング
			if strings.HasPrefix(field.Name, "MA") && strings.Contains(field.Name, "JC") {
				idx := strings.Index(field.Name, "JC")
				masterName := field.Name[idx:]
				if masterField := jcVal.FieldByName(masterName); masterField.IsValid() {
					recVal.Field(i).SetString(masterField.String())
				}
			}
		}
	}

	// jancode からも同様にコピー（フィールド名に "JA" を含むもの）
	if len(ja) > 0 {
		jaVal := reflect.ValueOf(ja[0])
		for i := 0; i < recVal.NumField(); i++ {
			field := recVal.Type().Field(i)
			if strings.HasPrefix(field.Name, "MA") && strings.Contains(field.Name, "JA") {
				idx := strings.Index(field.Name, "JA")
				masterName := field.Name[idx:]
				if masterField := jaVal.FieldByName(masterName); masterField.IsValid() {
					recVal.Field(i).SetString(masterField.String())
				}
			}
		}
	}

	// 主キー（JANコード）の設定
	rec.MA000JC000JanCode = jan

	// 3) INSERT OR IGNORE により DB へ新規レコード挿入
	if err := InsertIgnore(DB, []MA0Record{rec}); err != nil {
		return MA0Record{}, false, fmt.Errorf("ma0 insert error: %v", err)
	}
	return rec, true, nil
}

// ProcessMA0Record は、dat.go や usage.go から呼び出され、
// 与えられたデータスライスから JAN コードを抽出して CheckOrCreateMA0 を実行します。
// ※ data の 3 番目の要素（インデックス2）が JAN コードであるという慣例に従います。
func ProcessMA0Record(data []string) error {
	if len(data) < 3 {
		return fmt.Errorf("insufficient fields: %v", data)
	}
	jan := data[2]
	_, _, err := CheckOrCreateMA0(jan)
	return err
}

// InsertDATRecord は、与えられた model.DATRecord を datrecords テーブルに挿入します。
// organizedFlag には、1 (organized) または 0 (disorganized) を指定します。
func InsertDATRecord(db *sql.DB, rec model.DATRecord, organizedFlag int) error {
	stmt := `
		INSERT OR IGNORE INTO datrecords (
			CurrentOroshiCode,
			DatDate,
			DatDeliveryFlag,
			DatReceiptNumber,
			DatLineNumber,
			DatJanCode,
			DatProductName,
			DatQuantity,
			DatUnitPrice,
			DatSubtotal,
			DatPackagingDrugPrice,
			DatExpiryDate,
			DatLotNumber,
			organizedFlag
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(stmt,
		rec.CurrentOroshiCode, // DatOroshiCode 列へ
		rec.DatDate,           // DatDate 列へ
		rec.DatFlag,           // DatDeliveryFlag 列へ（旧：DatDeliveryFlag → DatFlag）
		rec.DatRecNo,          // DatReceiptNumber 列へ
		rec.DatLineNo,         // DatLineNumber 列へ
		rec.DatJan,            // DatJanCode 列へ
		rec.DatProductName,    // DatProductName 列へ
		rec.DatQty,            // DatQuantity 列へ
		rec.DatUnit,           // DatUnitPrice 列へ
		rec.DatSub,            // DatSubtotal 列へ
		rec.DatPkg,            // DatPackagingDrugPrice 列へ
		rec.DatExp,            // DatExpiryDate 列へ
		rec.DatLot,            // DatLotNumber 列へ
		organizedFlag,
	)
	if err != nil {
		return fmt.Errorf("failed to insert DATRecord: %w", err)
	}
	return nil
}

// InsertUsageRecord inserts one USAGERecord into the "usage_records" table.
func InsertUsageRecord(db *sql.DB, rec model.USAGERecord) error {
	stmt := `
		INSERT OR IGNORE INTO usage_records (
			usageDate,
			usageYjCode,
			usageJanCode,
			usageProductName,
			usageAmount,
			usageUnit,
			usageUnitName,
			organizedFlag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := db.Exec(stmt,
		rec.UsageDate,
		rec.UsageYjCode,
		rec.UsageJanCode,
		rec.UsageProductName,
		rec.UsageAmount,
		rec.UsageUnit,
		rec.UsageUnitName,
		rec.OrganizedFlag,
	)
	if err != nil {
		return fmt.Errorf("failed to insert USAGE record: %w", err)
	}
	return nil
}
