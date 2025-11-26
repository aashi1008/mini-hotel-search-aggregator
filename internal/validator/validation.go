package validator

import (
	"errors"
	"strings"
	"time"
)

func ValidateCity(s string) (string, error) {
	c := strings.TrimSpace(strings.ToLower(s))
	if len(c) < 2 {
		return "", errors.New("invalid city")
	}
	return c, nil
}

func ValidateDate(dateStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid checkin date")
	}
	return t, nil
}
