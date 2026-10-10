package web

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexghr/alexghr.me/internal/state"
)

func (h handler) servePosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListPosts(r.Context(), state.PostStatusPublished)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list posts", "error", err)
		h.renderError(w, r)
		return
	}
	h.renderPostPage(w, r, "posts", struct {
		model
		Posts []state.PostMetadata
	}{model{InstanceID: h.id, Debug: h.debug}, posts})
}

func (h handler) servePost(w http.ResponseWriter, r *http.Request) {
	post, err := h.store.GetPost(r.Context(), r.PathValue("slug"))
	if errors.Is(err, sql.ErrNoRows) {
		h.renderNotFound(w, r)
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
	}{model{InstanceID: h.id, Debug: h.debug}, post, body})
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
