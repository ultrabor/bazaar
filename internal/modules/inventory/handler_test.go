package inventory

import (
	"bazaar/internal/platform/httpx/httpxtest"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeInventoryService struct {
	err      error
	balances []Balance
	bfilter  BalanceFilter
	mfilter  MovementFilter
}

func (f *fakeInventoryService) ListBalances(_ context.Context, filter BalanceFilter) ([]Balance, error) {
	f.bfilter = filter
	return f.balances, f.err
}

func (f *fakeInventoryService) ListMovements(_ context.Context, filter MovementFilter) ([]Movement, error) {
	f.mfilter = filter
	return []Movement{}, f.err
}

func newInventoryRouter(svc inventoryService) http.Handler {
	h := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)
	r := chi.NewRouter()
	r.Get("/location/{id}/stock", h.ListStock)
	r.Get("/location/{id}/stock/{productId}/movements", h.ListMovements)
	return r
}

func TestListStockReturnsBalancesOfCallerCompany(t *testing.T) {
	svc := &fakeInventoryService{balances: []Balance{{ProductID: "p1", Quantity: 1500}}}

	res := httpxtest.Serve(t, newInventoryRouter(svc), http.MethodGet, "/location/loc-1/stock?limit=10&offset=20", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", res.Code, res.Body)
	}
	want := BalanceFilter{CompanyID: httpxtest.Caller.CompanyId, LocationID: "loc-1", Limit: 10, Offset: 20}
	if svc.bfilter != want {
		t.Errorf("filter = %+v, want %+v", svc.bfilter, want)
	}
	var got []Balance
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].Quantity != 1500 {
		t.Errorf("body = %+v", got)
	}
}

func TestListMovementsPassesPathParams(t *testing.T) {
	svc := &fakeInventoryService{}

	res := httpxtest.Serve(t, newInventoryRouter(svc), http.MethodGet, "/location/loc-1/stock/prod-1/movements", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", res.Code, res.Body)
	}
	want := MovementFilter{CompanyID: httpxtest.Caller.CompanyId, LocationID: "loc-1", ProductID: "prod-1"}
	if svc.mfilter != want {
		t.Errorf("filter = %+v, want %+v", svc.mfilter, want)
	}
}

func TestInventoryHandlerErrors(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		err    error
		status int
		code   string
	}{
		{"bad limit", "/location/l/stock?limit=abc", nil, http.StatusBadRequest, "invalid_pagination"},
		{"negative offset", "/location/l/stock/p/movements?offset=-1", nil, http.StatusBadRequest, "invalid_pagination"},
		{"location not a uuid", "/location/abc/stock", ErrInvalidLocationID, http.StatusBadRequest, "invalid_location_id"},
		{"product not a uuid", "/location/l/stock/abc/movements", ErrInvalidProductID, http.StatusBadRequest, "invalid_product_id"},
		{"unknown location", "/location/l/stock", ErrLocationNotFound, http.StatusNotFound, "location_not_found"},
		{"foreign location movements", "/location/l/stock/p/movements", ErrLocationNotFound, http.StatusNotFound, "location_not_found"},
		{"timeout", "/location/l/stock", context.DeadlineExceeded, http.StatusServiceUnavailable, "service_unavailable"},
		{"storage failure", "/location/l/stock", errors.New("boom"), http.StatusInternalServerError, "internal_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httpxtest.Serve(t, newInventoryRouter(&fakeInventoryService{err: tt.err}), http.MethodGet, tt.path, "")

			if res.Code != tt.status {
				t.Fatalf("status = %d, want %d", res.Code, tt.status)
			}
			if code := httpxtest.ErrorCode(t, res); code != tt.code {
				t.Errorf("code = %q, want %q", code, tt.code)
			}
		})
	}
}

func TestNonUUIDPathReturns400ThroughRealService(t *testing.T) {
	repo := &fakeRepo{}
	router := newInventoryRouter(NewService(repo, fakeTx{}))

	paths := map[string]string{
		"/location/abc/stock":                                "invalid_location_id",
		"/location/" + testLocation + "/stock/abc/movements": "invalid_product_id",
	}
	for path, wantCode := range paths {
		res := httpxtest.Serve(t, router, http.MethodGet, path, "")

		if res.Code != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want 400", path, res.Code)
			continue
		}
		if code := httpxtest.ErrorCode(t, res); code != wantCode {
			t.Errorf("GET %s code = %q, want %q", path, code, wantCode)
		}
	}
	if repo.filter != (MovementFilter{}) || repo.bfilter != (BalanceFilter{}) {
		t.Errorf("repository queried with an invalid id")
	}
}
