package handler

import (
	"encoding/json"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
)

func (h *handler) createCapsule(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	var input domain.CreateCapsuleDTO
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		handleRequestError(w, err)
		return
	}

	capsule, err := h.svc.CreateCapsule(r.Context(), userID, input)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	newJSONResponse(w, capsule)
	return
}
