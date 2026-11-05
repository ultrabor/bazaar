package inventory

import (
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"testing"
)

type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeRepo struct {
	archived    bool
	lockErr     error
	productErr  error
	increaseErr error
	decreaseErr error
	insertErr   error
	balance     int64
	balanceErr  error
	locationErr error

	changed  int64
	inserted []Movement
	filter   MovementFilter
	bfilter  BalanceFilter
}

func (f *fakeRepo) LocationExists(context.Context, string, string) error { return f.locationErr }

func (f *fakeRepo) ListBalances(_ context.Context, filter BalanceFilter) ([]Balance, error) {
	f.bfilter = filter
	return []Balance{}, nil
}

func (f *fakeRepo) LockLocation(context.Context, string, string) (bool, error) {
	return f.archived, f.lockErr
}

func (f *fakeRepo) ProductExists(context.Context, string, string) error { return f.productErr }

func (f *fakeRepo) Increase(_ context.Context, _, _, _ string, qty int64) error {
	if f.increaseErr != nil {
		return f.increaseErr
	}
	f.changed += qty
	return nil
}

func (f *fakeRepo) Decrease(_ context.Context, _, _, _ string, qty int64) error {
	if f.decreaseErr != nil {
		return f.decreaseErr
	}
	f.changed -= qty
	return nil
}

func (f *fakeRepo) GetBalance(context.Context, string, string, string) (int64, error) {
	return f.balance, f.balanceErr
}

func (f *fakeRepo) InsertMovement(_ context.Context, m Movement) (*Movement, error) {
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	m.ID = "movement-1"
	f.inserted = append(f.inserted, m)
	return &m, nil
}

func (f *fakeRepo) ListMovements(_ context.Context, filter MovementFilter) ([]Movement, error) {
	f.filter = filter
	return []Movement{}, nil
}

func validChange(reason Reason) Change {
	return Change{
		CompanyID:  "company",
		LocationID: "location",
		ProductID:  "product",
		UserID:     "user",
		Qty:        1500,
		Reason:     reason,
	}
}

func TestIncreaseBalanceRecordsPositiveMovement(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeTx{})
	c := validChange(ReasonPurchase)
	c.RefType, c.RefID = RefTypePurchase, "purchase-1"

	m, err := svc.IncreaseBalance(context.Background(), c)

	if err != nil {
		t.Fatalf("IncreaseBalance: %v", err)
	}
	if repo.changed != 1500 {
		t.Errorf("balance change = %d, want 1500", repo.changed)
	}
	if m.Qty != 1500 || m.Reason != ReasonPurchase || m.ID == "" {
		t.Errorf("movement = %+v", m)
	}
	if m.RefType == nil || *m.RefType != RefTypePurchase || m.RefID == nil || *m.RefID != "purchase-1" {
		t.Errorf("ref = %v %v, want purchase purchase-1", m.RefType, m.RefID)
	}
}

func TestDecreaseBalanceRecordsNegativeMovement(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeTx{})

	m, err := svc.DecreaseBalance(context.Background(), validChange(ReasonSale))

	if err != nil {
		t.Fatalf("DecreaseBalance: %v", err)
	}
	if repo.changed != -1500 {
		t.Errorf("balance change = %d, want -1500", repo.changed)
	}
	if m.Qty != -1500 || m.RefType != nil || m.RefID != nil {
		t.Errorf("movement = %+v", m)
	}
}

func TestChangeValidation(t *testing.T) {
	tests := []struct {
		name     string
		increase bool
		edit     func(*Change)
		want     error
	}{
		{"missing company", true, func(c *Change) { c.CompanyID = "" }, apperror.ErrInvalidCred},
		{"missing location", true, func(c *Change) { c.LocationID = "" }, apperror.ErrInvalidCred},
		{"missing product", false, func(c *Change) { c.ProductID = "" }, apperror.ErrInvalidCred},
		{"missing user", false, func(c *Change) { c.UserID = "" }, apperror.ErrInvalidCred},
		{"zero qty", true, func(c *Change) { c.Qty = 0 }, ErrInvalidQuantity},
		{"negative qty", false, func(c *Change) { c.Qty = -5 }, ErrInvalidQuantity},
		{"qty above numeric(14,3)", true, func(c *Change) { c.Qty = MaxQty + 1 }, ErrInvalidQuantity},
		{"sale cannot increase", true, func(c *Change) { c.Reason = ReasonSale }, ErrInvalidReason},
		{"purchase cannot decrease", false, func(c *Change) { c.Reason = ReasonPurchase }, ErrInvalidReason},
		{"unknown reason", true, func(c *Change) { c.Reason = "gift" }, ErrInvalidReason},
		{"ref type without id", true, func(c *Change) { c.RefType = RefTypePurchase }, ErrInvalidRef},
		{"ref id without type", false, func(c *Change) { c.RefID = "x" }, ErrInvalidRef},
		{"unknown ref type", true, func(c *Change) { c.RefType, c.RefID = "invoice", "x" }, ErrInvalidRef},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := NewService(repo, fakeTx{})
			reason := ReasonSale
			if tt.increase {
				reason = ReasonPurchase
			}
			c := validChange(reason)
			tt.edit(&c)

			var err error
			if tt.increase {
				_, err = svc.IncreaseBalance(context.Background(), c)
			} else {
				_, err = svc.DecreaseBalance(context.Background(), c)
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if repo.changed != 0 || len(repo.inserted) != 0 {
				t.Errorf("repository touched: change %d, movements %d", repo.changed, len(repo.inserted))
			}
		})
	}
}

