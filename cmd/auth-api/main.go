package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func main() {
	const MaxConcurrentRequest = 100
	var dbUser, dbPassword, dbHost, dbTable string
	flag.StringVar(&dbUser, "dbuser", "postgres", "DB username")
	flag.StringVar(&dbPassword, "dbpassword", "", "DB password")
	flag.StringVar(&dbHost, "dbhost", "localhost:5432", "DB host with port")
	flag.StringVar(&dbTable, "dbtable", "auction", "DB password")
	flag.Parse()
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&fallback_application_name=fintech-app", dbUser, dbPassword, dbHost, dbTable)
	_, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("can't connect to database, dsn %s:%s", dsn, err)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Throttle(MaxConcurrentRequest))

	// r.Use(middleware.Timeout(5 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", PostSignup)
		r.Post("/signin", PostSignin)
		r.Put("/users/{id}", PutUser)
	})

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "swagger")
	FileServer(r, "/swagger", http.Dir(filesDir))
	if err := http.ListenAndServe(":5000", r); err != nil {
		log.Fatal(err)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
