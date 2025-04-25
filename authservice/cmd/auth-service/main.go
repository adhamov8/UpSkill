package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/upskill/authservice/internal/handlers"
	"github.com/upskill/authservice/internal/middleware"
	"github.com/upskill/authservice/internal/models"
	"github.com/upskill/authservice/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title UpSkill AuthService
// @version 1.4
// @BasePath /api/auth
func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_DSN")), &gorm.Config{})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.RefreshToken{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	utils.InitEmail()

	h := handlers.NewAuthHandler(db)

	r := chi.NewRouter()
	r.Use(chimw.Logger, chimw.Timeout(60*time.Second), chimw.Recoverer)

	r.Route("/api/auth", func(rt chi.Router) {
		rt.Post("/register", h.Register)
		rt.Get("/verify", h.VerifyEmail)
		rt.Post("/login", h.Login)
		rt.Post("/refresh", h.RefreshToken)
		rt.With(middleware.AuthMW).Post("/logout", h.Logout)

		rt.Post("/password/forgot", h.ForgotPassword)
		rt.Post("/password/reset", h.ResetPassword)
	})

	adm := r.With(middleware.AuthMW, middleware.RoleMW(models.RoleAdmin))
	adm.Post("/api/auth/admin/users/{id}/role", h.ChangeUserRole)

	log.Println("AuthService :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
