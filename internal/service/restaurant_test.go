package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

type restaurantCatalogRepositoryStub struct {
	restaurants []domain.Restaurant
	restaurant  domain.Restaurant
	err         error

	listCalls int
	getCalls  int

	receivedID int64
}

func (r *restaurantCatalogRepositoryStub) ListActive(
	_ context.Context,
) ([]domain.Restaurant, error) {
	r.listCalls++

	if r.err != nil {
		return nil, r.err
	}

	return r.restaurants, nil
}

func (r *restaurantCatalogRepositoryStub) GetByID(
	_ context.Context,
	id int64,
) (domain.Restaurant, error) {
	r.getCalls++
	r.receivedID = id

	if r.err != nil {
		return domain.Restaurant{}, r.err
	}

	return r.restaurant, nil
}

type restaurantMenuRepositoryStub struct {
	items    []domain.MenuItem
	menuItem domain.MenuItem
	err      error

	listCalls   int
	getCalls    int
	createCalls int
	updateCalls int

	receivedRestaurantID int64
	receivedMenuItemID   int64
	receivedCreate       domain.MenuItem
	receivedUpdate       domain.MenuItem
}

func (r *restaurantMenuRepositoryStub) ListByRestaurantID(
	_ context.Context,
	restaurantID int64,
) ([]domain.MenuItem, error) {
	r.listCalls++
	r.receivedRestaurantID = restaurantID

	if r.err != nil {
		return nil, r.err
	}

	return r.items, nil
}

func (r *restaurantMenuRepositoryStub) GetByID(
	_ context.Context,
	id int64,
) (domain.MenuItem, error) {
	r.getCalls++
	r.receivedMenuItemID = id

	if r.err != nil {
		return domain.MenuItem{}, r.err
	}

	return r.menuItem, nil
}

func (r *restaurantMenuRepositoryStub) Create(
	_ context.Context,
	item domain.MenuItem,
) (domain.MenuItem, error) {
	r.createCalls++
	r.receivedCreate = item

	if r.err != nil {
		return domain.MenuItem{}, r.err
	}

	item.ID = 100

	return item, nil
}

func (r *restaurantMenuRepositoryStub) Update(
	_ context.Context,
	item domain.MenuItem,
) (domain.MenuItem, error) {
	r.updateCalls++
	r.receivedUpdate = item

	if r.err != nil {
		return domain.MenuItem{}, r.err
	}

	return item, nil
}

func TestRestaurantService_ListRestaurants(t *testing.T) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurants: []domain.Restaurant{
			{
				ID:       1,
				Name:     "Pizza House",
				IsActive: true,
			},
			{
				ID:       2,
				Name:     "Burger House",
				IsActive: true,
			},
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	restaurants, err := restaurantService.ListRestaurants(
		context.Background(),
	)

	require.NoError(t, err)
	require.Len(t, restaurants, 2)
	require.Equal(t, 1, restaurantRepo.listCalls)
	require.Equal(t, "Pizza House", restaurants[0].Name)
}

func TestRestaurantService_GetMenu(t *testing.T) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			Name:     "Pizza House",
			IsActive: true,
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{
		items: []domain.MenuItem{
			{
				ID:           10,
				RestaurantID: 1,
				Name:         "Pizza",
				Price:        70000,
				IsAvailable:  true,
			},
			{
				ID:           20,
				RestaurantID: 1,
				Name:         "Cola",
				Price:        15000,
				IsAvailable:  false,
			},
		},
	}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	menu, err := restaurantService.GetMenu(
		context.Background(),
		1,
	)

	require.NoError(t, err)

	require.Equal(t, int64(1), restaurantRepo.receivedID)
	require.Equal(t, int64(1), menuRepo.receivedRestaurantID)

	require.Len(t, menu, 2)

	// Недоступные позиции тоже должны присутствовать в меню.
	require.False(t, menu[1].IsAvailable)
}

func TestRestaurantService_GetMenu_RestaurantUnavailable(
	t *testing.T,
) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: false,
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	_, err := restaurantService.GetMenu(
		context.Background(),
		1,
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrRestaurantUnavailable,
	)

	// Если ресторан недоступен, меню читать не нужно.
	require.Equal(t, 0, menuRepo.listCalls)
}

func TestRestaurantService_CreateMenuItem(t *testing.T) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			Name:     "Pizza House",
			IsActive: false,
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	created, err := restaurantService.CreateMenuItem(
		context.Background(),
		CreateMenuItemInput{
			RestaurantID: 1,
			Name:         "  Margarita  ",
			Description:  "Tomato and mozzarella",
			Price:        59000,
		},
	)

	require.NoError(t, err)

	require.Equal(t, 1, menuRepo.createCalls)

	require.Equal(
		t,
		int64(1),
		menuRepo.receivedCreate.RestaurantID,
	)
	require.Equal(
		t,
		"Margarita",
		menuRepo.receivedCreate.Name,
	)
	require.Equal(
		t,
		int64(59000),
		menuRepo.receivedCreate.Price,
	)
	require.True(
		t,
		menuRepo.receivedCreate.IsAvailable,
	)

	require.Equal(t, int64(100), created.ID)
}

func TestRestaurantService_UpdateMenuItem(t *testing.T) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurant: domain.Restaurant{
			ID: 1,
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{
		menuItem: domain.MenuItem{
			ID:           10,
			RestaurantID: 1,
			Name:         "Pizza",
			Description:  "Old description",
			Price:        50000,
			IsAvailable:  true,
		},
	}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	isAvailable := false

	updated, err := restaurantService.UpdateMenuItem(
		context.Background(),
		UpdateMenuItemInput{
			RestaurantID: 1,
			MenuItemID:   10,
			IsAvailable:  &isAvailable,
		},
	)

	require.NoError(t, err)

	require.Equal(t, 1, menuRepo.updateCalls)

	// Не переданные PATCH-поля должны остаться прежними.
	require.Equal(t, "Pizza", updated.Name)
	require.Equal(t, "Old description", updated.Description)
	require.Equal(t, int64(50000), updated.Price)

	// Изменяется только переданное поле.
	require.False(t, updated.IsAvailable)
}

func TestRestaurantService_UpdateMenuItem_RestaurantMismatch(
	t *testing.T,
) {
	restaurantRepo := &restaurantCatalogRepositoryStub{
		restaurant: domain.Restaurant{
			ID: 1,
		},
	}

	menuRepo := &restaurantMenuRepositoryStub{
		menuItem: domain.MenuItem{
			ID:           10,
			RestaurantID: 2,
			Name:         "Pizza",
			Price:        50000,
			IsAvailable:  true,
		},
	}

	restaurantService := NewRestaurantService(
		restaurantRepo,
		menuRepo,
	)

	price := int64(60000)

	_, err := restaurantService.UpdateMenuItem(
		context.Background(),
		UpdateMenuItemInput{
			RestaurantID: 1,
			MenuItemID:   10,
			Price:        &price,
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrMenuItemRestaurantMismatch,
	)

	// Чужую позицию обновлять нельзя.
	require.Equal(t, 0, menuRepo.updateCalls)
}
