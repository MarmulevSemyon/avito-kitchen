package domain

import "time"

type MenuItem struct {
	ID           int64
	RestaurantID int64
	Name         string
	Description  string
	Price        int64
	IsAvailable  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
