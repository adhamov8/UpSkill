package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/upskill/userservice/internal/models"
	"gorm.io/gorm"
)

type MentorHandler struct{ db *gorm.DB }

func NewMentorHandler(db *gorm.DB) *MentorHandler { return &MentorHandler{db: db} }

// ── DTO ──────────────────────────────────────────────────────────

type mentorCard struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Track       string `json:"track"`
	ExperienceY uint   `json:"experience_years"`
	Education   string `json:"education"`
	AvatarURL   string `json:"avatar_url"`
	ContactURL  string `json:"contact_url"`
}

func toCard(m models.Mentor) mentorCard {
	return mentorCard{
		ID: m.ID, FirstName: m.FirstName, LastName: m.LastName,
		Track: m.Track, ExperienceY: m.ExperienceY,
		Education: m.Education, AvatarURL: m.AvatarURL, ContactURL: m.ContactURL,
	}
}

// ── list + filters ──────────────────────────────────────────────
// GET /api/mentors?track=&exp_min=&education=&gender=
func (h *MentorHandler) List(w http.ResponseWriter, r *http.Request) {
	var ms []models.Mentor
	q := h.db

	if v := strings.TrimSpace(r.URL.Query().Get("track")); v != "" {
		q = q.Where("LOWER(track)=?", strings.ToLower(v))
	}
	if v := r.URL.Query().Get("exp_min"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			q = q.Where("experience_y >= ?", n)
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("education")); v != "" {
		q = q.Where("LOWER(education) LIKE ?", "%"+strings.ToLower(v)+"%")
	}
	if v := strings.TrimSpace(r.URL.Query().Get("gender")); v != "" {
		q = q.Where("gender = ?", strings.ToUpper(v))
	}
	q.Find(&ms)

	out := make([]mentorCard, len(ms))
	for i, m := range ms {
		out[i] = toCard(m)
	}
	_ = json.NewEncoder(w).Encode(out)
}

// ── single ──────────────────────────────────────────────────────
// GET /api/mentors/{id}
func (h *MentorHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var m models.Mentor
	if h.db.First(&m, id).Error != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(m)
}
