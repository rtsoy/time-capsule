package handler

import (
	"encoding/json"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
)

type tokenResponse struct {
	Token string `json:"token"`
}

// SignIn | Login
//
//	@Summary      SignIn
//	@Description  Log in
//	@Tags         Auth
//	@Accept       json
//	@Produce      json
//	@Param        input body      domain.LogInUserDTO true "Input"
//	@Success      200   {object}  tokenResponse
//	@Failure      400   {object}  errorResponse
//	@Failure      401   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/sign-in [post]
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

	newJSONResponse(w, tokenResponse{Token: token})
	return
}

// SignUp | Creates New Account
//
//	@Summary      SignUp
//	@Description  Creates a new account
//	@Tags         Auth
//	@Accept       json
//	@Produce      json
//	@Param        input body      domain.CreateUserDTO true "Input"
//	@Success      201   {object}  domain.User
//	@Failure      400   {object}  errorResponse
//	@Failure      409   {object}  errorResponse
//	@Failure      500   {object}  errorResponse
//	@Router       /api/v1/sign-up [post]
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
