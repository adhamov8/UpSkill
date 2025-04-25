// userservice/internal/handlers/user.go
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/upskill/userservice/internal/middleware"
	"github.com/upskill/userservice/internal/models"
	"gorm.io/gorm"
)

/* ──────────────────────────────────────────────────────────────
   TYPES & HELPERS
   ──────────────────────────────────────────────────────────────*/

type UserHandler struct{ db *gorm.DB }

func NewUserHandler(db *gorm.DB) *UserHandler { return &UserHandler{db: db} }

type updateReq struct {
	Track string `json:"track"`
	Goal  string `json:"goal"`
}

type profileResp struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Track     string    `json:"track"`
	Goal      string    `json:"goal"`
	AvatarURL string    `json:"avatar_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResp(p models.Profile) profileResp {
	return profileResp{
		ID:        p.ID,
		UserID:    p.UserID,
		Track:     p.Track,
		Goal:      p.Goal,
		AvatarURL: p.AvatarURL,
		UpdatedAt: p.UpdatedAt,
	}
}

/* ──────────────────────────────────────────────────────────────
   HANDLERS
   ──────────────────────────────────────────────────────────────*/

// GET /api/user/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	var p models.Profile
	err := h.db.Where("user_id = ?", uid).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// профиль ещё не создан → создаём «пустой»
		p = models.Profile{UserID: uid}
		if err := h.db.Create(&p).Error; err != nil {
			http.Error(w, "cannot create profile", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toResp(p))
}

// PUT /api/user/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	req.Track = strings.TrimSpace(req.Track)
	req.Goal = strings.TrimSpace(req.Goal)
	if len(req.Track) == 0 || len(req.Track) > 100 ||
		len(req.Goal) == 0 || len(req.Goal) > 255 {
		http.Error(w, "validation error", http.StatusBadRequest)
		return
	}

	var p models.Profile
	switch err := h.db.Where("user_id = ?", uid).First(&p).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound):
		p = models.Profile{UserID: uid, Track: req.Track, Goal: req.Goal}
		h.db.Create(&p)
	case err == nil:
		p.Track, p.Goal = req.Track, req.Goal
		h.db.Save(&p)
	default:
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toResp(p))
}

// GET /api/user/{id}
func (h *UserHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var p models.Profile
	if h.db.Where("user_id = ?", id).First(&p).Error != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp := toResp(p)
	resp.Goal = "" // цель скрываем
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PUT /api/user/me/avatar  multipart/form-data; file=<binary>
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file err", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(path.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		http.Error(w, "only jpg/png", http.StatusBadRequest)
		return
	}

	dstPath := fmt.Sprintf("./static/avatars/%d%s", uid, ext)
	_ = os.MkdirAll("./static/avatars", 0o755)

	out, _ := os.Create(dstPath)
	_, _ = io.Copy(out, file)
	_ = out.Close()

	url := "/static/avatars/" + strconv.Itoa(int(uid)) + ext
	h.db.Model(&models.Profile{}).Where("user_id = ?", uid).
		Update("avatar_url", url)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"avatar_url": url})
}
