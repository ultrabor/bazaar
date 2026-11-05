package inventory

import (
	"errors"
	"time"
)

// QtyScale is how many stored units make one unit of a product.
// Quantities are int64 thousandths: 1.5 kg is 1500. The database keeps
// them as NUMERIC(14,3), so the conversion is exact.
const QtyScale = 1000

// Reason says why stock moved.
type Reason string

const (
	ReasonPurchase    Reason = "purchase"
	ReasonSale        Reason = "sale"
	ReasonReturn      Reason = "return"
	ReasonTransferIn  Reason = "transfer_in"
	ReasonTransferOut Reason = "transfer_out"
	ReasonStocktake   Reason = "stocktake"
	ReasonWriteOff    Reason = "write_off"
)

type RefType string

const (
	RefTypePurchase  RefType = "purchase"
	RefTypeSale      RefType = "sale"
	RefTypeReturn    RefType = "return"
	RefTypeTransfer  RefType = "transfer"
	RefTypeStocktake RefType = "stocktake"
)

type Movement struct {
	ID         string    `json:"id"`
	CompanyID  string    `json:"company_id"`
	LocationID string    `json:"location_id"`
	ProductID  string    `json:"product_id"`
	UserID     string    `json:"user_id"`
	Qty        int64     `json:"qty"`
	Reason     Reason    `json:"reason"`
	RefType    *RefType  `json:"ref_type"`
	RefID      *string   `json:"ref_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Balance struct {
	CompanyID  string    `json:"company_id"`
	LocationID string    `json:"location_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int64     `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Change struct {
	CompanyID  string
	LocationID string
	ProductID  string
	UserID     string
	Qty        int64
	Reason     Reason
	RefType    RefType
	RefID      string
}

type MovementFilter struct {
	CompanyID  string
	LocationID string
	ProductID  string
	Limit      int
	Offset     int
}

type BalanceFilter struct {
	CompanyID  string
	LocationID string
	Limit      int
	Offset     int
}

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 200
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrLocationArchived  = errors.New("location is archived")
	ErrLocationNotFound  = errors.New("location not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrProductArchived   = errors.New("product is archived")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInvalidLocationID = errors.New("invalid location id")
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrInvalidReason     = errors.New("invalid reason for this operation")
	ErrInvalidRef        = errors.New("ref_type and ref_id must be set together")
)
