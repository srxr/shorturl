package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	// Logging
	"github.com/unrolled/logger"

	// Stats/Metrics
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/thoas/stats"

	"github.com/GeertJohan/go.rice"
	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

// Counters ...
type Counters struct {
	r metrics.Registry
}

func NewCounters() *Counters {
	counters := &Counters{
		r: metrics.NewRegistry(),
	}
	return counters
}

func (c *Counters) Inc(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(1)
}

func (c *Counters) Dec(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(1)
}

func (c *Counters) IncBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(n)
}

func (c *Counters) DecBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(n)
}

// Server ...
type Server struct {
	bind      string
	config    Config
	templates *Templates
	router    *httprouter.Router

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *Counters
	stats    *stats.Stats
}

func (s *Server) render(name string, w http.ResponseWriter, ctx interface{}) {
	buf, err := s.templates.Exec(name, ctx)
	if err != nil {
		log.Printf("error rendering template %s: %s", name, err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("error writing template buffer: %s", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
}

type IndexContext struct {
	URLList []*URL
}

// IndexHandler ...
func (s *Server) IndexHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.counters.Inc("n_index")

		var urlList []*URL
		err := db.All(&urlList)
		if err != nil {
			log.Printf("error querying urls index: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		ctx := &IndexContext{
			URLList: urlList,
		}

		s.render("index", w, ctx)
	}
}

// ShortenHandler ...
func (s *Server) ShortenHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_shorten")

		u, err := NewURL(r.FormValue("url"))
		if err != nil {
			log.Printf("error creating new url: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		redirectURL, err := url.Parse(fmt.Sprintf("./u/%s", u.ID))
		if err != nil {
			log.Printf("error parsing redirect url ./u/%s: %s", u.ID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	}
}

// ViewHandler ...
func (s *Server) ViewHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var u URL

		s.counters.Inc("n_view")

		id := p.ByName("id")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &u)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("error looking up %s for viewing: %s", id, err)
			http.Error(w, "Iternal Error", http.StatusInternalServerError)
			return
		}

		baseURL, err := url.Parse(s.config.baseURL)
		if err != nil {
			log.Printf("error parsing config.baseURL: %s", s.config.baseURL)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		redirectURL, err := url.Parse(fmt.Sprintf("./r/%s", u.ID))
		if err != nil {
			log.Printf("error parsing redirect url ./r/%s: %s", u.ID, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		fullURL := baseURL.ResolveReference(redirectURL)

		s.render(
			"view", w,
			struct {
				ID  string
				URL string
				Ori string
			}{
				ID:  u.ID,
				URL: fullURL.String(),
				Ori: u.URL,
			},
		)
	}
}

// RedirectHandler ...
func (s *Server) RedirectHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var u URL

		s.counters.Inc("n_redirect")

		id := p.ByName("id")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &u)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("error looking up %s for redirect: %s", id, err)
			http.Error(w, "Iternal Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, u.URL, http.StatusFound)
	}
}

// StatsHandler ...
func (s *Server) StatsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(s.stats.Data())
		if err != nil {
			log.Printf("error marshalling stats: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	}
}

// EditHandler ...
func (s *Server) EditHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var u URL

		id := p.ByName("id")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &u)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("error looking up %s for editing: %s", id, err)
			http.Error(w, "Iternal Error", http.StatusInternalServerError)
			return
		}

		s.render(
			"edit", w,
			struct {
				ID  string
				URL string
			}{
				ID:  u.ID,
				URL: u.URL,
			},
		)
	}
}

// UpdateHandler ...
func (s *Server) UpdateHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var u URL

		id := p.ByName("id")
		target := r.FormValue("url")
		newID := r.FormValue("id")
		if id == "" || target == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := db.One("ID", id, &u)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("error looking up %s for editing: %v", id, err)
			http.Error(w, "Iternal Error", http.StatusInternalServerError)
			return
		}

		err = u.update(newID, target)
		if err != nil {
			log.Printf("error updating %s error: %v", id, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		redirectURL := fmt.Sprintf("/u/%s", newID)

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(
		http.ListenAndServe(
			s.bind,
			s.logger.Handler(
				s.stats.Handler(s.router),
			),
		),
	)
}

// DeleteHandler ...
func (s *Server) DeleteHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		id := p.ByName("id")
		if id == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		err := del(id)
		if err != nil {
			log.Printf("error delete id: %s: %v", id, err)
			http.Error(w, "Iternal Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.ServeFiles(
		"/css/*filepath",
		rice.MustFindBox("static/css").HTTPBox(),
	)

	s.router.ServeFiles(
		"/js/*filepath",
		rice.MustFindBox("static/js").HTTPBox(),
	)

	s.router.GET("/", s.IndexHandler())
	s.router.POST("/", s.ShortenHandler())
	s.router.GET("/u/:id", s.ViewHandler())
	s.router.GET("/r/:id", s.RedirectHandler())
	s.router.GET("/e/:id", s.EditHandler())
	s.router.POST("/e/:id", s.UpdateHandler())
	s.router.GET("/d/:id", s.DeleteHandler())
}

// NewServer ...
func NewServer(bind string, config Config) *Server {
	server := &Server{
		bind:      bind,
		config:    config,
		router:    httprouter.New(),
		templates: NewTemplates("base"),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "shorturl",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
			OutputFlags:          log.LstdFlags,
		}),

		// Stats/Metrics
		counters: NewCounters(),
		stats:    stats.New(),
	}

	// Templates
	box := rice.MustFindBox("templates")

	indexTemplate := template.New("index")
	template.Must(indexTemplate.Parse(box.MustString("index.html")))
	template.Must(indexTemplate.Parse(box.MustString("base.html")))

	viewTemplate := template.New("view")
	template.Must(viewTemplate.Parse(box.MustString("view.html")))
	template.Must(viewTemplate.Parse(box.MustString("base.html")))

	editTemplate := template.New("edit")
	template.Must(editTemplate.Parse(box.MustString("edit.html")))
	template.Must(editTemplate.Parse(box.MustString("base.html")))

	server.templates.Add("index", indexTemplate)
	server.templates.Add("view", viewTemplate)
	server.templates.Add("edit", editTemplate)

	server.initRoutes()

	return server
}
