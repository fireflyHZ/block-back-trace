package lotus

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/reward"
	"profit-allocation/models"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var setupLog = logging.Logger("lotus-setup")

func Setup() {
	reward.CreateLotusClient()
	collectTime := time.NewTicker(time.Second * time.Duration(30))

	defer collectTime.Stop()

	//完成数据初始化
	initTmpData()
	for {
		select {
		case <-collectTime.C:
			loop()
		}

	}
}

func loop() {
	sync.Wg.Add(2)
	go reward.CalculateMsgGasData()
	go reward.CollectTotalRerwardAndPledge()
	sync.Wg.Wait()
}

func initTmpData() {
	o := orm.NewOrm()
	minerInfo := make([]models.MinerInfo, 0)
	n, err := o.QueryTable("fly_miner_info").All(&minerInfo)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	pleagef02420, err := strconv.ParseFloat("63628.11301812885", 64)
	pleagef021695, err := strconv.ParseFloat("1751.8603603064178", 64)
	pleagef021704, err := strconv.ParseFloat("2070.172072901768", 64)
	pleagef044315, err := strconv.ParseFloat("2881.346726219449", 64)
	pleagef055446, err := strconv.ParseFloat("2009.8058479382062", 64)
	pleagef088290, err := strconv.ParseFloat("1452.7006079800967", 64)
	pleagef099132, err := strconv.ParseFloat("5702.729052391221", 64)
	pleagef0104398, err := strconv.ParseFloat("8388.527425174609", 64)
	pleagef0117450, err := strconv.ParseFloat("1236.490325142298", 64)
	pleagef0122533, err := strconv.ParseFloat("3940.5199383872227", 64)
	pleagef0129422, err := strconv.ParseFloat("0.0", 64)
	pleagef0130686, err := strconv.ParseFloat("4442.871949606818", 64)
	pleagef0144528, err := strconv.ParseFloat("37.651937681223465", 64)
	pleagef0144530, err := strconv.ParseFloat("934.4305342858444", 64)
	pleagef0148452, err := strconv.ParseFloat("0.0", 64)
	pleagef0161819, err := strconv.ParseFloat("0.0", 64)

	if err != nil {
		setupLog.Error("ParseFloat err:%+v", err)
	}
	if n == 0 {
		miner1 := models.MinerInfo{
			MinerId:      "f02420",
			QualityPower: 6124.375,
			Pleage:       pleagef02420,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner2 := models.MinerInfo{
			MinerId:      "f021695",
			QualityPower: 199.03125,
			Pleage:       pleagef021695,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner3 := models.MinerInfo{
			MinerId:      "f021704",
			QualityPower: 312.5625,
			Pleage:       pleagef021704,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner4 := models.MinerInfo{
			MinerId:      "f044315",
			QualityPower: 404.3125,
			Pleage:       pleagef044315,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner5 := models.MinerInfo{
			MinerId:      "f055446",
			QualityPower: 294.46875,
			Pleage:       pleagef055446,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner6 := models.MinerInfo{
			MinerId:      "f088290",
			QualityPower: 178.5,
			Pleage:       pleagef088290,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner7 := models.MinerInfo{
			MinerId:      "f099132",
			QualityPower: 678.96875,
			Pleage:       pleagef099132,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner8 := models.MinerInfo{
			MinerId:      "f0104398",
			QualityPower: 1015.4375,
			Pleage:       pleagef0104398,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner9 := models.MinerInfo{
			MinerId:      "f0117450",
			QualityPower: 142.46875,
			Pleage:       pleagef0117450,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner10 := models.MinerInfo{
			MinerId:      "f0122533",
			QualityPower: 458.25,
			Pleage:       pleagef0122533,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner11 := models.MinerInfo{
			MinerId:      "f0129422",
			QualityPower: 0.0,
			Pleage:       pleagef0129422,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner12 := models.MinerInfo{
			MinerId:      "f0130686",
			QualityPower: 505.0625,
			Pleage:       pleagef0130686,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner13 := models.MinerInfo{
			MinerId:      "f0144528",
			QualityPower: 4.25,
			Pleage:       pleagef0144528,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner14 := models.MinerInfo{
			MinerId:      "f0144530",
			QualityPower: 105.4375,
			Pleage:       pleagef0144530,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner15 := models.MinerInfo{
			MinerId:      "f0148452",
			QualityPower: 0.0,
			Pleage:       pleagef0148452,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		miner16 := models.MinerInfo{
			MinerId:      "f0161819",
			QualityPower: 0.0,
			Pleage:       pleagef0161819,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		minerInfo = append(minerInfo, miner1)
		minerInfo = append(minerInfo, miner2)
		minerInfo = append(minerInfo, miner3)
		minerInfo = append(minerInfo, miner4)
		minerInfo = append(minerInfo, miner5)
		minerInfo = append(minerInfo, miner6)
		minerInfo = append(minerInfo, miner7)
		minerInfo = append(minerInfo, miner8)
		minerInfo = append(minerInfo, miner9)
		minerInfo = append(minerInfo, miner10)
		minerInfo = append(minerInfo, miner11)
		minerInfo = append(minerInfo, miner12)
		minerInfo = append(minerInfo, miner13)
		minerInfo = append(minerInfo, miner14)
		minerInfo = append(minerInfo, miner15)
		minerInfo = append(minerInfo, miner16)

		//minerInfo=append(minerInfo,miner1)
		n, err = o.InsertMulti(16, minerInfo)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}

	minerAndWalletRelations := make([]models.MinerAndWalletRelation, 0)
	n, err = o.QueryTable("fly_miner_and_wallet_relation").All(&minerAndWalletRelations)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n == 0 {
		minerAndWalletRelation1 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3rmhlmqfaph6czwiqwlg3kfjgejugt5thcviowlmt3l42464q25ptk3znphuuiwrdbyumun3ui7q2gut7v2da",
		}
		minerAndWalletRelation2 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3wykhltf7g7guch6sz6u6hs4gdmvxrz2likki7aaf7th23jmofrswbd6rmlhbxx4urt6ycbtmlhitgmynky3a",
		}
		minerAndWalletRelation3 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3vfgq65omcht6hbmlwe2g7mowf334zyoa6zcqm543vtmb3uqpnpei4bwhbmo2qi3qntrfiojhcnpciakea6ma",
		}
		minerAndWalletRelation4 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3va7lv4wkcfq5mmqirr4pyrogtnuknw2hma5y6luwbx6iv4qcwgrvzyn2zljgbgtmv7lxr3jsa4eo2az3kqra",
		}
		minerAndWalletRelation5 := models.MinerAndWalletRelation{
			MinerId:  "f021695",
			WalletId: "f3qqdp53ooe4xvqwt4dmoixb6ej6jgmk7zbkjaiujfmfmuyrpenewqre6tlokcxnwp7zpmq3ohlw2wheqir2ga",
		}
		minerAndWalletRelation6 := models.MinerAndWalletRelation{
			MinerId:  "f021695",
			WalletId: "f3wqijosc44y6a6nckbobrwmq6cocoja3lgrly462z3sjwigyi6pzltourrk4lk4jkt332yr5k4xb6mxmct25a",
		}
		minerAndWalletRelation7 := models.MinerAndWalletRelation{
			MinerId:  "f021695",
			WalletId: "f3rs635d7ujd6g6ofhmkrybvhvvrebv5wap3y7yzui2hymjc2u5s3tgidcr3skoxspbjhfbyvewqyjlybx2ajq",
		}
		minerAndWalletRelation8 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3spvlhfuga45prd7fg7dswphgm4hotpxmydyzpjloy2rekpyfnwpbdnd7wuyael2pryb3xztp4k56ju3ib5sq",
		}
		minerAndWalletRelation9 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3skdqsai23rhavva77g7nkr736j7mjql53xv7362ovlw7o3yz334ajchyb7fir35cnutijfusp6mngobyjvya",
		}
		minerAndWalletRelation10 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3sc6mo6jiwxwwgsx4gwz5vbpcn4p6ejybgogocfntujmjaibluzm6ngj7qqj72gck7rtuibtgsow6ttuq43dq",
		}
		minerAndWalletRelation11 := models.MinerAndWalletRelation{
			MinerId:  "f044315",
			WalletId: "f3ritrdadwyomhkfcqrif7qutqa7e3xurxlgdl6rui6l6mbuibhzrrsqhzoklwxrkxtmsudtwhj5aca4uafreq",
		}
		minerAndWalletRelation12 := models.MinerAndWalletRelation{
			MinerId:  "f044315",
			WalletId: "f3xfk65yvwynfvcfdqnyb6abcsslblu6ju5qh3illjnxy4favaescrseuuydeopgdhiik6ooo4o3pvrk32y6yq",
		}
		minerAndWalletRelation13 := models.MinerAndWalletRelation{
			MinerId:  "f044315",
			WalletId: "f3xgszxngrizxkzo2ou6njw5mrhbfvz6zfhfvwfmbc3km7tth6p6irnwksdqxtcwscp3vjgynj6ijts4dd2gga",
		}
		minerAndWalletRelation14 := models.MinerAndWalletRelation{
			MinerId:  "f055446",
			WalletId: "f3wih6tpcyk6mfg6o55sdgfw5efvbbuzvd4p7mfjl67dwaeoef6wnlaqzejdnisa4wzdajviasyv6ipufrttmq",
		}
		minerAndWalletRelation15 := models.MinerAndWalletRelation{
			MinerId:  "f055446",
			WalletId: "f3rrrwlcgctgosfpfztex2kvbsyxyho3k2wiikxds3vxmpuhiy7odnom25ou4tl3dvq7ta7uoi6blsfx25w73q",
		}
		minerAndWalletRelation16 := models.MinerAndWalletRelation{
			MinerId:  "f055446",
			WalletId: "f3wc24qesjodezvjnt3luoh2xjaej7q3b3iryzxizphsh2ofqleduccqzlyqjhvgmudeyw3hx2hsvw4x4r3qqq",
		}
		minerAndWalletRelation17 := models.MinerAndWalletRelation{
			MinerId:  "f088290",
			WalletId: "f3sapm3ztthk5kucytirqxvou6hfaczz6hulymfrslym3co42u73d3sq6xszhfdubrtrkxxg4liyorsk5deloq",
		}
		minerAndWalletRelation18 := models.MinerAndWalletRelation{
			MinerId:  "f088290",
			WalletId: "f3vif2ni6y27lo34666olsz7qi6i64h2nviacv7z4x75k6ud3g6lvox5k7lz2h7bydsexjvueeismmd43vtq6q",
		}
		minerAndWalletRelation19 := models.MinerAndWalletRelation{
			MinerId:  "f088290",
			WalletId: "f3ssrkr5hzsam23xa3f7y7ck6uvvfrosdmf3zqb4lmwfh4wakwstzwpuvy23fj7z274kd6ywqisltway3oklmq",
		}
		minerAndWalletRelation20 := models.MinerAndWalletRelation{
			MinerId:  "f099132",
			WalletId: "f3vlodr4d3v2btencffsvggqfndhrlepzvd7hrxq6mz2tbspsnc2u7nurp5dtyfnxvxticdxrtdeuwsjps3yna",
		}
		minerAndWalletRelation21 := models.MinerAndWalletRelation{
			MinerId:  "f099132",
			WalletId: "f3qqwprq2lczgwqkse45wxo2oeqfkjipf42nhc6sxlfcmqpxui4a42daclqma4nopagigifvaqrosdmj4vzxpq",
		}
		minerAndWalletRelation22 := models.MinerAndWalletRelation{
			MinerId:  "f099132",
			WalletId: "f3vz453mhi2zwhphkoubxagxkgdukkp4o66rtpkecamcqqd3dgyaowbswttvtqt6iwkpbefsxxws2yho4lqxsa",
		}
		minerAndWalletRelation23 := models.MinerAndWalletRelation{
			MinerId:  "f0104398",
			WalletId: "f3rcerdbxglklcr6hozfrvmlg3e2xf53x35nd3sxqcxk6pahmdrxnz5ebi5tbrswgap3f3hs4ezxnhdk5e6oeq",
		}
		minerAndWalletRelation24 := models.MinerAndWalletRelation{
			MinerId:  "f0104398",
			WalletId: "f3vbqg6ttmmlwj73ng3rxjotzmeeccygbdmnsxpeu2tc6tjomi54pmtbdd6fd5a6efwxanwwxgv4dnxaorhbpa",
		}
		minerAndWalletRelation25 := models.MinerAndWalletRelation{
			MinerId:  "f0104398",
			WalletId: "f3rg5se4ndmh7xhbwxztjifkubnphdgpetej3xf6ob75fzc52scsggoiyntcznmicezbzkamax25tszb2cs7gq",
		}
		minerAndWalletRelation26 := models.MinerAndWalletRelation{
			MinerId:  "f0117450",
			WalletId: "f3ufsnk2uu6naqhe6ssbtjzdsclgpqstbr4gtlofm5uu32vj7sbjmtef35ynxpebm6yutbuwxjafh7jyrzpaqq",
		}
		minerAndWalletRelation27 := models.MinerAndWalletRelation{
			MinerId:  "f0117450",
			WalletId: "f3rrrr2vmhrkdpb4aqyozgnja47ewki3bbl2sv4mfenapjod3volvtk74zjshgi4txehbbw6bkginwxiuthkca",
		}
		minerAndWalletRelation28 := models.MinerAndWalletRelation{
			MinerId:  "f0117450",
			WalletId: "f3xbh3oswkxw6bglnkyljvgktiv2iiqk5zco6ektg2wwyvrtopiym52zoxrxn7cz2p7ye7m2254qwqsjrikfla",
		}
		minerAndWalletRelation29 := models.MinerAndWalletRelation{
			MinerId:  "f0122533",
			WalletId: "f3uvwmuwlaz4qr7i4xlxucg25wgy6vc2afgj5idabyoh6umpj25ugdt3bbmmcg3nprtzq7okhdziljnoa7pj7q",
		}
		minerAndWalletRelation30 := models.MinerAndWalletRelation{
			MinerId:  "f0122533",
			WalletId: "f3q5g4miz74mjmjk5stin344rvhiay7eeei3drdin4kzmnbli4dmsmuirxlyxk6v4luot67fxzqf5vg5gdlxhq",
		}
		minerAndWalletRelation31 := models.MinerAndWalletRelation{
			MinerId:  "f0122533",
			WalletId: "f3s3urn3k2mk3y3utllw327r5fgbmq5xipizmrjpxqjlohv2eklua7pmt3kz3zv6dd73sxe2v4olgloi4vfvqq",
		}
		minerAndWalletRelation32 := models.MinerAndWalletRelation{
			MinerId:  "f0129422",
			WalletId: "f3rcvnwwoesheupiy3fdeus34zba7qouooobl3wlluquid667bi3cpawxje35qyyrahhkgup2wx2e3bmvue7kq",
		}
		minerAndWalletRelation33 := models.MinerAndWalletRelation{
			MinerId:  "f0130686",
			WalletId: "f3wiohczzvnaci3xtrs7d367fpaydbm3ee7roqvqz6vnigvxijb3egep7taktldoywn3prqrxjm5wlzt5pstua",
		}
		minerAndWalletRelation34 := models.MinerAndWalletRelation{
			MinerId:  "f0130686",
			WalletId: "f3ww2i22r2dwjpkfjuflzsxj2nnhkuhxu6hqxw4jlc42dl6fo7wzwo46ullkt6xiql6n7zgv2dkrzy4u7wyecq",
		}
		minerAndWalletRelation35 := models.MinerAndWalletRelation{
			MinerId:  "f0130686",
			WalletId: "f3qirv77wfhf5j3ddxz4mdjrhtkjyaf65e62iod2355r7h3v3hyjj4urnnb4poo7gxt6xkacfkjhs7toi2qfkq",
		}
		minerAndWalletRelation36 := models.MinerAndWalletRelation{
			MinerId:  "f0144528",
			WalletId: "f3sws2i5yuuwu53aoy7rekbslp7oni22yathzw5o446l7jdo4wb3jefmkyyybvnibihegts4mjhqpv5slyo7la",
		}
		minerAndWalletRelation37 := models.MinerAndWalletRelation{
			MinerId:  "f0144528",
			WalletId: "f3qoad6nqnx2tfe7tikceszepraes7ag5wvrtohq2qtr7r77veps7hxjd3nbk4jhcipjz7mz4cytzri5skazna",
		}
		minerAndWalletRelation38 := models.MinerAndWalletRelation{
			MinerId:  "f0144528",
			WalletId: "f3sqr4chsmih6rd2dwzlqkf4b4riqopo3wopoymex2he3vh335jlqs67nwevppv6ykthmtlzz5grpeyjzrursq",
		}
		minerAndWalletRelation39 := models.MinerAndWalletRelation{
			MinerId:  "f0144530",
			WalletId: "f3vpd647fgcuif4lrujkugnej3adwpqotnebek3xoox64a5l5zjpt3ddz2w2pfbl3nkcetl7okz5gakaxblaza",
		}
		minerAndWalletRelation40 := models.MinerAndWalletRelation{
			MinerId:  "f0144530",
			WalletId: "f3w42z5aabtis6svnawas77hmorbz6znmpaolrdfekx5edjxulmeb4snnowxntkptwilayaghdsgcem4chpoya",
		}
		minerAndWalletRelation41 := models.MinerAndWalletRelation{
			MinerId:  "f0144530",
			WalletId: "f3r6thwlcdmuovjh2qzvhkqj4nk2ggsu5edonjkbywldb27lbw4sbzvgcvf4gkqnwh7pb4bln5jrygnw6aaaya",
		}
		minerAndWalletRelation42 := models.MinerAndWalletRelation{
			MinerId:  "f0148452",
			WalletId: "f3vaysv4sxivsb2e4r5tvtgamtk6u4avzcpit4kjbaar675l67dlqpam5f6j5m3uuvamic3rx7g3wsofhvnxfa",
		}
		minerAndWalletRelation43 := models.MinerAndWalletRelation{
			MinerId:  "f0148452",
			WalletId: "f3qob65qjwku2l76w5r23ra3jqhfyml32tcltnc4ygcmfkibkamjol4gzlvni3jyn7m5cuk42nhdv4xvmc7d4q",
		}
		minerAndWalletRelation44 := models.MinerAndWalletRelation{
			MinerId:  "f0148452",
			WalletId: "f3q4kbsf52s7ifr5psmro6i2cea2nh6ojfe62lx3gmwt2mufr4ukffw6xhcqyowzl6thizmcr7brb3kvnop7ca",
		}
		minerAndWalletRelation45 := models.MinerAndWalletRelation{
			MinerId:  "f0161819",
			WalletId: "f3xe7pd2bict4ro5p7cmtlqvqlyqtwicazz326hg5tibyqkohbafboqcbpphd4dedp43k7zalhzj2jjtschufq",
		}
		minerAndWalletRelation46 := models.MinerAndWalletRelation{
			MinerId:  "f0161819",
			WalletId: "f3vi2vnsvmszsgpfbcz6u7uj6eownxzektnw3haq2nwn6sbwbuwlkbd2atlujy2ty3tdgxbkhsgh6xqjezv5qa",
		}
		minerAndWalletRelation47 := models.MinerAndWalletRelation{
			MinerId:  "f0161819",
			WalletId: "f3wrfovianak3onbbbx3ob5iyvqdzqmymv5iax7v5ccdqmpudv2kh4bjt3ir5eukvwyibhremmxozv5edjvuka",
		}
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation1)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation2)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation3)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation4)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation5)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation6)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation7)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation8)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation9)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation10)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation11)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation12)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation13)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation14)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation15)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation16)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation17)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation18)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation19)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation20)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation21)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation22)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation23)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation24)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation25)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation26)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation27)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation28)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation29)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation30)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation31)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation32)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation33)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation34)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation35)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation36)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation37)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation38)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation39)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation40)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation41)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation42)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation43)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation44)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation45)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation46)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation47)
		n, err = o.InsertMulti(47, minerAndWalletRelations)
		if err != nil {
			fmt.Println("insert minerAndWalletRelations err:", err)
		}
	}
}
