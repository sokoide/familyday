package domain

import "testing"

func TestLivesApply(t *testing.T) {
	cases := []struct {
		cur   Lives
		delta int
		want  Lives
	}{
		{3, -1, 2},
		{3, 0, 3},
		{1, -1, 0},
		{0, -1, 0}, // 0未満にならない
		{2, 5, 3},  // Max 超えない
	}
	for _, c := range cases {
		got := c.cur.Apply(c.delta)
		if got != c.want {
			t.Errorf("Lives(%d).Apply(%d) = %d, want %d", c.cur, c.delta, got, c.want)
		}
	}
	if (Lives(0)).Dead() != true {
		t.Error("Lives(0) should be dead")
	}
	if (Lives(3)).Full() != true {
		t.Error("Lives(3) should be full")
	}
}

func TestVerdictRules(t *testing.T) {
	if VerdictGreat.LivesDelta() != 0 || VerdictGreat.Advances() != true {
		t.Error("Great: no damage, advances")
	}
	if VerdictGood.LivesDelta() != -1 || VerdictGood.Advances() != true {
		t.Error("Good: -1, advances")
	}
	if VerdictBad.LivesDelta() != -1 || VerdictBad.Advances() != false {
		t.Error("Bad: -1, retry")
	}
}

func TestDecideEnding(t *testing.T) {
	cases := []struct {
		lives   Lives
		cleared bool
		route   DragonRoute
		want    EndingType
	}{
		{3, true, RouteDefeat, EndingGreat},    // 満ライフクリア
		{2, true, RouteBefriend, EndingGreat},  // 友好
		{2, true, RouteDefeat, EndingSuccess},  // 1-2で撃退
		{1, true, RouteDefeat, EndingSuccess},
		{0, true, RouteDefeat, EndingGameOver}, // ライフ0
		{3, false, RouteNone, EndingGameOver},  // 未クリア
	}
	for _, c := range cases {
		got := DecideEnding(c.lives, c.cleared, c.route)
		if got != c.want {
			t.Errorf("DecideEnding(%d,%v,%q) = %q, want %q", c.lives, c.cleared, c.route, got, c.want)
		}
	}
}

func TestParseStageID(t *testing.T) {
	if _, err := ParseStageID("stage1"); err != nil {
		t.Error(err)
	}
	if _, err := ParseStageID("bogus"); err == nil {
		t.Error("expected error for bogus stage")
	}
	if StageGate.Next() != StageFire || StageFire.Next() != StageDragon || !StageDragon.IsLast() {
		t.Error("stage order wrong")
	}
}
