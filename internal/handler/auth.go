package handler

import (
	"encoding/json"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
)

func (h *handler) signIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input domain.LogInUserDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handleRequestError(w, err)
		return
	}

	token, err := h.svc.GenerateToken(r.Context(), input.Email, input.Password)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	newJSONResponse(w, map[string]string{
		"token": token,
	})
	return
}

func (h *handler) signUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input domain.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		handleRequestError(w, err)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), input)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	newJSONResponse(w, user, http.StatusCreated)
	return
}
