package post

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"news-svc/internal/entity/post"
)

type (
	service interface {
		Create(ctx context.Context, post *post.Post) (string, error)
		GetAll(ctx context.Context, page, limit int64) ([]*post.Post, int64, error)
		GetByID(ctx context.Context, id string) (*post.Post, error)
		Update(ctx context.Context, post *post.Post) error
		Delete(ctx context.Context, id string) error
		Search(ctx context.Context, query string, page, limit int64) ([]*post.Post, int64, error)
		GetRecent(ctx context.Context, limit int64) ([]*post.Post, error)
	}

	templateRenderer interface {
		Render(wr io.Writer, name string, data any) error
	}

	handler struct {
		svc  service
		tmpl templateRenderer
		l    *slog.Logger
	}
)

func InitHandler(
	mux *http.ServeMux,
	svc service,
	l *slog.Logger,
) {
	h := handler{svc, newTemplates(), l}

	mux.HandleFunc("/", h.Index)

	mux.HandleFunc("GET /posts", h.List)
	mux.HandleFunc("POST /posts", h.Create)
	mux.HandleFunc("GET /posts/create", h.CreateForm)

	mux.HandleFunc("GET /posts/{id}", h.Show)
	mux.HandleFunc("GET /posts/{id}/edit", h.EditForm)
	mux.HandleFunc("PATCH /posts/{id}", h.Update)
	mux.HandleFunc("DELETE /posts/{id}", h.Delete)

	return
}

type (
	ListPageData struct {
		Posts      []*post.Post
		Recent     []*post.Post
		Search     string
		Page       int64
		Limit      int64
		Total      int64
		TotalPages int64
		Title      string
		Content    string
		Error      string
	}

	CreateFormData struct {
		Title   string
		Content string
		Error   string
	}

	EditFormData struct {
		ID      string
		Title   string
		Content string
		Error   string
	}

	ErrorData struct {
		Error string
	}

	PostsData struct {
		Posts []*post.Post
	}
)
