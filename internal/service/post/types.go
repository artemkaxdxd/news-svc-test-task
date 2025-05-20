package post

import (
	"context"
	"news-svc/internal/entity/post"
)

type (
	repository interface {
		Create(ctx context.Context, post *post.Post) (string, error)
		GetAll(ctx context.Context, page, limit int64) ([]*post.Post, int64, error)
		GetByID(ctx context.Context, id string) (*post.Post, error)
		Update(ctx context.Context, post *post.Post) error
		Delete(ctx context.Context, id string) error
		Search(ctx context.Context, query string, page, limit int64) ([]*post.Post, int64, error)
		GetRecent(ctx context.Context, limit int64) ([]*post.Post, error)
	}

	service struct {
		repo repository
	}
)

func New(repo repository) service {
	return service{repo}
}
