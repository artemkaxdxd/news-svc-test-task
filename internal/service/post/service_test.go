package post

import (
	"context"
	"testing"

	"news-svc/internal/entity/post"

	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	createFn    func(ctx context.Context, p *post.Post) (string, error)
	getAllFn    func(ctx context.Context, page, limit int64) ([]*post.Post, int64, error)
	getByIDFn   func(ctx context.Context, id string) (*post.Post, error)
	updateFn    func(ctx context.Context, p *post.Post) error
	deleteFn    func(ctx context.Context, id string) error
	searchFn    func(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error)
	getRecentFn func(ctx context.Context, limit int64) ([]*post.Post, error)
}

func (m *mockRepo) Create(ctx context.Context, p *post.Post) (string, error) {
	return m.createFn(ctx, p)
}
func (m *mockRepo) GetAll(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
	return m.getAllFn(ctx, page, limit)
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (*post.Post, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) Update(ctx context.Context, p *post.Post) error {
	return m.updateFn(ctx, p)
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockRepo) Search(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error) {
	return m.searchFn(ctx, q, page, limit)
}
func (m *mockRepo) GetRecent(ctx context.Context, limit int64) ([]*post.Post, error) {
	return m.getRecentFn(ctx, limit)
}

func TestCreateSuccess(t *testing.T) {
	svc := New(&mockRepo{
		createFn: func(ctx context.Context, p *post.Post) (string, error) {
			return "123", nil
		},
	})

	p := &post.Post{Title: "Test", Content: "Content"}
	id, err := svc.Create(context.Background(), p)

	assert.NoError(t, err)
	assert.Equal(t, "123", id)
}

func TestCreateValidationError(t *testing.T) {
	svc := New(&mockRepo{})

	p := &post.Post{} // missing title/content
	_, err := svc.Create(context.Background(), p)

	assert.Error(t, err)
}

func TestGetAllDefaults(t *testing.T) {
	called := false
	svc := New(&mockRepo{
		getAllFn: func(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
			called = true
			assert.Equal(t, int64(1), page)
			assert.Equal(t, int64(10), limit)
			return []*post.Post{}, 0, nil
		},
	})

	_, _, err := svc.GetAll(context.Background(), 0, 0)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestGetByID(t *testing.T) {
	example := &post.Post{ID: "id1"}
	svc := New(&mockRepo{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "id1", id)
			return example, nil
		},
	})

	res, err := svc.GetByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, example, res)
}

func TestUpdateSuccess(t *testing.T) {
	p := &post.Post{ID: "id1", Title: "T", Content: "C"}
	svc := New(&mockRepo{
		updateFn: func(ctx context.Context, post *post.Post) error {
			assert.Equal(t, p, post)
			return nil
		},
	})

	err := svc.Update(context.Background(), p)
	assert.NoError(t, err)
}

func TestUpdateValidationError(t *testing.T) {
	svc := New(&mockRepo{})
	p := &post.Post{} // invalid
	err := svc.Update(context.Background(), p)
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	called := false
	id := "id2"
	svc := New(&mockRepo{
		deleteFn: func(ctx context.Context, got string) error {
			called = true
			assert.Equal(t, id, got)
			return nil
		},
	})

	err := svc.Delete(context.Background(), id)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSearchDefaults(t *testing.T) {
	called := false
	svc := New(&mockRepo{
		searchFn: func(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error) {
			called = true
			assert.Equal(t, "query", q)
			assert.Equal(t, int64(1), page)
			assert.Equal(t, int64(10), limit)
			return []*post.Post{}, 0, nil
		},
	})

	_, _, err := svc.Search(context.Background(), "query", 0, 0)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestGetRecentDefault(t *testing.T) {
	called := false
	svc := New(&mockRepo{
		getRecentFn: func(ctx context.Context, limit int64) ([]*post.Post, error) {
			called = true
			assert.Equal(t, int64(5), limit)
			return []*post.Post{}, nil
		},
	})

	_, err := svc.GetRecent(context.Background(), 0)
	assert.NoError(t, err)
	assert.True(t, called)
}
