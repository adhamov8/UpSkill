package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	authURL := mustParse(os.Getenv("AUTH_URL"), "http://authservice:8080")
	userURL := mustParse(os.Getenv("USER_URL"), "http://userservice:8081")
	aiURL := mustParse(os.Getenv("AI_URL"), "http://aiservice:8082")
	progressURL := mustParse(os.Getenv("PROGRESS_URL"), "http://progressservice:8083")

	r := chi.NewRouter()
	r.Use(chimw.RealIP, chimw.Logger, chimw.Recoverer, chimw.Timeout(60*time.Second))
	r.Use(corsMW) // разрешаем фронту с 3000/5173 и т.д.

	// ---------- маршруты ----------
	// 1. /api/auth – без проверки токена
	r.Mount("/api/auth", newProxy(authURL, true))

	// 2. защищённые сервисы
	jwtSec := func(u *url.URL) http.Handler { return jwtProxy(u) }
	r.Mount("/api/user", jwtSec(userURL))
	r.Mount("/api/ai", jwtSec(aiURL))
	r.Mount("/api/progress", jwtSec(progressURL))

	log.Println("Gateway listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

/* ---------- helpers ---------- */

func mustParse(raw, def string) *url.URL {
	if raw == "" {
		raw = def
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("bad url %q: %v", raw, err)
	}
	return u
}

// newProxy(target, stripAuth) – создаёт reverse proxy.
// stripAuth=true → удаляем Authorization перед отправкой (AuthService сам не проверяет JWT).
func newProxy(target *url.URL, stripAuth bool) *httputil.ReverseProxy {
	p := httputil.NewSingleHostReverseProxy(target)
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy to %s: %v", target, err)
		http.Error(w, "upstream error", http.StatusBadGateway)
	}
	if stripAuth {
		orig := p.Director
		p.Director = func(r *http.Request) {
			r.Header.Del("Authorization")
			orig(r)
		}
	}
	return p
}

// проверяем JWT, кладём X-User-ID, затем проксируем дальше
func jwtProxy(target *url.URL) http.Handler {
	proxy := newProxy(target, false)
	secret := []byte(os.Getenv("JWT_SECRET"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		t, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
			if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || !t.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub := t.Claims.(jwt.MapClaims)["sub"]
		r.Header.Set("X-User-ID", toString(sub))
		proxy.ServeHTTP(w, r)
	})
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case float64:
		return strconv.FormatInt(int64(x), 10)
	default:
		return fmt.Sprintf("%v", x)
	}
}

/* ---------- very simple CORS ---------- */

func corsMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
