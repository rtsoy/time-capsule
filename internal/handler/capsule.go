package handler

import (
	"encoding/json"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
)

// DeleteCapsule | Removes Capsule
//
//	@Summary      GetCapsule
//	@Security     ApiKeyAuth
//	@Description  Removes capsule
//	@Tags         Capsules
//	@Produce      json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Success      204
//	@Failure      401   {object}  errorResponse
//	@Failure      403   {object}  errorResponse
//	@Failure      404   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID} [delete]
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
		newErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

// UpdateCapsule | Updates Capsule
//
//	@Summary      GetCapsule
//	@Security     ApiKeyAuth
//	@Description  Updates capsule
//	@Tags         Capsules
//	@Produce      json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Param        input body      domain.UpdateCapsuleDTO true "input"
//	@Success      204
//	@Failure      400   {object}  errorResponse
//	@Failure      401   {object}  errorResponse
//	@Failure      403   {object}  errorResponse
//	@Failure      404   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID} [patch]
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

// GetCapsule | Retrieves Capsule By ID
//
//	@Summary      GetCapsule
//	@Security     ApiKeyAuth
//	@Description  Retrieves capsule by ID
//	@Tags         Capsules
//	@Produce      json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Success      200   {object}  domain.Capsule
//	@Failure      401   {object}  errorResponse
//	@Failure      403   {object}  errorResponse
//	@Failure      404   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID} [get]
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

// GetCapsules | Retrieves All Capsules
//
//	@Summary      GetCapsules
//	@Security     ApiKeyAuth
//	@Description  Retrieves all capsules
//	@Tags         Capsules
//	@Produce      json
//	@Success      200   {array}   domain.Capsule
//	@Failure      401   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/capsules [get]
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

// CreateCapsule | Creates New Capsule
//
//	@Summary      CreateCapsule
//	@Security     ApiKeyAuth
//	@Description  Creates new capsule
//	@Tags         Capsules
//	@Accept       json
//	@Produce      json
//	@Param        input body      domain.CreateCapsuleDTO true "input"
//	@Success      201   {object}  domain.Capsule
//	@Failure      400   {object}  errorResponse
//	@Failure      401   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/capsules [post]
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
