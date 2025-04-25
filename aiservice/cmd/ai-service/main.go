package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var aiEnabled bool
var client *openai.Client

func initOpenAI() {
	aiEnabled = os.Getenv("AI_ENABLED") == "true"
	if !aiEnabled {
		log.Println("[AI] offline-режим (AI_ENABLED=false)")
		return
	}

	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("[AI] переменная OPENAI_API_KEY не задана")
	}

	cfg := openai.DefaultConfig(key)
	cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second} // ↑ тай-аут
	client = openai.NewClientWithConfig(cfg)
	log.Println("[AI] online-режим, gpt-3.5-turbo")
}

func main() {
	_ = godotenv.Load()
	initOpenAI()

	r := chi.NewRouter()
	r.Use(chimw.Logger, chimw.Timeout(60*time.Second), chimw.Recoverer)

	r.Post("/api/ai/assessment", handleAssessment)
	r.Post("/api/ai/plan", handlePlan)

	log.Println("AI-Service :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}

// ---------- handlers --------------------------------------------------------

func handleAssessment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answers []string `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Answers) == 0 {
		http.Error(w, "bad json", 400)
		return
	}

	if !aiEnabled {
		_ = json.NewEncoder(w).Encode(mockAssessment())
		return
	}

	prompt := buildAssessmentPrompt(req.Answers)
	out, err := askGPT(r.Context(), prompt)
	if err != nil {
		writeAiErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func handlePlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Track string   `json:"track"`
		Level string   `json:"level"`
		Gaps  []string `json:"gaps"`
		Weeks int      `json:"weeks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	if !aiEnabled {
		_ = json.NewEncoder(w).Encode(mockPlan())
		return
	}

	prompt := fmt.Sprintf(
		"Составь %d-недельный график обучения «%s %s». "+
			"Выведи JSON [{week:1,topic:\"...\",resource:\"URL\"},...]. "+
			"Особый упор на пробелы: %s.",
		req.Weeks, req.Level, req.Track, strings.Join(req.Gaps, ", "),
	)

	out, err := askGPT(r.Context(), prompt)
	if err != nil {
		writeAiErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// ---------- OpenAI helper ---------------------------------------------------

func askGPT(ctx context.Context, prompt string) ([]byte, error) {
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "Ты – карьерный ИИ-наставник"},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	return []byte(resp.Choices[0].Message.Content), nil
}

func writeAiErr(w http.ResponseWriter, err error) {
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		log.Printf("[AI] %d %s", apiErr.HTTPStatusCode, apiErr.Message)

		// graceful-degradation: при 429 возвращаем мок-ответ
		if apiErr.HTTPStatusCode == http.StatusTooManyRequests {
			_ = json.NewEncoder(w).Encode(mockAssessment())
			return
		}

		http.Error(w, apiErr.Message, apiErr.HTTPStatusCode)
		return
	}
	log.Println("[AI]", err)
	http.Error(w, "ai error", http.StatusBadGateway)
}

// ---------- prompts & mocks -------------------------------------------------

func buildAssessmentPrompt(a []string) string {
	var sb strings.Builder
	sb.WriteString("Определи уровень junior/middle/senior и три основных пробела знаний. ")
	sb.WriteString(`Ответ JSON: {"level":"","gaps":["","",""]}.` + "\n")
	for i, ans := range a {
		sb.WriteString(time.Now().Format("15:04:05"))
		sb.WriteString(" #Q")
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(": ")
		sb.WriteString(ans)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func mockAssessment() map[string]any {
	return map[string]any{
		"level": "junior",
		"gaps":  []string{"Algorithms", "Databases", "Networking"},
	}
}
func mockPlan() []map[string]any {
	return []map[string]any{
		{"week": 1, "topic": "SQL basics", "resource": "https://sqlbolt.com"},
		{"week": 2, "topic": "Go + PostgreSQL", "resource": "https://go.dev/doc/database"},
	}
}
