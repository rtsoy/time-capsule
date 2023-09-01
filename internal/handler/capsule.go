package handler

import (
	"encoding/json"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
)

func (h *handler) deleteCapsule(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	capsuleID, err := parseObjectIDFromParam(params, pathCapsuleID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	if err = h.svc.DeleteCapsule(r.Context(), userID, capsuleID); err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (h *handler) updateCapsule(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	capsuleID, err := parseObjectIDFromParam(params, pathCapsuleID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	var input domain.UpdateCapsuleDTO
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		handleRequestError(w, err)
		return
	}

	if err = h.svc.UpdateCapsule(r.Context(), userID, capsuleID, input); err != nil {
		newErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (h *handler) getCapsuleByID(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	capsuleID, err := parseObjectIDFromParam(params, pathCapsuleID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	capsule, err := h.svc.GetCapsuleByID(r.Context(), userID, capsuleID)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	newJSONResponse(w, capsule)
	return
}

func (h *handler) getCapsules(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	capsules, err := h.svc.GetAllCapsules(r.Context(), userID)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	newJSONResponse(w, capsules)
	return
}

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

	newJSONResponse(w, capsule, http.StatusCreated)
	return
}
