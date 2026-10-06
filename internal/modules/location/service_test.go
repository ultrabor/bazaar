package location

import (
	"context"
	"errors"
	"testing"
)

func TestCreateLocationRejectsInvalidType(t *testing.T) {
	service := NewService(nil)
	request := &CreateLocationRequest{
		CompanyId: "company-id",
		Name:      "Main",
		Address:   "Address",
		Type:      "invalid",
	}

	_, err := service.CreateLocation(context.Background(), request)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("CreateLocation error = %v, want %v", err, ErrInvalidType)
	}
}

func TestUpdateLocationRejectsInvalidType(t *testing.T) {
	service := NewService(nil)
	request := &UpdateLocationRequest{
		LocationId: "location-id",
		CompanyId:  "company-id",
		Name:       "Main",
		Address:    "Address",
		Type:       "invalid",
	}

	_, err := service.UpdateLocation(context.Background(), request)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("UpdateLocation error = %v, want %v", err, ErrInvalidType)
	}
}

func TestLocationServiceRejectsNilRequest(t *testing.T) {
	service := NewService(nil)

	if _, err := service.CreateLocation(context.Background(), nil); !errors.Is(err, ErrInvalidCred) {
		t.Fatalf("CreateLocation error = %v, want %v", err, ErrInvalidCred)
	}
	if _, err := service.UpdateLocation(context.Background(), nil); !errors.Is(err, ErrInvalidCred) {
		t.Fatalf("UpdateLocation error = %v, want %v", err, ErrInvalidCred)
	}
}
