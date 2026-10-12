package web

import (
	"archive/zip"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/alexghr/alexghr.me/internal/state"
)

type handler struct {
	templates *templates
	logger    *slog.Logger
	id        string
	store     *state.Store
	debug     bool
	publicURL *url.URL
	archives  *archiveCache
}

type model struct {
	InstanceID  string
	Debug       bool
	CurrentPage string
}

type HandlerOptions struct {
	Debug           bool
	RateLimitBurst  uint64
	RateLimitRefill float64
	AdminPassword   string
	DraftPassword   string
	PublicURL       string
}

func NewHTTPHandler(store *state.Store, logger *slog.Logger, options HandlerOptions) (http.Handler, error) {
	static, err := newStaticFileServer(options.Debug)
	if err != nil {
		return nil, fmt.Errorf("static file server: %w", err)
	}

	templates, err := newTemplates(options.Debug)
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	publicURL, err := url.Parse(options.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("public URL %s: %w", options.PublicURL, err)
	}

	id := strconv.FormatUint(rand.Uint64(), 16)

	archives := new(archiveCache)
	archives.cache = make(map[string]*zip.Reader)
	h := handler{
		id:        id,
		store:     store,
		templates: templates,
		archives:  archives,
		debug:     options.Debug,
		publicURL: publicURL,
		logger:    logger,
	}

	if !options.Debug {
		static = cache(static, 600*time.Second)
	}

	rateLimit := newLeakyBucket(logger,
		float64(options.RateLimitBurst), options.RateLimitRefill)

	adminPasswordBytes, err := decodePassword(options.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("decode admin password: %w", err)
	}

	draftPasswordBytes, err := decodePassword(options.DraftPassword)
	if err != nil {
		return nil, fmt.Errorf("decode draft password: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", static))

	mux.HandleFunc("GET /{$}", h.serveIndex)
	mux.HandleFunc("GET /blog/{$}", h.servePosts)

	mux.Handle("GET /blog/{slug}/{$}", h.servePost(true))
	mux.Handle("GET /blog/{slug}/assets/{path...}", h.servePostAsset(true))

	mux.Handle("GET /draft/blog/{slug}/{$}", basicAuth(h.servePost(false), "draft", "", draftPasswordBytes))
	mux.Handle("GET /draft/blog/{slug}/assets/{path...}", basicAuth(h.servePostAsset(false), "draft", "", draftPasswordBytes))

	mux.HandleFunc("GET /robots.txt", h.serveRobots)
	mux.HandleFunc("GET /asset/{hash}/{path...}", h.serveAsset)
	mux.HandleFunc("GET /rss.xml", h.serverRSS)
	if options.Debug {
		mux.HandleFunc("GET /api/reload", h.notifyReload)
	}

	admin := basicAuth(adminHandler(h), "admin", "admin", adminPasswordBytes)
	mux.Handle("/admin/", http.StripPrefix("/admin", admin))

	mux.HandleFunc("/{redirect...}", h.redirect)

	var result http.Handler = accessLog(logger, rateLimit.wrap(compress(mux)))
	if !options.Debug {
		// Development uses inline JavaScript and a reload script.
		result = securityHeaders(result)
	}
	return result, nil
}

func (h handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := h.templates.get("index")

	if err != nil {
		h.renderError(w, r)
		return
	}

	m := model{InstanceID: h.id, Debug: h.debug, CurrentPage: "home"}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", m); err != nil {
		h.logger.ErrorContext(r.Context(), "serve index", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (h handler) serveRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	io.WriteString(w, `User-agent: *
Disallow: /api/
Disallow: /admin
`)
}

func (h handler) redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	redir, err := h.store.GetRedirect(r.Context(), r.URL.Path)
	if errors.Is(err, sql.ErrNoRows) {
		h.renderNotFound(w, r)
		return
	}

	if err != nil {
		h.renderError(w, r)
		return
	}

	if !redir.Active {
		h.renderNotFound(w, r)
		return
	}

	to, err := url.Parse(redir.ToURL)
	if err != nil {
		h.renderNotFound(w, r)
		return
	}

	if to.Path == r.URL.Path {
		h.renderNotFound(w, r)
		return
	}

	if to.RawQuery == "" {
		to.RawQuery = r.URL.RawQuery
	} else if r.URL.RawQuery != "" {
		query := r.URL.Query()
		// Configured destination parameters take precedence over incoming ones.
		for key, values := range to.Query() {
			query[key] = values
		}
		to.RawQuery = query.Encode()
	}

	code := http.StatusTemporaryRedirect
	if redir.Permanent {
		code = http.StatusPermanentRedirect
	}

	http.Redirect(w, r, to.String(), code)
}

func (h handler) renderError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	_, err := w.Write([]byte("<body>there was an error</body>"))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "render 500", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (h handler) renderNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

//go:embed static
var embeddedStaticFS embed.FS

func newStaticFileServer(debug bool) (http.Handler, error) {
	var files fs.FS
	if debug {
		root, err := os.OpenRoot("internal/web/static")
		if err != nil {
			return nil, fmt.Errorf("open local static/: %w", err)
		}

		files = root.FS()
	} else {
		fs, err := fs.Sub(embeddedStaticFS, "static")

		if err != nil {
			return nil, fmt.Errorf("open embedded static/: %w", err)
		}

		files = fs
	}

	return http.FileServerFS(files), nil
}

//go:embed templates
var embeddedTemplateFS embed.FS

type templates struct {
	fs    fs.FS
	cache map[string]*template.Template
	mu    sync.Mutex
}

func newTemplates(debug bool) (*templates, error) {
	if debug {
		root, err := os.OpenRoot("internal/web/templates")
		if err != nil {
			return nil, fmt.Errorf("open local templates: %w", err)
		}

		return &templates{fs: root.FS(), cache: nil}, nil
	}

	fs, err := fs.Sub(embeddedTemplateFS, "templates")
	if err != nil {
		return nil, err
	}

	return &templates{
		fs:    fs,
		cache: make(map[string]*template.Template),
	}, nil
}

func (t *templates) get(name string) (*template.Template, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cache != nil {
		tmpl, found := t.cache[name]
		if found {
			return tmpl, nil
		}
	}

	// Load shared defaults first so each page can override them.
	tmpl, err := template.ParseFS(t.fs, "base.html", name+".html")
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}

	if t.cache != nil {
		t.cache[name] = tmpl
	}

	return tmpl, nil
}
