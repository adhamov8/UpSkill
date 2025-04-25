package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/upskill/progressservice/internal/middleware"
	"github.com/upskill/progressservice/internal/models"
	"gorm.io/gorm"
)

type Handler struct{ db *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{db: db} }

type itemResp struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	ResourceURL string     `json:"resource"`
	StartAt     time.Time  `json:"start_at"`
	DueAt       time.Time  `json:"due_at"`
	Done        bool       `json:"done"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

// GET /api/progress
func (h *Handler) ListPlan(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)

	var items []models.PlanItem
	h.db.Order("start_at").Find(&items)

	var prog []models.Progress
	h.db.Where("user_id = ?", uid).Find(&prog)
	pMap := make(map[uint64]models.Progress, len(prog))
	for _, p := range prog {
		pMap[p.PlanItemID] = p
	}

	resp := make([]itemResp, 0, len(items))
	for _, it := range items {
		pr := pMap[it.ID]
		resp = append(resp, itemResp{
			ID: it.ID, Title: it.Title, ResourceURL: it.ResourceURL,
			StartAt: it.StartAt, DueAt: it.DueAt,
			Done: pr.Done, FinishedAt: pr.FinishedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// PATCH /api/progress/{id}
func (h *Handler) PatchProgress(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r)
	id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	var p models.Progress
	err := h.db.Where("user_id=? AND plan_item_id=?", uid, id).First(&p).Error
	switch err {
	case gorm.ErrRecordNotFound:
		now := time.Now()
		p = models.Progress{UserID: uid, PlanItemID: id, Done: true, FinishedAt: &now}
		h.db.Create(&p)
	case nil:
		if !p.Done {
			now := time.Now()
			h.db.Model(&p).Updates(models.Progress{Done: true, FinishedAt: &now})
		}
	default:
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
