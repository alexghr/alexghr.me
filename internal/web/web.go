package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// set during release build
var Version = "dev"

type handler struct {
	templates *templates
	logger    *slog.Logger
	id        string
	debug     bool
}

type model struct {
	InstanceID string
	Debug      bool
}

func NewHTTPHandler(logger *slog.Logger, debug bool, rateLimitBurst uint64, rateLimitRefill float64) (http.Handler, error) {
	static, err := newStaticFileServer(debug)
	if err != nil {
		return nil, fmt.Errorf("static file server: %w", err)
	}

	templates, err := newTemplates(debug)
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	id := strconv.FormatUint(rand.Uint64(), 16)
	h := handler{templates: templates, logger: logger, id: id, debug: debug}
	if !debug {
		static = cache(static, 600*time.Second)
	}

	rateLimit := newLeakyBucket(logger, float64(rateLimitBurst), rateLimitRefill)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", static))
	mux.HandleFunc("GET /{$}", h.serveIndex)
	if debug {
		mux.HandleFunc("GET /api/reload", h.notifyReload)
	}

	return rateLimit.wrap(compress(mux)), nil
}

func (h handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := h.templates.get("index")

	if err != nil {
		h.renderError(w, r)
		return
	}

	m := model{InstanceID: h.id, Debug: h.debug}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", m); err != nil {
		h.logger.ErrorContext(r.Context(), "serve index", "method", r.Method, "path", r.URL.Path, "error", err)
	}
}

func (h handler) renderError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	_, err := w.Write([]byte("<body>there was an error</body>"))
	if err != nil {
		h.logger.ErrorContext(r.Context(), "serve index", "method", r.Method, "path", r.URL.Path, "error", err)
	}
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
