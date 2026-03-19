package httpjson

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/service"
)

type BankHandler struct {
	service *service.BankService
}

func NewBankHandler(service *service.BankService) *BankHandler {
	return &BankHandler{service: service}
}

func (h *BankHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{bankID}", h.getByID)
	return r
}

type createBankRequest struct {
	Name string          `json:"name"`
	Type domain.BankType `json:"type"`
}

func (h *BankHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createBankRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	bank, err := h.service.Create(r.Context(), domain.Bank{
		Name: request.Name,
		Type: request.Type,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, bank)
}

func (h *BankHandler) getByID(w http.ResponseWriter, r *http.Request) {
	bank, err := h.service.GetByID(r.Context(), chi.URLParam(r, "bankID"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, bank)
}

func (h *BankHandler) list(w http.ResponseWriter, r *http.Request) {
	banks, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, banks)
}
