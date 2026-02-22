package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}

	videoType := header.Header.Get("Content-Type")
	fmt.Printf("MEDIA_TYPE: %s\n", videoType)
	var fileData []byte
	fileData, err = io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}
	videoIdUUID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}
	videoRow, err := cfg.db.GetVideo(videoIdUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	if videoRow.UserID.String() != userID.String() {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	fmt.Printf("FILE_DATA: %+v", fileData)
	videoThumbnails[videoRow.ID] = thumbnail{
		data:      fileData,
		mediaType: videoType,
	}
	*videoRow.ThumbnailURL = "/api/thumbnail/" + videoIDString
	err = cfg.db.UpdateVideo(videoRow)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	fmt.Printf("THUMBNAIL: %+v\n", videoThumbnails[videoRow.ID])
	respondWithJSON(w, http.StatusOK, &videoRow)
}
