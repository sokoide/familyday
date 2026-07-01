package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
)

// Presentation 層は UseCase を通さず、port を stub した UseCase で検証する。
// ここでは UseCase のフェイクを組み立てて handler を叩く。

// --- UseCase を丸ごと差し替える最小スタブ(handler は concrete な UseCase 型を要求するため、
//     テスト専用にフェイク gateway を注入して UseCase を構築) ---

type stubJudge struct {
	out usecase.JudgeOutput
	err error
}
type stubRepo struct {
	e   domain.Ending
	err error
}
type stubImg struct{}
type stubStory struct{ text string }

func (s *stubJudge) Judge(ctx context.Context, st domain.Stage, in string, lang domain.Lang) (usecase.JudgeResult, error) {
	return usecase.JudgeResult{Verdict: s.out.Verdict, Route: domain.DragonRoute(s.out.Route), Message: s.out.Message}, s.err
}
func (s *stubStory) Generate(ctx context.Context, t domain.EndingType, l domain.Lives, r domain.DragonRoute, lang domain.Lang) (string, error) {
	return s.text, nil
}
func (s *stubImg) Generate(ctx context.Context, t domain.EndingType, r domain.DragonRoute) (usecase.Image, error) {
	return usecase.Image{}, nil
}
func (s *stubRepo) Save(ctx context.Context, e domain.Ending, img usecase.Image) error { return nil }
func (s *stubRepo) Load(ctx context.Context, id string) (domain.Ending, error)          { return s.e, s.err }

type fakeLimiter struct{ allow bool }
type fakeID struct{ id string }
type fakeClock struct{ ts string }

func (f *fakeLimiter) Allow(context.Context, string, int) bool { return f.allow }
func (f *fakeID) NewID() string                                { return f.id }
func (f *fakeClock) NowISO() string                            { return f.ts }

func newHandlerWithStubs(t *testing.T, judgeOut usecase.JudgeOutput, judgeErr error, ending domain.Ending, loadErr error) *Handler {
	t.Helper()
	juc := usecase.NewJudgeUseCase(&stubJudge{out: judgeOut, err: judgeErr}, &fakeLimiter{allow: true})
	euc := usecase.NewEndingUseCase(&stubStory{text: "s"}, &stubImg{}, &stubRepo{e: ending, err: loadErr}, &fakeLimiter{allow: true}, &fakeID{id: "abc"}, &fakeClock{ts: "ts"})
	return NewHandler(juc, euc, "https://example.com")
}

func TestJudgeHandler_OK(t *testing.T) {
	h := newHandlerWithStubs(t, usecase.JudgeOutput{Verdict: domain.VerdictGreat, Message: "ok", Advance: true}, nil, domain.Ending{}, nil)
	body := strings.NewReader(`{"stageId":"stage1","sessionId":"s1","input":"どいて!"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/judge", body)
	h.Judge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var res JudgeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Verdict != "Great" || !res.Advance {
		t.Errorf("unexpected: %+v", res)
	}
}

func TestJudgeHandler_BadInput(t *testing.T) {
	h := newHandlerWithStubs(t, usecase.JudgeOutput{}, nil, domain.Ending{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/judge", strings.NewReader(`{"stageId":"stage1","input":""}`))
	h.Judge(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestJudgeHandler_Upstream(t *testing.T) {
	h := newHandlerWithStubs(t, usecase.JudgeOutput{}, errors.Join(domain.ErrUpstream), domain.Ending{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/judge", strings.NewReader(`{"stageId":"stage1","input":"x"}`))
	h.Judge(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
}

func TestResultHandler_NotFound(t *testing.T) {
	h := newHandlerWithStubs(t, usecase.JudgeOutput{}, nil, domain.Ending{}, domain.ErrNotFound)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/result/nope", nil)
	req.SetPathValue("id", "nope")
	h.Result(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestResultHandler_OK_URLsAbsolute(t *testing.T) {
	e := domain.Ending{ID: "xyz", EndingType: domain.EndingSuccess, Story: "story", ImageFile: "xyz.png", CreatedAt: "2026-07-02T00:00:00Z"}
	h := newHandlerWithStubs(t, usecase.JudgeOutput{}, nil, e, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/result/xyz", nil)
	h.Result(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var res ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.ImageURL != "https://example.com/img/xyz.png" {
		t.Errorf("bad image url: %s", res.ImageURL)
	}
	if res.ResultURL != "https://example.com/r/xyz" {
		t.Errorf("bad result url: %s", res.ResultURL)
	}
}
