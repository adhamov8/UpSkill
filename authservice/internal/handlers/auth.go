package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/upskill/authservice/internal/models"
	"github.com/upskill/authservice/internal/utils"
)

type AuthHandler struct{ db *gorm.DB }

func NewAuthHandler(db *gorm.DB) *AuthHandler { return &AuthHandler{db: db} }

/* ---------- REG ---------- */

type regReq struct {
	Email, Password, FirstName, LastName string
}

// POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req regReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if len(req.Password) < 8 {
		http.Error(w, "weak password", 400)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleUser,
	}
	if err := h.db.Create(&u).Error; err != nil {
		http.Error(w, "email exists", 490)
		return
	}

	// письмо или авто-подтверждение
	if os.Getenv("EMAIL_ENABLED") == "true" {
		token, _ := utils.NewRefresh(u.ID)
		link := fmt.Sprintf("%s/api/auth/verify?token=%s", os.Getenv("BASE_URL"), token)
		utils.Send(u.Email, "Verify your email",
			`<p>Ссылка для подтверждения: <a href="`+link+`">verify</a></p>`)
	} else {
		h.db.Model(&u).Update("email_verified", true)
	}

	w.WriteHeader(http.StatusCreated)
}

/* ---------- LOGIN ---------- */

type loginReq struct{ Email, Password string }

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	var u models.User
	if h.db.Where("email = ?", strings.ToLower(req.Email)).First(&u).Error != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	if !u.EmailVerified {
		http.Error(w, "not verified", 403)
		return
	}

	acc, _ := utils.NewAccess(u.ID, string(u.Role))
	ref, _ := utils.NewRefresh(u.ID)
	h.db.Create(&models.RefreshToken{Token: ref, UserID: u.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour)})

	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  acc,
		"refresh_token": ref,
	})
}

/* ---------- REFRESH ---------- */

type refreshReq struct{ RefreshToken string }

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	uid, err := utils.ValidateRefresh(req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid token", 401)
		return
	}
	var rt models.RefreshToken
	if h.db.Where("token=? AND revoked=false", req.RefreshToken).
		First(&rt).Error != nil || rt.ExpiresAt.Before(time.Now()) {
		http.Error(w, "invalid token", 401)
		return
	}
	h.db.Model(&rt).Update("revoked", true) // one-time

	var u models.User
	h.db.First(&u, uid)

	acc, _ := utils.NewAccess(u.ID, string(u.Role))
	newRef, _ := utils.NewRefresh(u.ID)
	h.db.Create(&models.RefreshToken{Token: newRef, UserID: uid,
		ExpiresAt: time.Now().Add(24 * time.Hour)})

	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  acc,
		"refresh_token": newRef,
	})
}

/* ---------- LOGOUT ---------- */

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	h.db.Model(&models.RefreshToken{}).Where("token=?", req.RefreshToken).
		Update("revoked", true)
	w.WriteHeader(http.StatusNoContent)
}

/* ---------- PWD RESET ---------- */

type forgotReq struct{ Email string }

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	var u models.User
	if h.db.Where("email=?", strings.ToLower(req.Email)).First(&u).Error != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	token, _ := utils.NewRefresh(u.ID)
	utils.Send(u.Email, "Password reset",
		os.Getenv("BASE_URL")+"/reset?token="+token)
	w.WriteHeader(http.StatusNoContent)
}

type resetReq struct{ Token, NewPassword string }

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	uid, err := utils.ValidateRefresh(req.Token)
	if err != nil {
		http.Error(w, "bad token", 400)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	h.db.Model(&models.User{}).Where("id=?", uid).
		Update("password_hash", string(hash))
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

/* ---------- EMAIL VERIFY ---------- */

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	uid, err := utils.ValidateRefresh(token)
	if err != nil {
		http.Error(w, "bad token", 400)
		return
	}
	h.db.Model(&models.User{}).Where("id=?", uid).
		Update("email_verified", true)
	w.Write([]byte("verified"))
}

/* ---------- ADMIN ---------- */

func (h *AuthHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(
		r.URL.Path, "/api/auth/admin/users/"), "/role"))

	var body struct{ Role models.Role }
	_ = json.NewDecoder(r.Body).Decode(&body)

	switch body.Role {
	case models.RoleUser, models.RoleMentor, models.RoleAdmin:
	default:
		http.Error(w, "bad role", 400)
		return
	}
	h.db.Model(&models.User{}).Where("id=?", id).
		Update("role", body.Role)
	w.WriteHeader(http.StatusNoContent)
}
