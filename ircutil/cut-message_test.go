package ircutil_test

import (
	"fmt"
	"strings"
	"testing"

	"git.aiterp.net/gisle/irc/ircutil"
)

func TestCuts(t *testing.T) {
	t.Log("Testing that long messages can be cut up and put back together, and that no cut is greater than 510 - overhead")

	table := []struct {
		Overhead int
		Space    bool
		Text     string
	}{
		{
			ircutil.MessageOverhead("Longer_Name", "mircuser", "some-long-hostname-from-some-isp.com", "#Test", true), true,
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed maximus urna eu tincidunt lacinia. Morbi malesuada lacus placerat, ornare tellus a, scelerisque nunc. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam placerat sem aliquet elit pharetra consectetur. Pellentesque ultrices turpis erat, et ullamcorper magna blandit vitae. Morbi aliquam, turpis at dictum hendrerit, mi urna mattis mi, non vulputate ligula sapien non urna. Nulla sed lorem lorem. Proin auctor ante et ligula aliquam lacinia. Sed pretium lacinia varius. Donec urna nibh, aliquam at metus ac, lobortis venenatis sem. Etiam et risus pellentesque diam faucibus faucibus. Vestibulum ornare, erat sit amet dapibus eleifend, arcu erat consectetur enim, id posuere ipsum enim eget metus. Aliquam erat volutpat. Nunc eget neque suscipit nisl fermentum hendrerit. Suspendisse congue turpis non tortor fermentum, vulputate egestas nibh tristique. Sed purus purus, pharetra ac luctus ut, accumsan et enim. Quisque lacus tellus, ullamcorper eu lacus aliquet, facilisis sodales mauris. Quisque fringilla, odio quis laoreet sagittis, urna leo commodo urna, eu auctor arcu arcu ac nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Suspendisse accumsan leo sed sollicitudin dignissim. Aliquam et facilisis turpis. Morbi finibus nisi ut elit eleifend cursus. Donec eu imperdiet nulla. Vestibulum eget varius dui. Morbi dapibus leo sit amet ipsum porta, et volutpat lectus condimentum. Integer nec mi dui. Suspendisse ac tortor et tortor tempus imperdiet. Aenean erat ante, ultricies eget blandit eu, sollicitudin vel nibh. Vestibulum eget dolor urna. Proin sit amet nulla eu urna dictum dignissim. Nulla sit amet velit eu magna feugiat ultricies. Sed venenatis rutrum urna quis malesuada. Curabitur pretium molestie mi eget aliquam. Sed eget est non sem ornare tincidunt. Vestibulum mollis ultricies tellus sit amet fringilla. Vestibulum quam est, blandit venenatis iaculis id, bibendum sit amet purus. Nullam laoreet pellentesque vulputate. Curabitur porttitor massa justo, id pharetra purus ultricies et. Aliquam finibus molestie turpis quis mattis. Nulla pretium mauris dolor, quis porta arcu pulvinar eu. Nam tincidunt ac odio in hendrerit. Pellentesque elementum porttitor dui, at laoreet erat ultrices at. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed porttitor libero magna, vitae malesuada sapien blandit ut. Maecenas tempor auctor tortor eu mollis. Integer tempus mollis euismod. Nunc ligula ligula, dignissim sit amet tempor eget, pharetra lobortis risus. Ut ut libero risus. Integer tempus mauris nec quam volutpat tristique. Maecenas id lacus et metus condimentum placerat. Vestibulum eget mauris eros. Nulla sollicitudin libero id dui imperdiet, at ornare nibh sollicitudin. Pellentesque laoreet mollis nunc aliquam interdum. Phasellus egestas suscipit turpis in laoreet.",
		},
		{
			ircutil.MessageOverhead("=Scene=", "SceneAuthor", "npc.fakeuser.invalid", "#LongChannelName32", false), true,
			"Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum..",
		},
		{
			ircutil.MessageOverhead("=Scene=", "Gissleh", "npc.fakeuser.invalid", "#Channel3", false), true,
			"A really short message that will not be cut.",
		},
		{
			ircutil.MessageOverhead("=Scene=", "Gissleh", "npc.fakeuser.invalid", "#Channel3", false), false,
			"123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		},
		{
			ircutil.MessageOverhead("=Scene=", "Gissleh", "npc.fakeuser.invalid", "#Channel3", false), false,
			// It's just japanese lorem ipsun just to see that multi-byte runes don't get cut wrong.
			"弟ノネセ設35程カメ分軽談オイヲ趣英シ破与預ニ細試かゅ給桐やんぱ新交エムイ招地なる稿訓誘入倒がぞょあ。9未わの表画フ標係暖りかす権邦ざらびフ第木庭げ司第すに芸型ず内兼んほでな答将携きあ大念じぽろ状表そむぞ。販ぜひゆべ万53火6飾び付界ー兵供コ援仕シチヲ決表ウユアエ生記続ヌ金見貢ユ相帯ソ問禁ンる盟策れとぞの難屋セアヌネ記由僚ど物2理シホヤニ数重くむも。米ヱ給辞めつへ長順チ稿転地ヌラヒエ稿横イ検地えつずた質信ヘヨ電本タウ測30終真いず章年雪しあすそ済座づずら期出必ツエキハ色罪ょイえ再危ケトリヘ東宅ちんり京倒停塚称ひほぱ。報ワモネ意取やま画支完ユヱホ一崎そく子聞ハシ始家文なリさ新誕えぱつ渡講ニ著今就せド観3判成トフケユ食別ゆ績生タ告候ざンちぐ芸部幸績養はそを。評ぶラ路2十ぐるッー本米ぼ新性うひゆ詠持ほかな測委もめべに犯九だへほう金作かドえは民53棋倉1無タウコフ真読おだぼ重訃壮憶軒研てだざ。火皇テミヘユ関評レクな記本ラ日設識こへぎ読認水リるっ定件ラリレロ裁写フ記気やい縦写ヤコロ糸取ニワ金朝ウルオ世康でてめ氷諭ソフ副際ロワ念促縮繰やだつせ。重身ケ容6契竹せぴま法能れ改長ひ出葉チソユ得4帯ツキヤホ込養フハケス言杯ネ策振オセメヘ合億育閥班綸諮らせ。算甲ミカ夕支フ疲水ナ度先稿テ定特ぴ問触べ陸月は販93作意ぱへ以分げらご算路亡とスひ。歓にレ完指リ覧論ぱょ中審ロ期旨ヲメヘケ記言ク構早べ埼組党高雄ぽ館世ウ通画ケ裁督え隆学びいゅ交利ヤトワ宮81明よ員乞伍ゅず。更ナヲ士座ヘモレホ意有イリル表半ほラ採政イテ判募相37一対て配小ウオ広更モケヘヲ山週ト難覚ホセク小届角み。3読コロ返立クあまゆ探気休るけをこ安金る展無にりひ聞説ね我郎みゆめ左州フメチヘ気席ん見夏くフには的8実ヒカ表更フ世教聞ロハヒウ引康ばぽわク見測希動五トしげ。囲ニ通2政済キ少罪づあふ止政せげイ四内ラ劇中題ホ感負んけーゅ際禁オテ等鳥通県的せ。議19賠ヒヌニ止牛気ぞぴわょ出来をるおぴ覧5法ヒヱウ金断舞つ発都芸トユ買将カテ需覧ほごレラ必鈴ト部部エムモ無学りぎ掲死お。化もと康集ひ頼禁モツサ覧能べばゆぐ工9奪御セテキリ時者ゆちな美録江レユチウ誠遺目モヲ更新ふあぽら読時り問特リナケク子活マネオリ彌個べざを理時ずルゃ身払縄びひ。学粉捜ワニヲ遂題ル読8野ネリユ世検顔るごかラ作類べ並90弟意リルづゆ証利ミ止9年フ細協づつ。件ドつっお友載モヤヌ占教オ国射ホク部措昨ょげ初勝鋭ヒテワハ女実ゅおねて意情ろく性市へちイぱ務哲れんてか暮両ゆごこ今節ライげ。向タ歩56崎っフゆ庭教ぞめ舞吉タ作道マノ報康エモチニ決欲トヨ棋郵産サレ挙写ル胸覚エネ心耀るざラぱ陽高ネメテ以査ず盾際ハセヌツ領証ルミヌ無不レセエナ同可析ドずと。降月ず自八ヨロナ避詐ドか月買ばこ姿徴ぱ遺6低2紆囲へぽもに権場ひおもわ芋首とごも然得まや点人三ぶ改在ネノヘ時式頑威敵はつく。在う協完ふリ大殺をり容賃エ更50事ワマ木再クびド康決転場コリ上初ミイヱ山第ま費禁トぼをぐ童載私海陸ね。佐なぶゃ早五税イミホ話秋情ト発窃ね替究エコ郎著心ホ編今セシルウ金4本サヨ設中学ざど容迷もそ記主サツカ都枝ぐ哲速ご踊大にど。短マラフチ理玉めフど展掲ばょい皿4題ゅっスぼ性五形53討ごぜン満給ユナツ人不ぼずス全読ゅやみろ赤95俣妃巳ス。原安イ球竹ッょごせ向審ラルサ波野ゃ約球やむリぎ神情21就オサチ覧見式で所雑ンだ延芝ネ推事護技スルセヤ働6典伯ごえ。豊レカリ河楽ラアカイ論投写ー生上ひされ整表ハス断見ヲヱホヤ光五ず申獄さに被度ラ動量ねぐっ席2権方求ぎイきっ因表イょ重稿えどク崎学ほスク。数ヘイ見近ぞが選信ねトに新期ワホ闘府要フぶ立空クしよほ久素アケナモ朝37視ワチ朝93送て民目ヨホラク載径猶くイげ。",
		},
	}

	sep := map[bool]string{false: "", true: " "}

	for i, row := range table {
		t.Run(fmt.Sprintf("Row_%d", i), func(t *testing.T) {
			cuts := ircutil.CutMessage(row.Text, row.Overhead)
			joined := strings.Join(cuts, sep[row.Space])

			for i, cut := range cuts {
				t.Logf("Length %d: %d", i, len(cut))
				t.Logf("Cut    %d: %s", i, cut)

				if len(cut) > (510 - row.Overhead) {
					t.Error("Cut was too long")
				}
			}

			if joined != row.Text {
				t.Error("Cut failed:")
				t.Error("  Result:", joined)
				t.Error("  Expected:", row.Text)
			}
		})
	}
}
