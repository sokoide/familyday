package usecase

import "testing"

// sanitizeReason は reason からライフ/ダメージ言及を除去する重要なガード。
// 壊れると「ライフは減らないよ」という矛盾表示が復活するため、回帰テストで守る。
func TestSanitizeReason(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "安全な理由はそのまま",
			in:   "工夫の余地があった",
			want: "工夫の余地があった。",
		},
		{
			name: "ライフ言及を含むセグメントを除去",
			in:   "もう少し工夫があるとGreatだったよ！ライフは減らないよ。",
			want: "もう少し工夫があるとGreatだったよ。",
		},
		{
			name: "ハート言及を含むセグメントを除去",
			in:   "优しく誘えたね。ハートはへらないよ!",
			want: "优しく誘えたね。",
		},
		{
			name: "全セグメントが禁止語 → フォールバック",
			in:   "ライフは減るよ。ハートも減る!",
			want: "がんばったね！",
		},
		{
			name: "空文字はそのまま",
			in:   "",
			want: "",
		},
		{
			name: "英語 damage を除去(区切りは! だが結合後に。が付く)",
			in:   "Good try! No damage!",
			want: "Good try。",
		},
		{
			name: "英語 life を除去",
			in:   "Creative! But you lost a life.",
			want: "Creative。",
		},
		{
			name: "複数セグメント混在: 安全なものだけ残る",
			in:   "優しかったよ。ライフは減らない。でも工夫できるよ!",
			want: "優しかったよ。でも工夫できるよ。",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := sanitizeReason(c.in)
			if got != c.want {
				t.Errorf("sanitizeReason(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
