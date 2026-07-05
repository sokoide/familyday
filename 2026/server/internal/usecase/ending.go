package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/sokoide/familyday/server/internal/domain"
)

// EndingInput は /api/ending の入力 DTO。
type EndingInput struct {
	Lives       int
	FinalAction string // "defeat"|"befriend"|"gameover"
	Cleared     bool
	Lang        domain.Lang
	History     []AdventureEvent
}

// EndingOutput はエンディング生成結果。URL の絶対化は Presentation 層が行う。
type EndingOutput struct {
	EndingID   string
	EndingType domain.EndingType
	Story      string
}

const endingLimit = 5 // 1キー 5回/分
const endingHistoryLimit = 16

// EndingUseCase はエンディング種の決定→ストーリー生成→永続化を指揮する。
type EndingUseCase struct {
	story StoryGenerator
	repo  EndingRepository
	lim   RateLimiter
	idgen IDGenerator
	clock Clock
	log   Logger
}

func NewEndingUseCase(s StoryGenerator, r EndingRepository, l RateLimiter, id IDGenerator, c Clock, log Logger) *EndingUseCase {
	if log == nil {
		log = NopLogger{}
	}
	return &EndingUseCase{story: s, repo: r, lim: l, idgen: id, clock: c, log: log}
}

func (u *EndingUseCase) Resolve(ctx context.Context, in EndingInput, sessionID string) (EndingOutput, error) {
	lives, err := domain.NewLives(in.Lives)
	if err != nil {
		return EndingOutput{}, err
	}
	// finalAction "gameover" は RouteNone へ正規化(解析前に処理)
	var route domain.DragonRoute
	if in.FinalAction == "gameover" {
		route = domain.RouteNone
	} else {
		route, err = domain.ParseDragonRoute(in.FinalAction)
		if err != nil {
			return EndingOutput{}, err
		}
	}

	if !u.lim.Allow(ctx, rateKey("ending", sessionID), endingLimit) {
		return EndingOutput{}, domain.ErrRateLimited
	}

	endingType := domain.DecideEnding(lives, in.Cleared, route)
	history := in.History
	if len(history) > endingHistoryLimit {
		history = history[len(history)-endingHistoryLimit:]
	}

	id := u.idgen.NewID()
	story := fallbackStory(endingType, in.Lang)
	if u.story != nil {
		if generated, err := u.story.Generate(ctx, StoryInput{
			EndingType: endingType,
			Lives:      lives,
			Route:      route,
			Lang:       in.Lang,
			History:    history,
		}); err != nil {
			u.log.Printf("ending: story generate failed (type=%s route=%s): %v", endingType, route, err)
		} else if s := strings.TrimSpace(generated); s != "" {
			story = s
		}
	}

	ending := domain.Ending{
		ID:         id,
		EndingType: endingType,
		Lives:      lives,
		Route:      route,
		Story:      story,
		CreatedAt:  u.clock.NowISO(),
	}

	if err := u.repo.Save(ctx, ending); err != nil {
		return EndingOutput{}, fmt.Errorf("save ending: %w", err)
	}

	return EndingOutput{
		EndingID:   id,
		EndingType: endingType,
		Story:      story,
	}, nil
}

// Load は QR 用に保存済みエンディングを取得する。
func (u *EndingUseCase) Load(ctx context.Context, id string) (domain.Ending, error) {
	return u.repo.Load(ctx, id)
}

func fallbackStory(t domain.EndingType, lang domain.Lang) string {
	if lang == domain.LangEN {
		switch t {
		case domain.EndingGreat:
			return "The hero cast a wonderful spell, befriended the dragon, and returned to town. A grand party with the princess begins!"
		case domain.EndingSuccess:
			return "Battered but brave, the hero drove the dragon away and rescued the princess. Everyone in the castle cheers: \"Thank you!\""
		default:
			return "Chased by the dragon, the hero escapes outside the castle! The princess waves from the window: \"Come save me again!\""
		}
	}
	switch t {
	case domain.EndingGreat:
		return "勇者は見事にドラゴンと仲良くなり、街へ帰りました。お姫様と盛大なパーティーが始まります!"
	case domain.EndingSuccess:
		return "満身創痍になりながらも、勇者はドラゴンを撃退し、お姫様を無事に救出しました。城のみんなが「ありがとう!」と叫びます。"
	default:
		return "勇者はドラゴンに追いかけられて、お城の外へ脱出!お姫様が窓から「また助けに来てねー!」と手を振っています。"
	}
}
