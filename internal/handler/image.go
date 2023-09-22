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

// RemoveImage | Removes Image
//
//	@Summary      RemoveImage
//	@Security     ApiKeyAuth
//	@Description  Removes an image
//	@Tags         Images
//	@Produce      json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Param        imageID      path      string true "imageID"
//	@Success      204
//	@Failure      401          {object}  errorResponse
//	@Failure      403          {object}  errorResponse
//	@Failure      404          {object}  errorResponse
//	@Failure      500          {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID}/images/{imageID} [delete]
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

	imageID, err := parseObjectIDFromParam(params, pathImageID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	if err = h.svc.RemoveImage(r.Context(), userID, capsuleID, imageID.Hex()); err != nil {
		log.Println(err)
		newErrorResponse(w, err)
		return
	}

	if err = h.storage.Delete(r.Context(), imageID.Hex()); err != nil {
		log.Println(err)
		newErrorResponse(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

// GetImage | Retrieves The Image
//
//	@Summary      GetImage
//	@Security     ApiKeyAuth
//	@Description  Retrieves the image
//	@Tags         Images
//	@Produce      image/png, image/jpeg, application/json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Param        imageID      path      string true "imageID"
//	@Success      200          {object}  domain.File
//	@Failure      401   	   {object}  errorResponse
//	@Failure      403   	   {object}  errorResponse
//	@Failure      404   	   {object}  errorResponse
//	@Failure      500   	   {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID}/images/{imageID} [get]
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

	imageID, err := parseObjectIDFromParam(params, pathImageID)
	if err != nil {
		newErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	if _, err = h.svc.GetCapsuleByID(r.Context(), userID, capsuleID); err != nil {
		newErrorResponse(w, err)
		return
	}

	file, err := h.storage.Get(r.Context(), imageID.Hex())
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

// AddImage | Adds An Image To The Capsule
//
//	@Summary      AddImage
//	@Security     ApiKeyAuth
//	@Description  Adds an image to the capsule
//	@Tags         Images
//	@Accept       mpfd
//	@Produce      json
//	@Param        capsuleID    path      string true "capsuleID"
//	@Param        image        formData  file true "image"
//	@Success      201          {object}  domain.File
//	@Failure      400          {object}  errorResponse
//	@Failure      401   	   {object}  errorResponse
//	@Failure      500          {object}  errorResponse
//	@Router       /api/v1/capsules/{capsuleID}/images [post]
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
		newErrorResponse(w, errors.New("unable to parse the form"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes := make([]byte, header.Size)
	if _, err = file.Read(fileBytes); err != nil {
		log.Println("addCapsuleImage", err)
		newErrorResponse(w, errors.New("failed to read the uploaded file"), http.StatusBadRequest)
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
