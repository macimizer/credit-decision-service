package httpjson

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/service"
)

type CreditHandler struct {
	service *service.CreditService
}

func NewCreditHandler(service *service.CreditService) *CreditHandler {
	return &CreditHandler{service: service}
}

func (h *CreditHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{creditID}", h.getByID)
	return r
}

type createCreditRequest struct {
	ClientID   string            `json:"client_id"`
	BankID     string            `json:"bank_id"`
	MinPayment float64           `json:"min_payment"`
	MaxPayment float64           `json:"max_payment"`
	TermMonths int               `json:"term_months"`
	CreditType domain.CreditType `json:"credit_type"`
}

func (h *CreditHandler) create(w http.ResponseWriter, r *http.Request) {
	var request createCreditRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	credit, err := h.service.Create(r.Context(), domain.Credit{
		ClientID:   request.ClientID,
		BankID:     request.BankID,
		MinPayment: request.MinPayment,
		MaxPayment: request.MaxPayment,
		TermMonths: request.TermMonths,
		CreditType: request.CreditType,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, credit)
}

func (h *CreditHandler) getByID(w http.ResponseWriter, r *http.Request) {
	credit, err := h.service.GetByID(r.Context(), chi.URLParam(r, "creditID"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, credit)
}

func (h *CreditHandler) list(w http.ResponseWriter, r *http.Request) {
	credits, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, credits)
}
