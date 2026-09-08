package httpapi

import (
	"errors"
	"net/http"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, domain.ErrInvalidUserID):
		writeBadRequest(
			w,
			"INVALID_USER_ID",
			domain.ErrInvalidUserID.Error(),
		)

	case errors.Is(err, domain.ErrInvalidRestaurantID):
		writeBadRequest(
			w,
			"INVALID_RESTAURANT_ID",
			domain.ErrInvalidRestaurantID.Error(),
		)

	case errors.Is(err, domain.ErrInvalidMenuItemID):
		writeBadRequest(
			w,
			"INVALID_MENU_ITEM_ID",
			domain.ErrInvalidMenuItemID.Error(),
		)

	case errors.Is(err, domain.ErrInvalidOrderID):
		writeBadRequest(
			w,
			"INVALID_ORDER_ID",
			domain.ErrInvalidOrderID.Error(),
		)

	case errors.Is(err, domain.ErrInvalidLimit):
		writeBadRequest(
			w,
			"INVALID_LIMIT",
			domain.ErrInvalidLimit.Error(),
		)

	case errors.Is(err, domain.ErrInvalidOrderStatus):
		writeBadRequest(
			w,
			"INVALID_ORDER_STATUS",
			domain.ErrInvalidOrderStatus.Error(),
		)

	case errors.Is(err, domain.ErrInvalidMenuItemName):
		writeBadRequest(
			w,
			"INVALID_MENU_ITEM_NAME",
			domain.ErrInvalidMenuItemName.Error(),
		)

	case errors.Is(err, domain.ErrInvalidMenuItemPrice):
		writeBadRequest(
			w,
			"INVALID_MENU_ITEM_PRICE",
			domain.ErrInvalidMenuItemPrice.Error(),
		)

	case errors.Is(err, domain.ErrEmptyMenuItemUpdate):
		writeBadRequest(
			w,
			"EMPTY_MENU_ITEM_UPDATE",
			domain.ErrEmptyMenuItemUpdate.Error(),
		)

	case errors.Is(err, domain.ErrEmptyOrder):
		writeBadRequest(
			w,
			"EMPTY_ORDER",
			domain.ErrEmptyOrder.Error(),
		)

	case errors.Is(err, domain.ErrInvalidOrderItemQuantity):
		writeBadRequest(
			w,
			"INVALID_QUANTITY",
			domain.ErrInvalidOrderItemQuantity.Error(),
		)

	case errors.Is(err, domain.ErrDuplicateMenuItem):
		writeBadRequest(
			w,
			"DUPLICATE_MENU_ITEM",
			domain.ErrDuplicateMenuItem.Error(),
		)

	case errors.Is(err, domain.ErrMenuItemRestaurantMismatch):
		writeBadRequest(
			w,
			"MENU_ITEM_RESTAURANT_MISMATCH",
			domain.ErrMenuItemRestaurantMismatch.Error(),
		)

	case errors.Is(err, domain.ErrRestaurantNotFound):
		writeAPIError(
			w,
			http.StatusNotFound,
			"RESTAURANT_NOT_FOUND",
			domain.ErrRestaurantNotFound.Error(),
		)

	case errors.Is(err, domain.ErrMenuItemNotFound):
		writeAPIError(
			w,
			http.StatusNotFound,
			"MENU_ITEM_NOT_FOUND",
			domain.ErrMenuItemNotFound.Error(),
		)

	case errors.Is(err, domain.ErrOrderNotFound):
		writeAPIError(
			w,
			http.StatusNotFound,
			"ORDER_NOT_FOUND",
			domain.ErrOrderNotFound.Error(),
		)

	case errors.Is(err, domain.ErrRestaurantUnavailable):
		writeAPIError(
			w,
			http.StatusConflict,
			"RESTAURANT_UNAVAILABLE",
			domain.ErrRestaurantUnavailable.Error(),
		)

	case errors.Is(err, domain.ErrMenuItemUnavailable):
		writeAPIError(
			w,
			http.StatusConflict,
			"MENU_ITEM_UNAVAILABLE",
			domain.ErrMenuItemUnavailable.Error(),
		)

	case errors.Is(err, domain.ErrOrderRestaurantMismatch):
		writeAPIError(
			w,
			http.StatusConflict,
			"ORDER_RESTAURANT_MISMATCH",
			domain.ErrOrderRestaurantMismatch.Error(),
		)

	case errors.Is(err, domain.ErrInvalidOrderStatusTransition):
		writeAPIError(
			w,
			http.StatusConflict,
			"INVALID_ORDER_STATUS_TRANSITION",
			domain.ErrInvalidOrderStatusTransition.Error(),
		)

	default:
		writeAPIError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
	}
}

func writeBadRequest(
	w http.ResponseWriter,
	code string,
	message string,
) {
	writeAPIError(
		w,
		http.StatusBadRequest,
		code,
		message,
	)
}

func writeAPIError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	writeJSON(
		w,
		status,
		errorResponse{
			Error: apiError{
				Code:    code,
				Message: message,
			},
		},
	)
}
