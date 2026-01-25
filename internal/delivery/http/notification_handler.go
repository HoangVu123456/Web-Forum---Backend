package http

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"my-chi-app/internal/database/repository"
)

// NotificationResponse is the payload response when returning notification information
type NotificationResponse struct {
	NotificationID    int64  `json:"notification_id"`
	ActorID           int64  `json:"actor_id"`
	ComponentInvolved string `json:"component_involved"`
	PostID            *int64 `json:"post_id"`
	CommentID         *int64 `json:"comment_id"`
	NotificationType  string `json:"notification_type"`
	Status            bool   `json:"status"`
}

// @Summary Get all notifications
// @Description Fetch paginated notifications of the authenticated user
// @Tags notifications
// @Security Bearer
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} NotificationResponse
// @Failure 401 {object} map[string]string
// @Router /notifications [get]
func HandleGetAllUserNotifications(notificationRepo *repository.NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		// Pagination
		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		list, err := notificationRepo.ListByOwner(r.Context(), userID, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		resp := make([]NotificationResponse, 0, len(list))
		for _, n := range list {
			var postID, commentID *int64
			switch n.ComponentType {
			case "post":
				postID = &n.ComponentID
			case "comment":
				commentID = &n.ComponentID
			}
			resp = append(resp, NotificationResponse{
				NotificationID:    n.ID,
				ActorID:           n.ActorID,
				ComponentInvolved: n.ComponentType,
				PostID:            postID,
				CommentID:         commentID,
				NotificationType:  n.NotificationType,
				Status:            n.Status,
			})
		}

		Success(w, resp)
	}
}

// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Security Bearer
// @Param notification_id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /notifications/{notification_id}/read [post]
func HandleMarkNotificationAsRead(notificationRepo *repository.NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "notification_id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid notification_id")
			return
		}

		n, err := notificationRepo.GetByID(r.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "notification not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}
		if n.OwnerID != userID {
			Forbidden(w, "cannot modify this notification")
			return
		}

		if err := notificationRepo.MarkRead(r.Context(), id); err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "notification not found")
				return
			}
			InternalError(w, err.Error())
			return
		}

		Success(w, map[string]string{"message": "Notification read!"})
	}
}

// @Summary Mark notification as unread
// @Description Mark a specific notification as unread
// @Tags notifications
// @Security Bearer
// @Param notification_id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /notifications/{notification_id}/unread [post]
func HandleMarkNotificationAsUnread(notificationRepo *repository.NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "notification_id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(w, "invalid notification_id")
			return
		}

		n, err := notificationRepo.GetByID(r.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "notification not found")
			} else {
				InternalError(w, err.Error())
			}
			return
		}
		if n.OwnerID != userID {
			Forbidden(w, "cannot modify this notification")
			return
		}

		if err := notificationRepo.MarkUnread(r.Context(), id); err != nil {
			if err == sql.ErrNoRows {
				NotFound(w, "notification not found")
				return
			}
			InternalError(w, err.Error())
			return
		}

		Success(w, map[string]string{"message": "Notification unread!"})
	}
}

// @Summary Get read notifications
// @Description Fetch paginated all notifications that are read of the authenticated user
// @Tags notifications
// @Security Bearer
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} NotificationResponse
// @Failure 401 {object} map[string]string
// @Router /notifications/read [get]
func HandleGetAllReadNotifications(notificationRepo *repository.NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		list, err := notificationRepo.ListByOwnerAndStatus(r.Context(), userID, true, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		resp := make([]NotificationResponse, 0, len(list))
		for _, n := range list {
			var postID, commentID *int64
			switch n.ComponentType {
			case "post":
				postID = &n.ComponentID
			case "comment":
				commentID = &n.ComponentID
			}
			resp = append(resp, NotificationResponse{
				NotificationID:    n.ID,
				ActorID:           n.ActorID,
				ComponentInvolved: n.ComponentType,
				PostID:            postID,
				CommentID:         commentID,
				NotificationType:  n.NotificationType,
				Status:            n.Status,
			})
		}

		Success(w, resp)
	}
}

// @Summary Get unread notifications
// @Description Fetch paginated all notifications that are unread of the authenticated user
// @Tags notifications
// @Security Bearer
// @Param limit query int false "Limit" default(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} NotificationResponse
// @Failure 401 {object} map[string]string
// @Router /notifications/unread [get]
func HandleGetAllUnreadNotifications(notificationRepo *repository.NotificationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			Unauthorized(w, "unauthorized")
			return
		}

		limit, offset := int32(1000), int32(0)
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.ParseInt(l, 10, 32); err == nil {
				limit = int32(v)
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if v, err := strconv.ParseInt(o, 10, 32); err == nil {
				offset = int32(v)
			}
		}

		list, err := notificationRepo.ListByOwnerAndStatus(r.Context(), userID, false, limit, offset)
		if err != nil {
			InternalError(w, err.Error())
			return
		}

		resp := make([]NotificationResponse, 0, len(list))
		for _, n := range list {
			var postID, commentID *int64
			switch n.ComponentType {
			case "post":
				postID = &n.ComponentID
			case "comment":
				commentID = &n.ComponentID
			}
			resp = append(resp, NotificationResponse{
				NotificationID:    n.ID,
				ActorID:           n.ActorID,
				ComponentInvolved: n.ComponentType,
				PostID:            postID,
				CommentID:         commentID,
				NotificationType:  n.NotificationType,
				Status:            n.Status,
			})
		}

		Success(w, resp)
	}
}
