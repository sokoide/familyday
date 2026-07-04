package usecase

import (
	"context"
	"fmt"

	"github.com/sokoide/familyday/server/internal/domain"
)

// EndingInput は /api/ending の入力 DTO。
type EndingInput struct {
	Lives       int
	FinalAction string // "defeat"|"befriend"|"gameover"
	Cleared     bool
	Lang        domain.Lang
}

// EndingOutput はエンディング生成結果。URL の絶対化は Presentation 層が行う。
type EndingOutput struct {
	EndingID   string
	EndingType domain.EndingType
	Story      string
	ImageFile  string // 相対ファイル名(例: {id}.png)
}

const endingLimit = 5 // 1キー 5回/分

// EndingUseCase はエンディング種の決定→ストーリー生成→画像生成→永続化を指揮する。
type EndingUseCase struct {
	story StoryGenerator
	image ImageGenerator
	repo  EndingRepository
	lim   RateLimiter
	idgen IDGenerator
	clock Clock
	log   Logger
}

func NewEndingUseCase(s StoryGenerator, im ImageGenerator, r EndingRepository, l RateLimiter, id IDGenerator, c Clock, log Logger) *EndingUseCase {
	if log == nil {
		log = NopLogger{}
	}
	return &EndingUseCase{story: s, image: im, repo: r, lim: l, idgen: id, clock: c, log: log}
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

	id := u.idgen.NewID()

	// ストーリー生成(失敗時はフォールバック文)
	story, err := u.story.Generate(ctx, endingType, lives, route, in.Lang)
	if err != nil {
		story = fallbackStory(endingType, in.Lang)
	}

	// 画像生成(失敗時は空 Image → Presentation が fallback 静的画像へ)。
	// エラーは上位へ伝播させないが、運用で原因(課金/クォータ等)が分かるようログ出力。
	img, imgErr := u.image.Generate(ctx, endingType, route)
	if imgErr != nil {
		u.log.Printf("ending: image generate failed (type=%s route=%s): %v", endingType, route, imgErr)
	}

	imageFile := ""
	if imgErr == nil {
		imageFile = id + ".png"
	}

	ending := domain.Ending{
		ID:         id,
		EndingType: endingType,
		Lives:      lives,
		Route:      route,
		Story:      story,
		ImageFile:  imageFile,
		CreatedAt:  u.clock.NowISO(),
	}

	// 画像生成失敗時は保存しない(Presentation が fallback 画像を指す)
	saveImg := Image{}
	if imgErr == nil {
		saveImg = img
	}

	if err := u.repo.Save(ctx, ending, saveImg); err != nil {
		return EndingOutput{}, fmt.Errorf("%w: save ending: %v", domain.ErrUpstream, err)
	}

	return EndingOutput{
		EndingID:   id,
		EndingType: endingType,
		Story:      story,
		ImageFile:  ending.ImageFile,
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
		return "勇者は見事な呪文を唱え、ドラゴンと仲良くなって街へ帰りました。お姫様と盛大なパーティーが始まります!"
	case domain.EndingSuccess:
		return "満身創痍になりながらも、勇者はドラゴンを撃退し、お姫様を無事に救出しました。城のみんなが「ありがとう!」と叫びます。"
	default:
		return "ドラゴンに追いかけられて、勇者は一度お城の外へ脱出!お姫様が窓から「また助けに来てねー!」と手を振っています。"
	}
}
