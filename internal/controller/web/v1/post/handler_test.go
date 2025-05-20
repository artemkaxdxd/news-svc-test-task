package post

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"news-svc/internal/entity/post"

	"github.com/stretchr/testify/assert"
)

type (
	mockTemplates struct {
		rendered []string
	}

	mockService struct {
		createFn    func(ctx context.Context, p *post.Post) (string, error)
		getAllFn    func(ctx context.Context, page, limit int64) ([]*post.Post, int64, error)
		searchFn    func(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error)
		getRecentFn func(ctx context.Context, limit int64) ([]*post.Post, error)
		getByIDFn   func(ctx context.Context, id string) (*post.Post, error)
		updateFn    func(ctx context.Context, p *post.Post) error
		deleteFn    func(ctx context.Context, id string) error
	}
)

func (f *mockTemplates) Render(w io.Writer, name string, data any) error {
	f.rendered = append(f.rendered, name)
	_, _ = w.Write([]byte(name))
	return nil
}

func (m *mockService) Create(ctx context.Context, p *post.Post) (string, error) {
	return m.createFn(ctx, p)
}
func (m *mockService) GetAll(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
	return m.getAllFn(ctx, page, limit)
}
func (m *mockService) Search(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error) {
	return m.searchFn(ctx, q, page, limit)
}
func (m *mockService) GetRecent(ctx context.Context, limit int64) ([]*post.Post, error) {
	return m.getRecentFn(ctx, limit)
}
func (m *mockService) GetByID(ctx context.Context, id string) (*post.Post, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockService) Update(ctx context.Context, p *post.Post) error {
	return m.updateFn(ctx, p)
}
func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func newHandler(ms *mockService) (*handler, *mockTemplates) {
	ft := &mockTemplates{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := &handler{svc: ms, tmpl: ft, l: logger}
	return h, ft
}

func TestIndexRedirectsToPosts(t *testing.T) {
	hs, _ := newHandler(nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	hs.Index(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/posts", rr.Header().Get("Location"))
}

func TestListSuccessNoQuery(t *testing.T) {
	calledGetAll := false
	calledGetRecent := false
	ms := &mockService{
		getAllFn: func(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
			calledGetAll = true
			assert.Equal(t, int64(1), page)
			assert.Equal(t, int64(3), limit)
			return []*post.Post{}, 0, nil
		},
		getRecentFn: func(ctx context.Context, limit int64) ([]*post.Post, error) {
			calledGetRecent = true
			assert.Equal(t, int64(5), limit)
			return []*post.Post{}, nil
		},
	}
	hs, ft := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)

	hs.List(rr, req)

	assert.True(t, calledGetAll)
	assert.True(t, calledGetRecent)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "base")
	assert.Contains(t, ft.rendered, "base")
}

func TestListSuccessWithQuery(t *testing.T) {
	calledSearch := false
	calledGetRecent := false
	ms := &mockService{
		searchFn: func(ctx context.Context, q string, page, limit int64) ([]*post.Post, int64, error) {
			calledSearch = true
			assert.Equal(t, "foo", q)
			assert.Equal(t, int64(2), page)
			assert.Equal(t, int64(5), limit)
			return []*post.Post{}, 0, nil
		},
		getRecentFn: func(ctx context.Context, limit int64) ([]*post.Post, error) {
			calledGetRecent = true
			return []*post.Post{}, nil
		},
	}
	hs, ft := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts?q=foo&page=2&limit=5", nil)

	hs.List(rr, req)

	assert.True(t, calledSearch)
	assert.True(t, calledGetRecent)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "base")
	assert.Contains(t, ft.rendered, "base")
}

func TestListError(t *testing.T) {
	ms := &mockService{
		getAllFn: func(ctx context.Context, page, limit int64) ([]*post.Post, int64, error) {
			return nil, 0, errors.New("fail")
		},
		getRecentFn: func(ctx context.Context, limit int64) ([]*post.Post, error) { return nil, nil },
	}
	hs, _ := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)

	hs.List(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, "server error\n", rr.Body.String())
}

