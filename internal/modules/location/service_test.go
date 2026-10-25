package location

import (
	"bazaar/internal/domain"
	"context"
	"errors"
	"testing"
)

func TestCreateLocationRejectsInvalidType(t *testing.T) {
	service := NewService(nil)
	request := &domain.CreateLocationRequest{
		CompanyId: "company-id",
		Name:      "Main",
		Address:   "Address",
		Type:      "invalid",
	}

	_, err := service.CreateLocation(context.Background(), request)
	if !errors.Is(err, domain.ErrInvalidLocationType) {
		t.Fatalf("CreateLocation error = %v, want %v", err, domain.ErrInvalidLocationType)
	}
}

func TestUpdateLocationRejectsInvalidType(t *testing.T) {
	service := NewService(nil)
	request := &domain.UpdateLocationRequest{
		LocationId: "location-id",
		CompanyId:  "company-id",
		Name:       "Main",
		Address:    "Address",
		Type:       "invalid",
	}

	_, err := service.UpdateLocation(context.Background(), request)
	if !errors.Is(err, domain.ErrInvalidLocationType) {
		t.Fatalf("UpdateLocation error = %v, want %v", err, domain.ErrInvalidLocationType)
	}
}

func TestLocationServiceRejectsNilRequest(t *testing.T) {
	service := NewService(nil)

	if _, err := service.CreateLocation(context.Background(), nil); !errors.Is(err, domain.ErrInvalidCred) {
		t.Fatalf("CreateLocation error = %v, want %v", err, domain.ErrInvalidCred)
	}
	if _, err := service.UpdateLocation(context.Background(), nil); !errors.Is(err, domain.ErrInvalidCred) {
		t.Fatalf("UpdateLocation error = %v, want %v", err, domain.ErrInvalidCred)
	}
}
