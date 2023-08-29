package handler

import (
	"errors"
	"log"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const maxUploadSize = 5 << 20 // 5 megabytes

func (h *handler) addCapsuleImage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	capsuleID, err := parseObjectIDFromParam(params, pathCapsuleID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		newErrorResponse(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("unable to parse the form"))
		return
	}
	defer file.Close()

	fileBytes := make([]byte, header.Size)
	if _, err = file.Read(fileBytes); err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("failed to read the uploaded file"))
		return
	}

	input := domain.File{
		Bytes: fileBytes,
		Name:  uuid.New().String(),
		Size:  header.Size,
	}

	if err = h.storage.Upload(r.Context(), input); err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("internal server error"))
		return
	}

	if err = h.svc.AddImage(r.Context(), userID, capsuleID, input.Name); err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("internal server error"))
		return
	}

	newJSONResponse(w, input, http.StatusCreated)
	return
}
