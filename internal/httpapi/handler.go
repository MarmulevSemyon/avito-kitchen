package httpapi

import (
	"net/http"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

type Handler struct {
	restaurantService *service.RestaurantService
	orderService      *service.OrderService
}

func NewHandler(
	restaurantService *service.RestaurantService,
	orderService *service.OrderService,
) *Handler {
	return &Handler{
		restaurantService: restaurantService,
		orderService:      orderService,
	}
}

func (h *Handler) health(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status": "ok",
		},
	)
}
