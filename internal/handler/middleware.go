package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

const (
	userCtx = "userID"

	requestRateTimeout = 1 * time.Second
	requestRateLimit   = 20
)

var limiter = rate.NewLimiter(rate.Every(requestRateTimeout), requestRateLimit)

func (h *handler) RateLimiter(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if !limiter.Allow() {
			newErrorResponse(w, errors.New("rate limit exceeded"), http.StatusTooManyRequests)
			return
		}

		next(w, r, params)
	}
}

func (h *handler) JWTAuthentication(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		header := r.Header.Get("Authorization")

		if header == "" {
			newErrorResponse(w, errors.New("empty auth header"), http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			newErrorResponse(w, errors.New("invalid auth header"), http.StatusUnauthorized)
			return
		}

		if len(headerParts[1]) == 0 {
			newErrorResponse(w, errors.New("token is empty"), http.StatusUnauthorized)
			return
		}

		claims, err := h.svc.ParseToken(headerParts[1])
		if err != nil {
			newErrorResponse(w, err)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			newErrorResponse(w, errors.New("invalid token"), http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), userCtx, userID)), params)
	}
}

func getUserID(r *http.Request) (primitive.ObjectID, error) {
	id, ok := r.Context().Value(userCtx).(string)
	if !ok {
		log.Println("getUserID", "failed to convert", r.Context().Value(userCtx))
		return primitive.NilObjectID, errors.New("internal server error")
	}

	if id == "" {
		log.Println("getUserID", "empty userId in context")
		return primitive.NilObjectID, errors.New("internal server error")
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("getUserID", err)
		return primitive.NilObjectID, errors.New("internal server error")
	}

	return oid, nil
}
