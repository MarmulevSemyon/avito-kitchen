package httpapi

import (
	"net/http"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

func NewRouter(
	restaurantService *service.RestaurantService,
	orderService *service.OrderService,
) http.Handler {
	handler := NewHandler(
		restaurantService,
		orderService,
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /healthz",
		handler.health,
	)

	// User API.

	mux.HandleFunc(
		"GET /api/v1/restaurants",
		handler.listRestaurants,
	)

	mux.HandleFunc(
		"GET /api/v1/restaurants/{restaurant_id}/menu-items",
		handler.getMenu,
	)

	mux.HandleFunc(
		"POST /api/v1/orders",
		handler.createOrder,
	)

	mux.HandleFunc(
		"GET /api/v1/orders/{order_id}",
		handler.getOrder,
	)

	// Restaurant API.

	mux.HandleFunc(
		"POST /api/v1/restaurants/{restaurant_id}/menu-items",
		handler.createMenuItem,
	)

	mux.HandleFunc(
		"PATCH /api/v1/restaurants/{restaurant_id}/menu-items/{menu_item_id}",
		handler.updateMenuItem,
	)

	mux.HandleFunc(
		"GET /api/v1/restaurants/{restaurant_id}/orders",
		handler.listRestaurantOrders,
	)

	mux.HandleFunc(
		"GET /api/v1/restaurants/{restaurant_id}/orders/{order_id}",
		handler.getRestaurantOrder,
	)

	mux.HandleFunc(
		"PATCH /api/v1/restaurants/{restaurant_id}/orders/{order_id}/status",
		handler.updateOrderStatus,
	)

	return mux
}
