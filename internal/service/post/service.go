package post

import (
	"context"
	"news-svc/internal/entity/post"
)

func (s service) Create(ctx context.Context, post *post.Post) (string, error) {
	if err := post.Validate(); err != nil {
		return "", err
	}

	return s.repo.Create(ctx, post)
}

func (s service) GetAll(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return s.repo.GetAll(ctx, page, limit)
}

func (s service) GetByID(ctx context.Context, id string) (*post.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s service) Update(ctx context.Context, post *post.Post) error {
	if err := post.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, post)
}

func (s service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s service) Search(ctx context.Context, query string, page, limit int64) ([]*post.Post, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return s.repo.Search(ctx, query, page, limit)
}

func (s service) GetRecent(ctx context.Context, limit int64) ([]*post.Post, error) {
	if limit <= 0 {
		limit = 5
	}

	return s.repo.GetRecent(ctx, limit)
}
