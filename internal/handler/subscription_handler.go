package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wthareutalking/subscription-service/internal/service"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// Create godoc
// @Summary      Создать подписку
// @Description  Создаёт новую запись о подписке
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription body service.CreateSubscriptionRequest true "Данные подписки"
// @Success      201  {object}  model.Subscription
// @Failure      400  {string}  string
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.svc.Create(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// GetByID godoc
// @Summary      Получить подписку по ID
// @Description  Возвращает одну запись о подписке по её UUID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID подписки"
// @Success      200  {object}  model.Subscription
// @Failure      404  {string}  string
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sub, err := h.svc.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

// Update godoc
// @Summary      Обновить подписку
// @Description  Полностью обновляет запись о подписке по её ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID подписки"
// @Param        subscription body service.CreateSubscriptionRequest true "Новые данные подписки"
// @Success      200  {object}  model.Subscription
// @Failure      400  {string}  string
// @Failure      404  {string}  string
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req service.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.svc.Update(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

// Delete godoc
// @Summary      Удалить подписку
// @Description  Удаляет запись о подписке по её ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID подписки"
// @Success      204  {string}  string
// @Failure      404  {string}  string
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      Список подписок
// @Description  Возвращает все подписки с возможностью фильтрации по user_id и service_name
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id query string false "UUID пользователя"
// @Param        service_name query string false "Название сервиса"
// @Success      200  {array}   model.Subscription
// @Failure      500  {string}  string
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	subs, err := h.svc.List(userID, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

// Summary godoc
// @Summary      Суммарная стоимость
// @Description  Возвращает суммарную стоимость всех подписок за выбранный период с возможностью фильтрации
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        from query string true "Начало периода (MM-YYYY)"
// @Param        to query string true "Конец периода (MM-YYYY)"
// @Param        user_id query string false "UUID пользователя"
// @Param        service_name query string false "Название сервиса"
// @Success      200  {object}  map[string]int
// @Failure      400  {string}  string
// @Router       /subscriptions/summary [get]
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		http.Error(w, "from and to query parameters are required", http.StatusBadRequest)
		return
	}

	total, err := h.svc.Summary(userID, serviceName, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"total_cost": total})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
