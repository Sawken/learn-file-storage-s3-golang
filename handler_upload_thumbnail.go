package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}

	fmt.Printf("MEDIA_TYPE: %s\n", mediaType)
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusUnauthorized, errors.New("File format unauthorized").Error(), err)
		return
	}

	splittedMediaType := strings.Split(mediaType, "/")
	if len(splittedMediaType) != 2 {
		respondWithError(w, http.StatusUnprocessableEntity, errors.New("Unable to process entity").Error(), errors.New(mediaType))
		return
	}
	fileExtension := splittedMediaType[1]

	//var fileData []byte
	//fileData, err = io.ReadAll(file)
	//if err != nil {
	//	respondWithError(w, http.StatusUnprocessableEntity, err.Error(), err)
	//	return
	//}
	//fileBase64 := base64.StdEncoding.EncodeToString(fileData)
	videoRow, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	if videoRow.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	//fmt.Printf("FILE_DATA: %+v", fileData)
	//videoThumbnails[videoRow.ID] = thumbnail{
	//	data:      fileData,
	//	mediaType: mediaType,
	//}

	fd, err := os.Create(filepath.Join(cfg.assetsRoot, videoIDString) + "." + fileExtension)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	defer fd.Close()

	_, err = io.Copy(fd, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	//newEncodedThumbnailUrl := "data:" + mediaType + ";base64," + fileBase64
	newThumbnailUrl := "http://localhost:" + filepath.Join(cfg.port, strings.Split(cfg.assetsRoot, "/")[1], videoIDString) + "." + fileExtension
	fmt.Printf("THUMBNAIL_URL: [%s]\n", newThumbnailUrl)
	videoRow.ThumbnailURL = &newThumbnailUrl
	err = cfg.db.UpdateVideo(videoRow)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error(), err)
		return
	}
	//fmt.Printf("THUMBNAIL: %+v\n", videoThumbnails[videoRow.ID])
	respondWithJSON(w, http.StatusOK, &videoRow)
}
