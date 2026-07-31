package service

import (
	"fmt"
	"time"

	"github.com/wthareutalking/subscription-service/internal/model"
	"github.com/wthareutalking/subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepo
}

func NewSubscriptionService(repo *repository.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// CreateSubscriptionRequest описывает входные данные для создания/обновления.
type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (s *SubscriptionService) Create(req CreateSubscriptionRequest) (model.Subscription, error) {
	if err := validateRequest(req); err != nil {
		return model.Subscription{}, err
	}

	startDate, endDate, err := parseDates(req.StartDate, req.EndDate)
	if err != nil {
		return model.Subscription{}, err
	}

	sub := model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	return s.repo.Create(sub)
}

func (s *SubscriptionService) GetByID(id string) (model.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionService) Update(id string, req CreateSubscriptionRequest) (model.Subscription, error) {
	if err := validateRequest(req); err != nil {
		return model.Subscription{}, err
	}

	startDate, endDate, err := parseDates(req.StartDate, req.EndDate)
	if err != nil {
		return model.Subscription{}, err
	}

	sub := model.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	return s.repo.Update(sub)
}

func (s *SubscriptionService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *SubscriptionService) List(userID, serviceName string) ([]model.Subscription, error) {
	return s.repo.List(userID, serviceName)
}

func (s *SubscriptionService) Summary(userID, serviceName, fromStr, toStr string) (int, error) {
	from, err := parseMonthYear(fromStr)
	if err != nil {
		return 0, fmt.Errorf("invalid from date: %w", err)
	}
	to, err := parseMonthYear(toStr)
	if err != nil {
		return 0, fmt.Errorf("invalid to date: %w", err)
	}
	to = time.Date(to.Year(), to.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	return s.repo.Summary(userID, serviceName, from, to)
}

func validateRequest(req CreateSubscriptionRequest) error {
	if req.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func parseDates(startStr string, endStr *string) (time.Time, *time.Time, error) {
	startDate, err := parseMonthYear(startStr)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("invalid start_date: %w", err)
	}
	var endDate *time.Time
	if endStr != nil && *endStr != "" {
		parsed, err := parseMonthYear(*endStr)
		if err != nil {
			return time.Time{}, nil, fmt.Errorf("invalid end_date: %w", err)
		}
		endDate = &parsed
	}
	return startDate, endDate, nil
}

func parseMonthYear(s string) (time.Time, error) {
	if len(s) != 7 {
		return time.Time{}, fmt.Errorf("expected format MM-YYYY")
	}
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
