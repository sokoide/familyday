package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/sokoide/familyday/server/internal/domain"
)

// --- フェイク実装(ポートを満たすテスト用ダブル) ---

type fakeJudge struct {
	result JudgeResult
	err    error
}
type fakeStory struct{ text string }
type fakeImage struct {
	img Image
	err error
}
type fakeRepo struct {
	saved   map[string]domain.Ending
	saveErr error
}
type fakeLimiter struct{ allow bool }
type fakeID struct{ id string }
type fakeClock struct{ ts string }

func (f *fakeJudge) Judge(ctx context.Context, s domain.Stage, input string, lang domain.Lang) (JudgeResult, error) {
	return f.result, f.err
}
func (f *fakeStory) Generate(ctx context.Context, t domain.EndingType, l domain.Lives, r domain.DragonRoute, lang domain.Lang) (string, error) {
	return f.text, nil
}
func (f *fakeImage) Generate(ctx context.Context, t domain.EndingType, r domain.DragonRoute) (Image, error) {
	return f.img, f.err
}
func (f *fakeRepo) Save(ctx context.Context, e domain.Ending, img Image) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.saved == nil {
		f.saved = map[string]domain.Ending{}
	}
	f.saved[e.ID] = e
	return nil
}
func (f *fakeRepo) Load(ctx context.Context, id string) (domain.Ending, error) {
	if e, ok := f.saved[id]; ok {
		return e, nil
	}
	return domain.Ending{}, domain.ErrNotFound
}
func (f *fakeLimiter) Allow(ctx context.Context, key string, lim int) bool  { return f.allow }
func (f *fakeID) NewID() string                                             { return f.id }
func (f *fakeClock) NowISO() string                                         { return f.ts }

// --- JudgeUseCase ---

func TestJudgeUseCase_Great(t *testing.T) {
	uc := NewJudgeUseCase(
		&fakeJudge{result: JudgeResult{Verdict: domain.VerdictGreat, Message: "ok"}},
		&fakeLimiter{allow: true},
	)
	out, err := uc.Judge(context.Background(), JudgeInput{StageID: "stage1", SessionID: "s1", Input: "どいて"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Verdict != domain.VerdictGreat || out.LivesDelta != 0 || !out.Advance {
		t.Errorf("unexpected great out: %+v", out)
	}
}

func TestJudgeUseCase_BadInput(t *testing.T) {
	uc := NewJudgeUseCase(&fakeJudge{}, &fakeLimiter{allow: true})
	if _, err := uc.Judge(context.Background(), JudgeInput{StageID: "stage1", Input: ""}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput, got %v", err)
	}
	if _, err := uc.Judge(context.Background(), JudgeInput{StageID: "bogus", Input: "x"}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput for bad stage, got %v", err)
	}
}

func TestJudgeUseCase_RateLimited(t *testing.T) {
	uc := NewJudgeUseCase(&fakeJudge{}, &fakeLimiter{allow: false})
	if _, err := uc.Judge(context.Background(), JudgeInput{StageID: "stage1", Input: "x"}); err != domain.ErrRateLimited {
		t.Errorf("want ErrRateLimited, got %v", err)
	}
}

// --- EndingUseCase ---

func TestEndingUseCase_Great(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewEndingUseCase(
		&fakeStory{text: "story"},
		&fakeImage{img: Image{Bytes: []byte("png"), MIME: "image/png"}},
		repo,
		&fakeLimiter{allow: true},
		&fakeID{id: "abc123"},
		&fakeClock{ts: "2026-07-02T00:00:00Z"},
		NopLogger{},
	)
	out, err := uc.Resolve(context.Background(), EndingInput{Lives: 3, FinalAction: "befriend", Cleared: true}, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if out.EndingType != domain.EndingGreat {
		t.Errorf("want great, got %q", out.EndingType)
	}
	if out.EndingID != "abc123" || out.ImageFile != "abc123.png" {
		t.Errorf("id/imagefile wrong: %+v", out)
	}
	if _, ok := repo.saved["abc123"]; !ok {
		t.Error("ending not saved")
	}
}

func TestEndingUseCase_GameOver(t *testing.T) {
	uc := NewEndingUseCase(
		&fakeStory{text: "s"},
		&fakeImage{},
		&fakeRepo{},
		&fakeLimiter{allow: true},
		&fakeID{id: "x"},
		&fakeClock{ts: "ts"},
		NopLogger{},
	)
	out, err := uc.Resolve(context.Background(), EndingInput{Lives: 0, FinalAction: "gameover", Cleared: false}, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if out.EndingType != domain.EndingGameOver {
		t.Errorf("want gameover, got %q", out.EndingType)
	}
}

func TestEndingUseCase_ImageFailureFallback(t *testing.T) {
	// 画像生成失敗でもストーリーは fallback せず生成済み、保存は成功する
	repo := &fakeRepo{}
	uc := NewEndingUseCase(
		&fakeStory{text: "story-ok"},
		&fakeImage{err: context.Canceled}, // 画像エラー
		repo,
		&fakeLimiter{allow: true},
		&fakeID{id: "id1"},
		&fakeClock{ts: "ts"},
		NopLogger{},
	)
	out, err := uc.Resolve(context.Background(), EndingInput{Lives: 2, FinalAction: "defeat", Cleared: true}, "s1")
	if err != nil {
		t.Fatal(err)
	}
	if out.Story != "story-ok" {
		t.Errorf("story should come from generator, got %q", out.Story)
	}
	if _, ok := repo.saved["id1"]; !ok {
		t.Error("should still save on image failure")
	}
}
