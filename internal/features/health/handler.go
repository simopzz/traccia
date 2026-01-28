package health

import (
	"encoding/json"
	"net/http"
	"traccia/internal/database"
)

type Handler struct {
	db database.Service
}

func NewHandler(db database.Service) *Handler {
	return &Handler{db: db}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(h.db.Health())
	_, _ = w.Write(jsonResp)
}
