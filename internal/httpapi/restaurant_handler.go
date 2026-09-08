package httpapi

import (
	"net/http"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

func (h *Handler) listRestaurants(
	w http.ResponseWriter,
	r *http.Request,
) {
	restaurants, err := h.restaurantService.ListRestaurants(
		r.Context(),
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := make(
		[]restaurantResponse,
		0,
		len(restaurants),
	)

	for _, restaurant := range restaurants {
		response = append(
			response,
			restaurantResponse{
				ID:          restaurant.ID,
				Name:        restaurant.Name,
				Description: restaurant.Description,
			},
		)
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"restaurants": response,
		},
	)
}

func (h *Handler) getMenu(
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

	menuItems, err := h.restaurantService.GetMenu(
		r.Context(),
		restaurantID,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := make(
		[]menuItemResponse,
		0,
		len(menuItems),
	)

	for _, menuItem := range menuItems {
		response = append(
			response,
			toMenuItemResponse(menuItem),
		)
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"restaurant_id": restaurantID,
			"items":         response,
		},
	)
}

func (h *Handler) createMenuItem(
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

	var request createMenuItemRequest

	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(
			w,
			"INVALID_REQUEST",
			err.Error(),
		)
		return
	}

	menuItem, err := h.restaurantService.CreateMenuItem(
		r.Context(),
		service.CreateMenuItemInput{
			RestaurantID: restaurantID,
			Name:         request.Name,
			Description:  request.Description,
			Price:        request.Price,
		},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		toMenuItemResponse(menuItem),
	)
}

func (h *Handler) updateMenuItem(
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

	menuItemID, err := parsePositiveID(
		r.PathValue("menu_item_id"),
	)
	if err != nil {
		writeBadRequest(
			w,
			"INVALID_MENU_ITEM_ID",
			err.Error(),
		)
		return
	}

	var request updateMenuItemRequest

	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(
			w,
			"INVALID_REQUEST",
			err.Error(),
		)
		return
	}

	menuItem, err := h.restaurantService.UpdateMenuItem(
		r.Context(),
		service.UpdateMenuItemInput{
			RestaurantID: restaurantID,
			MenuItemID:   menuItemID,
			Name:         request.Name,
			Description:  request.Description,
			Price:        request.Price,
			IsAvailable:  request.IsAvailable,
		},
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toMenuItemResponse(menuItem),
	)
}
