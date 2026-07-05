// Package persistence はファイルベースの永続化 adapter。
package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
)

// EndingRepo は1エンディング=1JSONファイルで保存する。
// 1ファイル1エントリなので3並列書き込みでも競合しない。
type EndingRepo struct {
	dir string // メタJSON格納先
}

func NewEndingRepo(dir string) (*EndingRepo, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &EndingRepo{dir: dir}, nil
}

func (r *EndingRepo) Save(ctx context.Context, e domain.Ending) error {
	metaPath := filepath.Join(r.dir, e.ID+".json")
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, b, 0o644)
}

func (r *EndingRepo) Load(ctx context.Context, id string) (domain.Ending, error) {
	// パストラバーサル対策: ID はファイル名のみ
	clean := filepath.Base(id + ".json")
	metaPath := filepath.Join(r.dir, clean)
	b, err := os.ReadFile(metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Ending{}, domain.ErrNotFound
		}
		return domain.Ending{}, err
	}
	var e domain.Ending
	if err := json.Unmarshal(b, &e); err != nil {
		return domain.Ending{}, err
	}
	return e, nil
}

var _ usecase.EndingRepository = (*EndingRepo)(nil)