func TestCreateFormRendersForm(t *testing.T) {
	hs, ft := newHandler(&mockService{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/create", nil)

	hs.CreateForm(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "create_form")
	assert.Contains(t, ft.rendered, "create_form")
}

func TestCreateSuccess(t *testing.T) {
	ms := &mockService{
		createFn: func(ctx context.Context, p *post.Post) (string, error) {
			assert.Equal(t, "T", p.Title)
			assert.Equal(t, "C", p.Content)
			return "id1", nil
		},
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "id1", id)
			return &post.Post{ID: id, Title: "T", Content: "C", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}
	hs, ft := newHandler(ms)

	form := "title=T&content=C"
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	hs.Create(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "postCreated", rr.Header().Get("HX-Trigger"))
	assert.Contains(t, rr.Body.String(), "item")
	assert.Contains(t, ft.rendered, "item")
}

func TestCreateError(t *testing.T) {
	ms := &mockService{
		createFn: func(ctx context.Context, p *post.Post) (string, error) {
			return "", errors.New("fail")
		},
	}
	hs, ft := newHandler(ms)

	form := "title=T&content=C"
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	hs.Create(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "form")
	assert.Contains(t, ft.rendered, "form")
}

func TestShowSuccess(t *testing.T) {
	ms := &mockService{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "123", id)
			return &post.Post{ID: id}, nil
		},
	}
	hs, ft := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}", hs.Show)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "base")
	assert.Contains(t, ft.rendered, "base")
}

func TestShowHTMXRequest(t *testing.T) {
	ms := &mockService{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "123", id)
			return &post.Post{ID: id}, nil
		},
	}
	hs, ft := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123", nil)
	req.Header.Set("HX-Request", "true")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}", hs.Show)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "show")
	assert.Contains(t, ft.rendered, "show")
	assert.NotContains(t, ft.rendered, "base")
}

func TestShowNotFound(t *testing.T) {
	ms := &mockService{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			return nil, errors.New("not found")
		},
	}
	hs, _ := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/404", nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}", hs.Show)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestEditFormRendersForm(t *testing.T) {
	ms := &mockService{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "123", id)
			return &post.Post{ID: id, Title: "T", Content: "C"}, nil
		},
	}
	hs, ft := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123/edit", nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}/edit", hs.EditForm)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "edit_form")
	assert.Contains(t, ft.rendered, "edit_form")
}

func TestEditFormNotFound(t *testing.T) {
	ms := &mockService{
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			return nil, errors.New("not found")
		},
	}
	hs, _ := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/404/edit", nil)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/{id}/edit", hs.EditForm)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateSuccess(t *testing.T) {
	ms := &mockService{
		updateFn: func(ctx context.Context, p *post.Post) error {
			assert.Equal(t, "123", p.ID)
			assert.Equal(t, "T2", p.Title)
			assert.Equal(t, "C2", p.Content)
			return nil
		},
		getByIDFn: func(ctx context.Context, id string) (*post.Post, error) {
			assert.Equal(t, "123", id)
			return &post.Post{ID: id, Title: "T2", Content: "C2"}, nil
		},
	}
	hs, ft := newHandler(ms)

	form := "title=T2&content=C2"
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/posts/123", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /posts/{id}", hs.Update)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "item")
	assert.Contains(t, ft.rendered, "item")
}

func TestUpdateError(t *testing.T) {
	ms := &mockService{
		updateFn: func(ctx context.Context, p *post.Post) error {
			return errors.New("fail")
		},
	}
	hs, ft := newHandler(ms)

	form := "title=T2&content=C2"
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/posts/123", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /posts/{id}", hs.Update)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "edit_form")
	assert.Contains(t, ft.rendered, "edit_form")
}

func TestDeleteSuccess(t *testing.T) {
	ms := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			assert.Equal(t, "123", id)
			return nil
		},
	}
	hs, _ := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/posts/123", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /posts/{id}", hs.Delete)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteError(t *testing.T) {
	ms := &mockService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("fail")
		},
	}
	hs, _ := newHandler(ms)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/posts/123", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /posts/{id}", hs.Delete)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}
