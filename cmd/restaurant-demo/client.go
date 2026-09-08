package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type apiClient struct {
	baseURL      string
	restaurantID int64
	client       *http.Client
}

type order struct {
	ID           int64       `json:"id"`
	UserID       int64       `json:"user_id"`
	RestaurantID int64       `json:"restaurant_id"`
	Status       string      `json:"status"`
	TotalPrice   int64       `json:"total_price"`
	Items        []orderItem `json:"items"`
}

type orderItem struct {
	MenuItemID int64  `json:"menu_item_id"`
	Name       string `json:"name"`
	UnitPrice  int64  `json:"unit_price"`
	Quantity   int    `json:"quantity"`
}

type listOrdersResponse struct {
	Orders []order `json:"orders"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func newAPIClient(
	baseURL string,
	restaurantID int64,
) *apiClient {
	return &apiClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		restaurantID: restaurantID,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *apiClient) listCreatedOrders(
	ctx context.Context,
) ([]order, error) {
	url := fmt.Sprintf(
		"%s/api/v1/restaurants/%d/orders?status=CREATED&limit=20",
		c.baseURL,
		c.restaurantID,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create list orders request: %w",
			err,
		)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"send list orders request: %w",
			err,
		)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, responseError(response)
	}

	var result listOrdersResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"decode list orders response: %w",
			err,
		)
	}

	return result.Orders, nil
}

func (c *apiClient) updateStatus(
	ctx context.Context,
	orderID int64,
	status string,
) error {
	body, err := json.Marshal(
		updateStatusRequest{
			Status: status,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"encode status request: %w",
			err,
		)
	}

	url := fmt.Sprintf(
		"%s/api/v1/restaurants/%d/orders/%d/status",
		c.baseURL,
		c.restaurantID,
		orderID,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf(
			"create status request: %w",
			err,
		)
	}

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf(
			"send status request: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return responseError(response)
	}

	return nil
}

func responseError(
	response *http.Response,
) error {
	body, err := io.ReadAll(
		io.LimitReader(
			response.Body,
			4096,
		),
	)
	if err != nil {
		return fmt.Errorf(
			"api returned status %d",
			response.StatusCode,
		)
	}

	return fmt.Errorf(
		"api returned status %d: %s",
		response.StatusCode,
		strings.TrimSpace(string(body)),
	)
}
