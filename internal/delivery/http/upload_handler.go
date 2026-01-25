package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"my-chi-app/internal/storage"
)

// JSONResponse writes a JSON response with the given status code and data
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// @Summary Get presigned upload URL
// @Description Generate a presigned S3 URl for uploading file/image to AWS S3
// @Tags uploads
// @Security Bearer
// @Param file_name query string true "File name (e.g., photo.jpg)"
// @Success 200 {object} map[string]string "presigned_url"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /uploads/presign [post]
func HandleGetPresignedUploadURL(s3Client *storage.S3Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())

		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		fileName := r.URL.Query().Get("file_name")

		if fileName == "" {
			BadRequest(w, "file_name query parameter is required")
			return
		}

		presignedURL, err := s3Client.CreatePresignedUploadURL(r.Context(), strconv.FormatInt(userID, 10)+"/"+fileName, 15*time.Minute)

		if err != nil {
			InternalError(w, "failed to generate presigned URL")
			return
		}

		resp := map[string]string{
			"presigned_url": presignedURL,
		}

		JSONResponse(w, http.StatusOK, resp)
	}
}
