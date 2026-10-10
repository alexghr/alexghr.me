package web

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/alexghr/alexghr.me/internal/state"
)

const pageSize = 8

// Absolute limit. Update as more posts get published
const maxPage = 10

type postsModel struct {
	model
	Posts                []state.PostMetadata
	Page, Next, Previous int
}

func (h handler) servePosts(w http.ResponseWriter, r *http.Request) {
	page := 1
	if value := r.URL.Query().Get("page"); value != "" {
		var err error
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 || page > maxPage {
			http.Error(w, "invalid page number", http.StatusBadRequest)
			return
		}
	}

	// Fetch one extra post to determine whether another page exists.
	posts, err := h.store.ListPosts(r.Context(), state.PostStatusPublished, pageSize+1, (page-1)*pageSize)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list posts", "error", err)
		h.renderError(w, r)
		return
	}

	if page > 1 && len(posts) == 0 {
		h.renderNotFound(w, r)
		return
	}
	pageModel := postsModel{model: model{InstanceID: h.id, Debug: h.debug, CurrentPage: "blog"}, Posts: posts, Page: page}
	if page > 1 {
		pageModel.Previous = page - 1
	}
	if len(posts) > pageSize {
		pageModel.Posts = posts[:pageSize]
		pageModel.Next = page + 1
	}

	h.renderPostPage(w, r, "posts", pageModel)
}

func (h handler) servePost(w http.ResponseWriter, r *http.Request) {
	post, err := h.store.GetPost(r.Context(), r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		h.redirect(w, r)
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "get post", "error", err)
		h.renderError(w, r)
		return
	}

	if !post.Status.Public() {
		h.logger.WarnContext(r.Context(), "render draft post", "slug", post.Slug, "error", err)
		h.renderNotFound(w, r)
		return
	}

	body, err := renderPostBody(post)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "render post body", "slug", post.Slug, "error", err)
		h.renderError(w, r)
		return
	}
	h.renderPostPage(w, r, "post", struct {
		model
		Post state.Post
		Body template.HTML
	}{model{InstanceID: h.id, Debug: h.debug, CurrentPage: "post"}, post, body})
}

func renderPostBody(post state.Post) (template.HTML, error) {
	switch post.BodyType {
	case state.PostBodyHtml:
		return template.HTML(post.Body), nil
	case state.PostBodyHtmlTemplate:
		t, err := template.New("body").Parse(post.Body)
		if err != nil {
			return "", err
		}
		var body bytes.Buffer
		if err := t.Execute(&body, post); err != nil {
			return "", err
		}
		return template.HTML(body.String()), nil
	default:
		return "", fmt.Errorf("unsupported post body type %q", post.BodyType)
	}
}

func (h handler) renderPostPage(w http.ResponseWriter, r *http.Request, name string, data any) {
	t, err := h.templates.get(name)
	var page bytes.Buffer
	if err == nil {
		err = t.ExecuteTemplate(&page, "base", data)
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "render post page", "template", name, "error", err)
		h.renderError(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := page.WriteTo(w); err != nil {
		h.logger.ErrorContext(r.Context(), "write post page", "error", err)
	}
}
