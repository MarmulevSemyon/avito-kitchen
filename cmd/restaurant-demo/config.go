package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIBaseURL   = "http://localhost:8080"
	defaultRestaurantID = int64(1)
	defaultPollInterval = 5 * time.Second
	defaultStatusDelay  = 2 * time.Second
)

type config struct {
	APIBaseURL   string
	RestaurantID int64
	PollInterval time.Duration
	StatusDelay  time.Duration
}

func loadConfig() (config, error) {
	restaurantID, err := envInt64(
		"RESTAURANT_ID",
		defaultRestaurantID,
	)
	if err != nil {
		return config{}, err
	}

	pollInterval, err := envDuration(
		"POLL_INTERVAL",
		defaultPollInterval,
	)
	if err != nil {
		return config{}, err
	}

	statusDelay, err := envDuration(
		"STATUS_DELAY",
		defaultStatusDelay,
	)
	if err != nil {
		return config{}, err
	}

	apiBaseURL := strings.TrimSpace(
		os.Getenv("API_BASE_URL"),
	)
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}

	return config{
		APIBaseURL:   strings.TrimRight(apiBaseURL, "/"),
		RestaurantID: restaurantID,
		PollInterval: pollInterval,
		StatusDelay:  statusDelay,
	}, nil
}

func envInt64(
	name string,
	defaultValue int64,
) (int64, error) {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseInt(
		value,
		10,
		64,
	)
	if err != nil || parsed <= 0 {
		return 0, errors.New(
			name + " must be a positive integer",
		)
	}

	return parsed, nil
}

func envDuration(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, errors.New(
			name + " must be a positive duration",
		)
	}

	return duration, nil
}
