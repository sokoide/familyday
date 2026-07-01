package domain

// Stages は3ステージの定義カタログ。ゲーム進行とプロンプト生成の双方で参照される。
// 内容(成功条件・状況描写)はドメイン知識なのでここに持つ。
var Stages = map[StageID]Stage{
	StageGate: {
		ID:        StageGate,
		Title:     "城門のゴーレム",
		Situation: "ドラゴン城の前に到着した。でも、大きくて怖い「門番のゴーレム」が立ちはだかっている!",
		Goal:      "ゴーレムをどかす、または通り抜ける呪文を唱えよう!",
		SuccessSpec: `プレイヤーがゴーレムをどかす・通り抜ける・説得する・なだめる・歌う・冗談で笑わせる・魔法で動かす等、
道を開くために有効な行動をとった場合、成功とする。
叩く・燃やす等の暴力でも成功だが、優しさや工夫のあれば Great、単なる暴力は Good にする。
ゴーレムや状況と無関係な発言、曖昧すぎる発言は Bad。`,
	},
	StageFire: {
		ID:        StageFire,
		Title:     "炎のトラップ",
		Situation: "城の中に入ると、通路が激しい炎で包まれていて進めない!",
		Goal:      "炎を消す、または安全に飛び越える呪文を唱えよう!",
		SuccessSpec: `プレイヤーが炎を消す・凍らせる・水をかける・盾で防ぐ・飛び越える・風で吹き飛ばす等、
火を無効化または回避する行動をとった場合、成功とする。
工夫や勇敢さがあれば Great、やや単調なら Good にする。
炎と無関係な発言は Bad。`,
	},
	StageDragon: {
		ID:               StageDragon,
		Title:            "ドラゴンとの最終決戦",
		Situation:        "最上階でお姫様を発見!でも、怒ったドラゴンが襲いかかってきた!",
		Goal:             "手持ちのアイテムや魔法で、ドラゴンを倒す、または仲良くする呪文を唱えよう!",
		NeedsDragonRoute: true,
		SuccessSpec: `プレイヤーがドラゴンを倒す(剣・魔法・弓等で打ち負かす)、または友だちになる(話しかける・撫でる・食べ物をあげる・歌う)の
いずれかの行動をとった場合、成功とする。
攻撃系は route=defeat、友好系は route=befriend と判定する。
工夫や勇敢さ・優しさがあれば Great、単調なら Good にする。
ドラゴンと無関係な発言は Bad。`,
	},
}

// StageOrder は進行順。
var StageOrder = []StageID{StageGate, StageFire, StageDragon}

// LookupStage は存在確認付きでステージを取得。
func LookupStage(id StageID) (Stage, bool) {
	s, ok := Stages[id]
	return s, ok
}
