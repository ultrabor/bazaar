package inventory

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"context"
	"fmt"

	"github.com/google/uuid"
)

const MaxQty int64 = 99_999_999_999_999

var (
	increaseReasons = map[Reason]bool{
		ReasonPurchase:   true,
		ReasonReturn:     true,
		ReasonTransferIn: true,
		ReasonStocktake:  true,
	}
	decreaseReasons = map[Reason]bool{
		ReasonSale:        true,
		ReasonTransferOut: true,
		ReasonStocktake:   true,
		ReasonWriteOff:    true,
	}
	refTypes = map[RefType]bool{
		RefTypePurchase:  true,
		RefTypeSale:      true,
		RefTypeReturn:    true,
		RefTypeTransfer:  true,
		RefTypeStocktake: true,
	}
)

type Service struct {
	repo repository
	tx   database.TxManager
}

func NewService(repo repository, tx database.TxManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func (s *Service) IncreaseBalance(ctx context.Context, c Change) (*Movement, error) {
	return s.apply(ctx, c, increaseReasons, +1, s.repo.Increase)
}

func (s *Service) DecreaseBalance(ctx context.Context, c Change) (*Movement, error) {
	return s.apply(ctx, c, decreaseReasons, -1, s.repo.Decrease)
}

func (s *Service) CheckStock(ctx context.Context, companyID, locationID, productID string, qty int64) error {
	if err := validQty(qty); err != nil {
		return err
	}
	balance, err := s.GetBalance(ctx, companyID, locationID, productID)
	if err != nil {
		return err
	}
	if balance < qty {
		return ErrInsufficientStock
	}
	return nil
}

func (s *Service) GetBalance(ctx context.Context, companyID, locationID, productID string) (int64, error) {
	if companyID == "" || locationID == "" || productID == "" {
		return 0, apperror.ErrInvalidCred
	}
	if err := validIDs(locationID, productID); err != nil {
		return 0, err
	}
	return s.repo.GetBalance(ctx, companyID, locationID, productID)
}

func (s *Service) ListMovements(ctx context.Context, f MovementFilter) ([]Movement, error) {
	if f.CompanyID == "" || f.LocationID == "" || f.ProductID == "" || f.Offset < 0 {
		return nil, apperror.ErrInvalidCred
	}
	if err := validIDs(f.LocationID, f.ProductID); err != nil {
		return nil, err
	}
	if err := s.repo.LocationExists(ctx, f.CompanyID, f.LocationID); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	return s.repo.ListMovements(ctx, f)
}

func (s *Service) ListBalances(ctx context.Context, f BalanceFilter) ([]Balance, error) {
	if f.CompanyID == "" || f.LocationID == "" || f.Offset < 0 {
		return nil, apperror.ErrInvalidCred
	}
	if _, err := uuid.Parse(f.LocationID); err != nil {
		return nil, ErrInvalidLocationID
	}
	if err := s.repo.LocationExists(ctx, f.CompanyID, f.LocationID); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	return s.repo.ListBalances(ctx, f)
}

type balanceChange func(ctx context.Context, companyID, locationID, productID string, qty int64) error

func (s *Service) apply(ctx context.Context, c Change, reasons map[Reason]bool, sign int64, change balanceChange) (*Movement, error) {
	if err := validateChange(c, reasons); err != nil {
		return nil, err
	}

	var movement *Movement
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.checkTargets(ctx, c); err != nil {
			return err
		}
		if err := change(ctx, c.CompanyID, c.LocationID, c.ProductID, c.Qty); err != nil {
			return err
		}
		m, err := s.repo.InsertMovement(ctx, newMovement(c, sign*c.Qty))
		if err != nil {
			return err
		}
		movement = m
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", c.Reason, c.ProductID, err)
	}
	return movement, nil
}

func (s *Service) checkTargets(ctx context.Context, c Change) error {
	archived, err := s.repo.LockLocation(ctx, c.CompanyID, c.LocationID)
	if err != nil {
		return err
	}
	if archived {
		return ErrLocationArchived
	}
	productArchived, err := s.repo.LockProduct(ctx, c.CompanyID, c.ProductID)
	if err != nil {
		return err
	}
	// An archived product can still leave stock, come back as a return or be
	// counted, so only new purchases are rejected.
	if productArchived && c.Reason == ReasonPurchase {
		return ErrProductArchived
	}
	return nil
}

func validateChange(c Change, reasons map[Reason]bool) error {
	if c.CompanyID == "" || c.LocationID == "" || c.ProductID == "" || c.UserID == "" {
		return apperror.ErrInvalidCred
	}
	if err := validIDs(c.LocationID, c.ProductID); err != nil {
		return err
	}
	if err := validQty(c.Qty); err != nil {
		return err
	}
	if !reasons[c.Reason] {
		return ErrInvalidReason
	}
	if (c.RefType == "") != (c.RefID == "") {
		return ErrInvalidRef
	}
	if c.RefType != "" && !refTypes[c.RefType] {
		return ErrInvalidRef
	}
	if _, err := uuid.Parse(c.RefID); c.RefID != "" && err != nil {
		return ErrInvalidRef
	}
	return nil
}

// validIDs rejects ids that are not UUIDs before they reach Postgres,
// which would fail with 22P02 and surface as a 500.
func validIDs(locationID, productID string) error {
	if _, err := uuid.Parse(locationID); err != nil {
		return ErrInvalidLocationID
	}
	if _, err := uuid.Parse(productID); err != nil {
		return ErrInvalidProductID
	}
	return nil
}

func validQty(qty int64) error {
	if qty <= 0 || qty > MaxQty {
		return ErrInvalidQuantity
	}
	return nil
}

func newMovement(c Change, signedQty int64) Movement {
	m := Movement{
		CompanyID:  c.CompanyID,
		LocationID: c.LocationID,
		ProductID:  c.ProductID,
		UserID:     c.UserID,
		Qty:        signedQty,
		Reason:     c.Reason,
	}
	if c.RefType != "" {
		refType, refID := c.RefType, c.RefID
		m.RefType, m.RefID = &refType, &refID
	}
	return m
}

func clampLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultPageLimit
	case limit > MaxPageLimit:
		return MaxPageLimit
	}
	return limit
}

type repository interface {
	LockLocation(ctx context.Context, companyID, locationID string) (bool, error)
	LocationExists(ctx context.Context, companyID, locationID string) error
	ListBalances(ctx context.Context, f BalanceFilter) ([]Balance, error)
	// LockProduct must run inside a transaction. It returns whether the
	// product is archived, or ErrProductNotFound.
	LockProduct(ctx context.Context, companyID, productID string) (bool, error)
	Increase(ctx context.Context, companyID, locationID, productID string, qty int64) error
	Decrease(ctx context.Context, companyID, locationID, productID string, qty int64) error
	GetBalance(ctx context.Context, companyID, locationID, productID string) (int64, error)
	InsertMovement(ctx context.Context, m Movement) (*Movement, error)
	ListMovements(ctx context.Context, f MovementFilter) ([]Movement, error)
}

var _ repository = (*Repository)(nil)
