package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// NotificationRepository manages notifications
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a new notification into the database
func (r *NotificationRepository) Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error) {
	const q = `
        INSERT INTO notifications (owner_id, actor_id, component_type, component_id, notification_type, status)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING notification_id, created_at, status
    `
	err := r.db.QueryRowContext(ctx, q, n.OwnerID, n.ActorID, n.ComponentType, n.ComponentID, n.NotificationType, n.Status).
		Scan(&n.ID, &n.CreatedAt, &n.Status)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// GetByID retrieves a notification by its ID
func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*entity.Notification, error) {
	const q = `
        SELECT notification_id, owner_id, actor_id, component_type, component_id, notification_type, status, created_at
        FROM notifications
        WHERE notification_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanNotification(row)
}

// ListByOwner returns notifications for a specific user
func (r *NotificationRepository) ListByOwner(ctx context.Context, ownerID int64, limit, offset int32) ([]*entity.Notification, error) {
	const q = `
        SELECT notification_id, owner_id, actor_id, component_type, component_id, notification_type, status, created_at
        FROM notifications
        WHERE owner_id = $1
        ORDER BY notification_id DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// ListByOwnerAndStatus returns notifications for a user filtered by read or unread status
func (r *NotificationRepository) ListByOwnerAndStatus(ctx context.Context, ownerID int64, status bool, limit, offset int32) ([]*entity.Notification, error) {
	const q = `
				SELECT notification_id, owner_id, actor_id, component_type, component_id, notification_type, status, created_at
				FROM notifications
				WHERE owner_id = $1 AND status = $2
				ORDER BY notification_id DESC
				LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, q, ownerID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// MarkRead marks a notification as read
func (r *NotificationRepository) MarkRead(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE notifications SET status = TRUE WHERE notification_id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// MarkUnread marks a notification as unread
func (r *NotificationRepository) MarkUnread(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE notifications SET status = FALSE WHERE notification_id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// notificationRowScanner defines the interface for scanning notification rows
type notificationRowScanner interface {
	Scan(dest ...any) error
}

// scanNotification scans a notification from the given row scanner
func scanNotification(rs notificationRowScanner) (*entity.Notification, error) {
	var n entity.Notification
	if err := rs.Scan(&n.ID, &n.OwnerID, &n.ActorID, &n.ComponentType, &n.ComponentID, &n.NotificationType, &n.Status, &n.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &n, nil
}
