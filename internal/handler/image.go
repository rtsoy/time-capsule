package handler

import (
	"errors"
	"log"
	"net/http"

	"time-capsule/internal/domain"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxUploadSize = 5 << 20 // 5 megabytes

var fileTypes = map[string]interface{}{
	"image/jpeg": nil,
	"image/png":  nil,
}

func (h *handler) removeCapsuleImage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	imageID := params.ByName(pathImageID)

	if err = h.storage.Delete(r.Context(), imageID); err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("not found"), http.StatusNotFound)
		return
	}

	if err = h.svc.RemoveImage(r.Context(), userID, capsuleID, imageID); err != nil {
		log.Println(err)
		newErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (h *handler) getCapsuleImage(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	if _, err = h.svc.GetCapsuleByID(r.Context(), userID, capsuleID); err != nil {
		newErrorResponse(w, err)
		return
	}

	imageID := params.ByName(pathImageID)

	file, err := h.storage.Get(r.Context(), imageID)
	if err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("not found"), http.StatusNotFound)
		return
	}

	contentType := http.DetectContentType(file.Bytes)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	w.Write(file.Bytes)
	return
}

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

	fileType := http.DetectContentType(fileBytes)
	if _, ok := fileTypes[fileType]; !ok {
		newErrorResponse(w, errors.New("invalid file type"), http.StatusBadRequest)
		return
	}

	input := domain.File{
		Bytes: fileBytes,
		Name:  primitive.NewObjectID().Hex(),
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
