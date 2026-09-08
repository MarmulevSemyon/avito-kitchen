package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

// RestaurantCatalogRepository contains restaurant operations required
// by RestaurantService.
//
// The interface is intentionally separate from RestaurantRepository
// used by OrderService.
type RestaurantCatalogRepository interface {
	ListActive(ctx context.Context) ([]domain.Restaurant, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (domain.Restaurant, error)
}

// RestaurantMenuRepository contains menu operations required
// by RestaurantService.
type RestaurantMenuRepository interface {
	ListByRestaurantID(
		ctx context.Context,
		restaurantID int64,
	) ([]domain.MenuItem, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (domain.MenuItem, error)

	Create(
		ctx context.Context,
		item domain.MenuItem,
	) (domain.MenuItem, error)

	Update(
		ctx context.Context,
		item domain.MenuItem,
	) (domain.MenuItem, error)
}

type RestaurantService struct {
	restaurants RestaurantCatalogRepository
	menuItems   RestaurantMenuRepository
}

func NewRestaurantService(
	restaurants RestaurantCatalogRepository,
	menuItems RestaurantMenuRepository,
) *RestaurantService {
	return &RestaurantService{
		restaurants: restaurants,
		menuItems:   menuItems,
	}
}

type CreateMenuItemInput struct {
	RestaurantID int64
	Name         string
	Description  string
	Price        int64
}

type UpdateMenuItemInput struct {
	RestaurantID int64
	MenuItemID   int64

	Name        *string
	Description *string
	Price       *int64
	IsAvailable *bool
}

func (s *RestaurantService) ListRestaurants(
	ctx context.Context,
) ([]domain.Restaurant, error) {
	restaurants, err := s.restaurants.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active restaurants: %w", err)
	}

	return restaurants, nil
}

func (s *RestaurantService) GetMenu(
	ctx context.Context,
	restaurantID int64,
) ([]domain.MenuItem, error) {
	if restaurantID <= 0 {
		return nil, domain.ErrInvalidRestaurantID
	}

	restaurant, err := s.restaurants.GetByID(
		ctx,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("get restaurant: %w", err)
	}

	if !restaurant.IsActive {
		return nil, domain.ErrRestaurantUnavailable
	}

	menuItems, err := s.menuItems.ListByRestaurantID(
		ctx,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list restaurant menu: %w", err)
	}

	return menuItems, nil
}

func (s *RestaurantService) CreateMenuItem(
	ctx context.Context,
	input CreateMenuItemInput,
) (domain.MenuItem, error) {
	if input.RestaurantID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidRestaurantID
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemName
	}

	if input.Price < 0 {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemPrice
	}

	// Проверяем существование ресторана.
	//
	// IsActive здесь намеренно не проверяем:
	// временно выключенный ресторан всё ещё может редактировать меню.
	_, err := s.restaurants.GetByID(
		ctx,
		input.RestaurantID,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf("get restaurant: %w", err)
	}

	menuItem := domain.MenuItem{
		RestaurantID: input.RestaurantID,
		Name:         name,
		Description:  input.Description,
		Price:        input.Price,
		IsAvailable:  true,
	}

	createdMenuItem, err := s.menuItems.Create(
		ctx,
		menuItem,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf("create menu item: %w", err)
	}

	return createdMenuItem, nil
}

func (s *RestaurantService) UpdateMenuItem(
	ctx context.Context,
	input UpdateMenuItemInput,
) (domain.MenuItem, error) {
	if input.RestaurantID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidRestaurantID
	}

	if input.MenuItemID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemID
	}

	if input.Name == nil &&
		input.Description == nil &&
		input.Price == nil &&
		input.IsAvailable == nil {
		return domain.MenuItem{}, domain.ErrEmptyMenuItemUpdate
	}

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemName
	}

	if input.Price != nil && *input.Price < 0 {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemPrice
	}

	// Проверяем существование ресторана из URL.
	_, err := s.restaurants.GetByID(
		ctx,
		input.RestaurantID,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf("get restaurant: %w", err)
	}

	menuItem, err := s.menuItems.GetByID(
		ctx,
		input.MenuItemID,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf("get menu item: %w", err)
	}

	// Не позволяем ресторану изменить чужую позицию.
	if menuItem.RestaurantID != input.RestaurantID {
		return domain.MenuItem{}, domain.ErrMenuItemRestaurantMismatch
	}

	if input.Name != nil {
		menuItem.Name = strings.TrimSpace(*input.Name)
	}

	if input.Description != nil {
		menuItem.Description = *input.Description
	}

	if input.Price != nil {
		menuItem.Price = *input.Price
	}

	if input.IsAvailable != nil {
		menuItem.IsAvailable = *input.IsAvailable
	}

	updatedMenuItem, err := s.menuItems.Update(
		ctx,
		menuItem,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf("update menu item: %w", err)
	}

	return updatedMenuItem, nil
}
