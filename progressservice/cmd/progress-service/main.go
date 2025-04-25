package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/upskill/progressservice/internal/handlers"
	"github.com/upskill/progressservice/internal/middleware"
	"github.com/upskill/progressservice/internal/models"
	"github.com/upskill/progressservice/internal/seed"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title UpSkill ProgressService API
// @version 1.0
// @BasePath /api/progress
func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_DSN")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(&models.PlanItem{}, &models.Progress{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	seed.Plan(db) // один раз положим 30 занятий, если их ещё нет

	h := handlers.New(db)

	r := chi.NewRouter()
	r.Use(chimw.RealIP, chimw.Logger, chimw.Recoverer, chimw.Timeout(60*time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })

	r.Route("/api/progress", func(rt chi.Router) {
		rt.With(middleware.AuthMW).Get("/", h.ListPlan)
		rt.With(middleware.AuthMW).Patch("/{id}", h.PatchProgress)
	})

	log.Println("ProgressService on :8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}
