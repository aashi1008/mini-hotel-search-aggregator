package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/example/mini-hotel-aggregator/internal/validator"
)

type SearchRequest struct {
	City    string
	Checkin string
	Nights  int
	Adults  int
}


func NewSearchRequest(city, checkin, nights, adults string) (*SearchRequest, error) {
	if city == "" || checkin == "" || nights == "" || adults == "" {
		return nil, fmt.Errorf("missing required params")
	}
	nightsInt, err := strconv.Atoi(nights)
	if err != nil {
		return nil, fmt.Errorf("invalid nights")
	}
	adultsInt, err := strconv.Atoi(adults)
	if err != nil {
		return nil, fmt.Errorf("invalid adults")
	}
	return &SearchRequest{
		City:    city,
		Checkin: checkin,
		Nights:  nightsInt,
		Adults:  adultsInt,
	}, nil
}

func (r *SearchRequest) Validate() error {
	var errs []string

	city, err := validator.ValidateCity(r.City)
	if err != nil {
		errs = append(errs, err.Error())
	} else {
		r.City = city // normalized
	}

	_, err = validator.ValidateDate(r.Checkin)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if r.Nights <= 0 || r.Nights > 365 {
		errs = append(errs, "invalid or excessive nights")
	}
	if r.Adults <= 0 || r.Adults > 100 {
		errs = append(errs, "invalid or excessive adults")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}
