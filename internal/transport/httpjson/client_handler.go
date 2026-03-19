package httpjson

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/service"
)

type ClientHandler struct {
	service *service.ClientService
}

func NewClientHandler(service *service.ClientService) *ClientHandler {
	return &ClientHandler{service: service}
}

func (h *ClientHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{clientID}", h.getByID)
	return r
}

type createClientRequest struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	BirthDate string `json:"birth_date"`
	Country   string `json:"country"`
}

func (h *ClientHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createClientRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	birthDate, err := time.Parse("2006-01-02", request.BirthDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "birth_date must use YYYY-MM-DD format"})
		return
	}

	client, err := h.service.Create(r.Context(), domain.Client{
		FullName:  request.FullName,
		Email:     request.Email,
		BirthDate: birthDate,
		Country:   request.Country,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, client)
}

func (h *ClientHandler) getByID(w http.ResponseWriter, r *http.Request) {
	client, err := h.service.GetByID(r.Context(), chi.URLParam(r, "clientID"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, client)
}

func (h *ClientHandler) list(w http.ResponseWriter, r *http.Request) {
	clients, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, clients)
}
