package post

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"news-svc/internal/entity/post"
)

func (h handler) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit < 1 {
		limit = 3
	}

	var (
		posts []*post.Post
		total int64
		err   error
	)
	ctx := r.Context()
	if strings.TrimSpace(q) != "" {
		posts, total, err = h.svc.Search(ctx, q, page, limit)
	} else {
		posts, total, err = h.svc.GetAll(ctx, page, limit)
	}
	if err != nil {
		h.l.Error("List error", "err", err)
		http.Error(w, "server error", http.StatusUnprocessableEntity)
		return
	}

	recent, _ := h.svc.GetRecent(ctx, 5)

	totalPages := int64(1)
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}

	data := ListPageData{
		Posts:      posts,
		Recent:     recent,
		Search:     q,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: max(totalPages, 1),
	}

	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.Render(w, "list", data)
		h.tmpl.Render(w, "pagination", data)
	} else {
		h.tmpl.Render(w, "base", data)
	}
}

func (h handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	p := &post.Post{
		Title:   r.Form.Get("title"),
		Content: r.Form.Get("content"),
	}

	id, err := h.svc.Create(r.Context(), p)
	if err != nil {
		h.l.Error("Create error", "err", err)
		h.tmpl.Render(w, "form", ErrorData{Error: err.Error()})
		return
	}

	created, _ := h.svc.GetByID(r.Context(), id)

	w.Header().Set("HX-Trigger", "postCreated")
	h.tmpl.Render(w, "item", created)
}

func (h handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Render(w, "create_form", CreateFormData{})
}

func (h handler) EditForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := EditFormData{
		ID:      p.ID,
		Title:   p.Title,
		Content: p.Content,
	}

	h.tmpl.Render(w, "edit_form", data)
}

func (h handler) Show(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.Render(w, "show", p)
	} else {
		h.tmpl.Render(w, "base", PostsData{Posts: []*post.Post{p}})
	}
}

func (h handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	p := &post.Post{
		ID:      id,
		Title:   r.Form.Get("title"),
		Content: r.Form.Get("content"),
	}

	if err := h.svc.Update(r.Context(), p); err != nil {
		h.l.Error("Update error", "err", err)
		h.tmpl.Render(w, "edit_form", ErrorData{Error: err.Error()})
		return
	}

	updated, _ := h.svc.GetByID(r.Context(), id)
	h.tmpl.Render(w, "item", updated)
}

func (h handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.svc.Delete(r.Context(), id)
	if err != nil {
		h.l.Error("Delete error", "err", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}
