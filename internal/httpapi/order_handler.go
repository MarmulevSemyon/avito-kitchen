package httpapi

import (
	"net/http"
	"strconv"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

func (h *Handler) createOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request createOrderRequest

	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(
			w,
			"INVALID_REQUEST",
			err.Error(),
		)
		return
	}

	items := make(
		[]service.CreateOrderItemInput,
		0,
		len(request.Items),
	)

	for _, item := range request.Items {
		items = append(
			items,
			service.CreateOrderItemInput{
				MenuItemID: item.MenuItemID,
				Quantity:   item.Quantity,
			},
		)
	}

	order, err := h.orderService.CreateOrder(
		r.Context(),
		service.CreateOrderInput{
			UserID:       request.UserID,
			RestaurantID: request.RestaurantID,
			Items:        items,
		},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		toOrderResponse(order),
	)
}

func (h *Handler) getOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	orderID, err := parsePositiveID(
		r.PathValue("order_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_ORDER_ID",
			err.Error(),
		)
		return
	}

	order, err := h.orderService.GetOrder(
		r.Context(),
		orderID,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toOrderResponse(order),
	)
}

func (h *Handler) getRestaurantOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	restaurantID, err := parsePositiveID(
		r.PathValue("restaurant_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_RESTAURANT_ID",
			err.Error(),
		)
		return
	}

	orderID, err := parsePositiveID(
		r.PathValue("order_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_ORDER_ID",
			err.Error(),
		)
		return
	}

	order, err := h.orderService.GetOrder(
		r.Context(),
		orderID,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Проверяем, что запрошенный заказ действительно
	// относится к ресторану из URL.
	if order.RestaurantID != restaurantID {
		writeServiceError(
			w,
			domain.ErrOrderRestaurantMismatch,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toOrderResponse(order),
	)
}

func (h *Handler) listRestaurantOrders(
	w http.ResponseWriter,
	r *http.Request,
) {
	restaurantID, err := parsePositiveID(
		r.PathValue("restaurant_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_RESTAURANT_ID",
			err.Error(),
		)
		return
	}

	limit := 0

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeBadRequest(
				w,
				"INVALID_LIMIT",
				"limit must be an integer",
			)
			return
		}

		limit = parsedLimit
	}

	var status *domain.OrderStatus

	if rawStatus := r.URL.Query().Get("status"); rawStatus != "" {
		value := domain.OrderStatus(rawStatus)
		status = &value
	}

	var cursor *service.OrderCursor

	if rawCursor := r.URL.Query().Get("cursor"); rawCursor != "" {
		cursor, err = decodeCursor(rawCursor)
		if err != nil {
			writeBadRequest(
				w,
				"INVALID_CURSOR",
				"invalid cursor",
			)
			return
		}
	}

	result, err := h.orderService.ListRestaurantOrders(
		r.Context(),
		service.ListRestaurantOrdersInput{
			RestaurantID: restaurantID,
			Status:       status,
			Limit:        limit,
			Cursor:       cursor,
		},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	orders := make(
		[]orderResponse,
		0,
		len(result.Orders),
	)

	for _, order := range result.Orders {
		orders = append(
			orders,
			toOrderResponse(order),
		)
	}

	var nextCursor *string

	if result.NextCursor != nil {
		encodedCursor, err := encodeCursor(
			result.NextCursor,
		)
		if err != nil {
			writeServiceError(w, err)
			return
		}

		nextCursor = &encodedCursor
	}

	writeJSON(
		w,
		http.StatusOK,
		listOrdersResponse{
			Orders:     orders,
			NextCursor: nextCursor,
		},
	)
}

func (h *Handler) updateOrderStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	restaurantID, err := parsePositiveID(
		r.PathValue("restaurant_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_RESTAURANT_ID",
			err.Error(),
		)
		return
	}

	orderID, err := parsePositiveID(
		r.PathValue("order_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_ORDER_ID",
			err.Error(),
		)
		return
	}

	var request updateOrderStatusRequest

	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(
			w,
			"INVALID_REQUEST",
			err.Error(),
		)
		return
	}

	order, err := h.orderService.UpdateStatus(
		r.Context(),
		service.UpdateOrderStatusInput{
			RestaurantID: restaurantID,
			OrderID:      orderID,
			Status:       request.Status,
		},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toOrderResponse(order),
	)
}
