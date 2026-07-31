package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/wthareutalking/subscription-service/internal/model"
)

type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// Create добавляет новую подписку.
func (r *SubscriptionRepo) Create(sub model.Subscription) (model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var endDate interface{}
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	} else {
		endDate = nil
	}

	err := r.db.QueryRow(
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		endDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return model.Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}
	return sub, nil
}

// GetByID возвращает одну подписку по ID.
func (r *SubscriptionRepo) GetByID(id string) (model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	var sub model.Subscription
	var endDate sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Subscription{}, fmt.Errorf("subscription not found")
		}
		return model.Subscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	return sub, nil
}

// Update обновляет запись и возвращает обновлённый объект.
func (r *SubscriptionRepo) Update(sub model.Subscription) (model.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	var endDate interface{}
	if sub.EndDate != nil {
		endDate = *sub.EndDate
	} else {
		endDate = nil
	}

	var updated model.Subscription
	var endDateScan sql.NullTime
	err := r.db.QueryRow(
		query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		endDate,
		sub.ID,
	).Scan(
		&updated.ID,
		&updated.ServiceName,
		&updated.Price,
		&updated.UserID,
		&updated.StartDate,
		&endDateScan,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Subscription{}, fmt.Errorf("subscription not found")
		}
		return model.Subscription{}, fmt.Errorf("failed to update subscription: %w", err)
	}
	if endDateScan.Valid {
		updated.EndDate = &endDateScan.Time
	}
	return updated, nil
}

// Delete удаляет подписку. Возвращает ошибку, если запись не найдена.
func (r *SubscriptionRepo) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

// List возвращает список подписок с возможностью фильтрации по user_id и service_name.
func (r *SubscriptionRepo) List(userID, serviceName string) ([]model.Subscription, error) {
	query := "SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, serviceName)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime
		if err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

// SummaryResult содержит результат агрегации.
type SummaryResult struct {
	TotalCost int `json:"total_cost"`
}

// Summary возвращает суммарную стоимость подписок за период с фильтрами.
func (r *SubscriptionRepo) Summary(userID, serviceName string, from, to time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date >= $1 AND start_date <= $2`
	args := []interface{}{from, to}
	argIndex := 3

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, serviceName)
		argIndex++
	}

	var total int
	err := r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate summary: %w", err)
	}
	return total, nil
}
