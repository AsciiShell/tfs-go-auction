package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/asciishell/tfs-go-auktion/internal/database"
	"gitlab.com/asciishell/tfs-go-auktion/pkg/log"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type config struct {
	DB          database.DBCredential
	HTTPAddress string
	HTTPTimeout time.Duration
	MaxRequests int
	Migrate     bool
}

func loadConfig() config {
	cfg := config{}
	flag.StringVar(&cfg.DB.User, "dbuser", "postgres", "DB username")
	flag.StringVar(&cfg.DB.Password, "dbpassword", "", "DB password")
	flag.StringVar(&cfg.DB.Host, "dbhost", "localhost:5432", "DB host with port")
	flag.StringVar(&cfg.DB.Table, "dbtable", "auction", "DB table")
	flag.StringVar(&cfg.HTTPAddress, "address", ":8000", "Server address")
	flag.IntVar(&cfg.MaxRequests, "max-requests", 100, "Maximum number of requests")
	flag.DurationVar(&cfg.HTTPTimeout, "http-timeout", 5*time.Second, "HTTP timeout")
	flag.BoolVar(&cfg.Migrate, "migrate", false, "Run migrations")
	flag.Parse()
	return cfg
}
func main() {
	cfg := loadConfig()

	db, err := database.NewDataBaseStorage(cfg.DB)
	if err != nil {
		log.New().Fatalf("can't use database:%s", err)
	}
	defer func() {
		_ = db.DB.Close()
	}()
	if cfg.Migrate {
		db.Migrate()
		log.New().Info("Migrate completed")
		return
	}

	handler := NewAuctionHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Throttle(cfg.MaxRequests))

	r.Use(middleware.Timeout(cfg.HTTPTimeout))

	r.Route("/v1/auction", func(r chi.Router) {
		r.Post("/signup", handler.PostSignup)
		r.Post("/signin", handler.PostSignin)
		r.Put("/users/{id}", handler.PutUser)
	})

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "swagger")
	FileServer(r, "/swagger", http.Dir(filesDir))
	if err := http.ListenAndServe(cfg.HTTPAddress, r); err != nil {
		log.New().Fatalf("server error:%s", err)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusTemporaryRedirect).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