func TestChangeRejectedByStorage(t *testing.T) {
	storageErr := errors.New("connection reset")
	tests := []struct {
		name     string
		repo     *fakeRepo
		increase bool
		want     error
	}{
		{"archived location on increase", &fakeRepo{archived: true}, true, ErrLocationArchived},
		{"archived location on decrease", &fakeRepo{archived: true}, false, ErrLocationArchived},
		{"location of another company", &fakeRepo{lockErr: ErrLocationNotFound}, true, ErrLocationNotFound},
		{"product of another company", &fakeRepo{productErr: ErrProductNotFound}, false, ErrProductNotFound},
		{"insufficient stock", &fakeRepo{decreaseErr: ErrInsufficientStock}, false, ErrInsufficientStock},
		{"increase fails", &fakeRepo{increaseErr: storageErr}, true, storageErr},
		{"movement insert fails", &fakeRepo{insertErr: storageErr}, true, storageErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, fakeTx{})

			var err error
			if tt.increase {
				_, err = svc.IncreaseBalance(context.Background(), validChange(ReasonPurchase))
			} else {
				_, err = svc.DecreaseBalance(context.Background(), validChange(ReasonSale))
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if tt.repo.insertErr == nil && len(tt.repo.inserted) != 0 {
				t.Errorf("movement recorded for a rejected change")
			}
		})
	}
}

func TestCheckStock(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
		qty     int64
		product string
		want    error
	}{
		{"enough", 2000, 1500, "product", nil},
		{"exactly enough", 1500, 1500, "product", nil},
		{"short", 1000, 1500, "product", ErrInsufficientStock},
		{"no balance row", 0, 1, "product", ErrInsufficientStock},
		{"zero qty", 2000, 0, "product", ErrInvalidQuantity},
		{"missing product", 2000, 1, "", apperror.ErrInvalidCred},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&fakeRepo{balance: tt.balance}, fakeTx{})

			err := svc.CheckStock(context.Background(), "company", "location", tt.product, tt.qty)

			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCheckStockPropagatesStorageError(t *testing.T) {
	storageErr := errors.New("timeout")
	svc := NewService(&fakeRepo{balanceErr: storageErr}, fakeTx{})

	err := svc.CheckStock(context.Background(), "company", "location", "product", 1)

	if !errors.Is(err, storageErr) {
		t.Fatalf("err = %v, want %v", err, storageErr)
	}
}

func TestListMovementsLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"default", 0, DefaultPageLimit},
		{"capped", 1000, MaxPageLimit},
		{"kept", 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := NewService(repo, fakeTx{})
			f := MovementFilter{CompanyID: "c", LocationID: "l", ProductID: "p", Limit: tt.limit}

			if _, err := svc.ListMovements(context.Background(), f); err != nil {
				t.Fatalf("ListMovements: %v", err)
			}

			if repo.filter.Limit != tt.want {
				t.Errorf("limit = %d, want %d", repo.filter.Limit, tt.want)
			}
		})
	}
}

func TestListMovementsValidation(t *testing.T) {
	svc := NewService(&fakeRepo{}, fakeTx{})
	filters := []MovementFilter{
		{LocationID: "l", ProductID: "p"},
		{CompanyID: "c", LocationID: "l", ProductID: "p", Offset: -1},
	}
	for _, f := range filters {
		if _, err := svc.ListMovements(context.Background(), f); !errors.Is(err, apperror.ErrInvalidCred) {
			t.Errorf("ListMovements(%+v) err = %v, want ErrInvalidCred", f, err)
		}
	}
}

func TestListBalancesLimit(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeTx{})

	if _, err := svc.ListBalances(context.Background(), BalanceFilter{CompanyID: "c", LocationID: "l", Limit: 1000}); err != nil {
		t.Fatalf("ListBalances: %v", err)
	}

	if repo.bfilter.Limit != MaxPageLimit {
		t.Errorf("limit = %d, want %d", repo.bfilter.Limit, MaxPageLimit)
	}
}

func TestListBalancesValidation(t *testing.T) {
	svc := NewService(&fakeRepo{}, fakeTx{})
	filters := []BalanceFilter{
		{LocationID: "l"},
		{CompanyID: "c"},
		{CompanyID: "c", LocationID: "l", Offset: -1},
	}
	for _, f := range filters {
		if _, err := svc.ListBalances(context.Background(), f); !errors.Is(err, apperror.ErrInvalidCred) {
			t.Errorf("ListBalances(%+v) err = %v, want ErrInvalidCred", f, err)
		}
	}
}

func TestReadsRejectUnknownLocation(t *testing.T) {
	svc := NewService(&fakeRepo{locationErr: ErrLocationNotFound}, fakeTx{})
	ctx := context.Background()

	_, balancesErr := svc.ListBalances(ctx, BalanceFilter{CompanyID: "c", LocationID: "l"})
	_, movementsErr := svc.ListMovements(ctx, MovementFilter{CompanyID: "c", LocationID: "l", ProductID: "p"})

	if !errors.Is(balancesErr, ErrLocationNotFound) {
		t.Errorf("ListBalances err = %v, want ErrLocationNotFound", balancesErr)
	}
	if !errors.Is(movementsErr, ErrLocationNotFound) {
		t.Errorf("ListMovements err = %v, want ErrLocationNotFound", movementsErr)
	}
}
