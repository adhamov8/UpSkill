package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/upskill/userservice/internal/handlers"
	"github.com/upskill/userservice/internal/middleware"
	"github.com/upskill/userservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_DSN")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(&models.Profile{}, &models.Mentor{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	hUser := handlers.NewUserHandler(db)
	hMentor := handlers.NewMentorHandler(db)

	r := chi.NewRouter()
	r.Use(chimw.Logger, chimw.Timeout(60*time.Second), chimw.Recoverer)

	/* ---------- users ---------- */
	r.Route("/api/user", func(rt chi.Router) {
		rt.With(middleware.AuthMW).Get("/me", hUser.GetMe)
		rt.With(middleware.AuthMW).Put("/me", hUser.UpdateMe)
		rt.With(middleware.AuthMW).Put("/me/avatar", hUser.UploadAvatar)
	})
	r.Get("/api/user/{id}", hUser.GetPublic)

	/* ---------- mentors ---------- */
	r.Route("/api/mentors", func(rt chi.Router) {
		rt.Get("/", hMentor.List)
		rt.Get("/{id}", hMentor.Get)
	})

	/* ---------- misc ---------- */
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	log.Println("UserService on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}
