package repository

import (
	"context"
	"database/sql"
	"errors"

	"my-chi-app/internal/domain/entity"
)

// NotificationRepository manages notifications.
type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

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

func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*entity.Notification, error) {
	const q = `
        SELECT notification_id, owner_id, actor_id, component_type, component_id, notification_type, status, created_at
        FROM notifications
        WHERE notification_id = $1
    `
	row := r.db.QueryRowContext(ctx, q, id)
	return scanNotification(row)
}

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

type notificationRowScanner interface {
	Scan(dest ...any) error
}

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
