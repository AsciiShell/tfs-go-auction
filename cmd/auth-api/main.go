package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/asciishell/tfs-go-auction/pkg/environment"

	"gitlab.com/asciishell/tfs-go-auction/internal/database"
	"gitlab.com/asciishell/tfs-go-auction/pkg/log"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type config struct {
	DB          database.DBCredential
	HTTPAddress string
	HTTPTimeout time.Duration
	MaxRequests int
	PrintConfig bool
}

func loadConfig() config {
	cfg := config{}
	cfg.DB.User = environment.GetStr("DB_USER", "auction")
	cfg.DB.Password = environment.GetStr("DB_PASSWORD", "postgres")
	cfg.DB.Database = environment.GetStr("DB_DATABASE", "auction")
	cfg.DB.Host = environment.GetStr("DB_HOST", "localhost:5432")
	cfg.DB.Repetitions = environment.GetInt("DB_ATTEMPTS", 10)
	cfg.MaxRequests = environment.GetInt("MAX_REQUESTS", 100)
	cfg.HTTPAddress = environment.GetStr("ADDRESS", ":8000")
	cfg.HTTPTimeout = environment.GetDuration("HTTP_TIMEOUT", 500*time.Second)
	cfg.PrintConfig = environment.GetBool("PRINT_CONFIG", false)
	if cfg.PrintConfig {
		log.New().Infof("%+v", cfg)
	}
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
	db.Migrate()
	log.New().Info("Migrate completed")

	handler := NewAuctionHandler(db)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Throttle(cfg.MaxRequests))

	//r.Use(middleware.Timeout(cfg.HTTPTimeout))

	r.Route("/v1/auction", func(r chi.Router) {
		r.Post("/signup", handler.PostSignup)
		r.Post("/signin", handler.PostSignin)
		r.Route("/users", func(r chi.Router) {
			r.Use(handler.Authenticator)
			r.Put("/{id}", handler.PutUser)
			r.Get("/{id}", handler.GetUser)
			r.Get("/{id}/lots", handler.NotImplemented)
		})
		r.Route("/lots", func(r chi.Router) {
			r.Get("/", handler.NotImplemented)
			r.Post("/", handler.NotImplemented)
			r.Put("/{id}/buy", handler.NotImplemented)
			r.Get("/{id}", handler.NotImplemented)
			r.Put("/{id}", handler.NotImplemented)
		})
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
