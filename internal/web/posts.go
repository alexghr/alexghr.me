package web

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

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

type postModel struct {
	model
	Post        state.Post
	Body        template.HTML
	Head        template.HTML
	PublishedAt string
	UpdatedAt   string
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
	posts, err := h.store.ListPosts(r.Context(), state.ListPostsOptions{
		Status: state.PostStatusPublished,
		Limit:  pageSize + 1,
		Offset: (page - 1) * pageSize,
	})
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "list posts", "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
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

func (h handler) servePost(onlyPublic bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post, err := h.store.GetPost(r.Context(), r.PathValue("slug"))
		if errors.Is(err, sql.ErrNoRows) {
			h.redirect(w, r)
			return
		}
		if err != nil {
			h.requestLogger(r).ErrorContext(r.Context(), "get post", "error", err)
			h.renderError(w, r, "internal server error", http.StatusInternalServerError)
			return
		}

		if onlyPublic && !post.Status.Public() {
			h.requestLogger(r).WarnContext(r.Context(), "render draft post", "slug", post.Slug, "error", err)
			h.renderNotFound(w, r)
			return
		}

		tpl, err := renderPostBody(post)
		if err != nil {
			h.requestLogger(r).ErrorContext(r.Context(), "render post body", "slug", post.Slug, "error", err)
			h.renderError(w, r, "internal server error", http.StatusInternalServerError)
			return
		}

		var publishedAt, updatedAt string

		if !post.PublishedAt.IsZero() {
			publishedAt = post.PublishedAt.Format(time.DateOnly)
		}

		// 24hr grace period for edits 😅
		if !post.UpdatedAt.IsZero() && post.UpdatedAt.Sub(post.PublishedAt) > 24*time.Hour {
			updatedAt = post.UpdatedAt.Format(time.DateOnly)
		}

		h.renderPostPage(w, r, "post", postModel{
			model:       model{InstanceID: h.id, Debug: h.debug, CurrentPage: "post"},
			Post:        post,
			Body:        tpl.Body,
			Head:        tpl.Head,
			PublishedAt: publishedAt,
			UpdatedAt:   updatedAt,
		})
	})
}

type renderedPost struct {
	Body template.HTML
	Head template.HTML
}

func renderPostBody(post state.Post) (renderedPost, error) {
	switch post.BodyType {
	case state.PostBodyHtml:
		return renderedPost{Body: template.HTML(post.Body)}, nil
	case state.PostBodyHtmlTemplate:
		t, err := template.New("body").Parse(post.Body)
		if err != nil {
			return renderedPost{}, err
		}
		var body, head bytes.Buffer
		if err := t.ExecuteTemplate(&body, "body", post); err != nil {
			return renderedPost{}, err
		}
		if t.Lookup("headEnd") != nil {
			if err := t.ExecuteTemplate(&head, "headEnd", post); err != nil {
				return renderedPost{}, err
			}
		}
		return renderedPost{
			Body: template.HTML(body.String()),
			Head: template.HTML(head.String()),
		}, nil

	default:
		return renderedPost{}, fmt.Errorf("unsupported post body type %q", post.BodyType)
	}
}

func (h handler) renderPostPage(w http.ResponseWriter, r *http.Request, name string, data any) {
	t, err := h.templates.get(name)
	var page bytes.Buffer
	if err == nil {
		err = t.ExecuteTemplate(&page, "base", data)
	}
	if err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "render post page", "template", name, "error", err)
		h.renderError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := page.WriteTo(w); err != nil {
		h.requestLogger(r).ErrorContext(r.Context(), "write post page", "error", err)
	}
}
