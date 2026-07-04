package domain

// Stages は4ステージの定義カタログ。ゲーム進行とプロンプト生成の双方で参照される。
// 内容(成功条件・状況描写)はドメイン知識なのでここに持つ。
var Stages = map[StageID]Stage{
	StageRiver: {
		ID:        StageRiver,
		Title:     "遠くのお城",
		Situation: "とおくにドラゴン城がみえる。おひめさまは そこにいるます。でも、目の前に 川が流れていて むこうにわたれない!",
		SuccessSpec: `プレイヤーが川を渡る・橋をかける・飛び越える・泳ぐ・船を出す・魔法で渡す等、
むこう岸にわたるために有効な行動をとった場合、成功とする。（例: 「はしをかけろ!」「およぐ!」「とびこえる!」等の短い言葉でも、行動として読み取れれば成功）
工夫や勇敢さがあれば Great、やや単調なら Good にする。
川や状況と無関係な発言は Bad。`,
	},
	StageGate: {
		ID:        StageGate,
		Title:     "城門のゴーレム",
		Situation: "ドラゴン城の前に到着した。でも、大きくて怖い「門番のゴーレム」が立ちはだかっている!",
		SuccessSpec: `プレイヤーがゴーレムをどかす・通り抜ける・説得する・なだめる・歌う・冗談で笑わせる・魔法で動かす等、
道を開くために有効な行動をとった場合、成功とする。（例: 「どいて!」「すりぬける!」「なかまになって!」等の短い言葉でも、行動として読み取れれば成功）
叩く・燃やす等でも成功(ただし工夫や優しさがあれば Great、単調だと Good)。
ゴーレムや状況と無関係な発言、曖昧すぎる発言は Bad。`,
	},
	StageFire: {
		ID:        StageFire,
		Title:     "炎のトラップ",
		Situation: "城の中に入ると、通路が激しい炎で包まれていて進めない!",
		SuccessSpec: `プレイヤーが炎を消す・凍らせる・水をかける・盾で防ぐ・飛び越える・風で吹き飛ばす等、
火を無効化または回避する行動をとった場合、成功とする。（例: 「ウォーター!」「こおり!」「とびこえる!」等の短い言葉でも、行動として読み取れれば成功）
工夫や勇敢さがあれば Great、やや単調なら Good にする。
炎と無関係な発言は Bad。`,
	},
	StageDragon: {
		ID:               StageDragon,
		Title:            "ドラゴンとの最終決戦",
		Situation:        "最上階でお姫様を発見!でも、怒ったドラゴンが襲いかかってきた!",
		NeedsDragonRoute: true,
		SuccessSpec: `プレイヤーがドラゴンを倒す(剣・魔法・弓等で打ち負かす)、または友だちになる(話しかける・撫でる・食べ物をあげる・歌う)の
いずれかの行動をとった場合、成功とする。（例: 「たおせ!」「なかまにする!」「たべものをあげる!」等の短い言葉でも、行動として読み取れれば成功）
攻撃系は route=defeat、友好系は route=befriend と判定する。
工夫や勇敢さ・優しさがあれば Great、単調なら Good にする。
ドラゴンと無関係な発言は Bad。`,
	},
}

// StageOrder は進行順。
var StageOrder = []StageID{StageRiver, StageGate, StageFire, StageDragon}

// LookupStage は存在確認付きでステージを取得。
func LookupStage(id StageID) (Stage, bool) {
	s, ok := Stages[id]
	return s, ok
}
