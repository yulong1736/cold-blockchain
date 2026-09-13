package services

import (
	"errors"
	"testing"
	"time"
)

func TestValidateTransportTimes(t *testing.T) {
	arrival := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)
	departure := time.Date(2026, 4, 7, 11, 0, 0, 0, time.UTC)

	if err := validateTransportTimes(arrival, departure, time.Time{}); err != nil {
		t.Fatalf("valid times should pass: %v", err)
	}

	err := validateTransportTimes(time.Time{}, departure, time.Time{})
	if err == nil || !errors.Is(err, ErrProductValidation) {
		t.Fatalf("zero arrival should be validation error, got: %v", err)
	}

	err = validateTransportTimes(arrival, arrival.Add(-time.Minute), time.Time{})
	if err == nil || !errors.Is(err, ErrProductValidation) {
		t.Fatalf("departure before arrival should be validation error, got: %v", err)
	}

	prevDep := arrival.Add(30 * time.Minute)
	err = validateTransportTimes(arrival, departure, prevDep)
	if err == nil || !errors.Is(err, ErrProductValidation) {
		t.Fatalf("arrival before prev departure should be validation error, got: %v", err)
	}
}

