package location

import (
	"bazaar/internal/platform/database"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetLocationById(ctx context.Context, locationId, companyId string) (*Location, error) {
	var location Location

	err := r.db.GetDB().QueryRow(ctx, `
		SELECT id, company_id, name, address
		FROM locations
		WHERE id = $1, company_id = $2 and deleted_at = 0
	`, locationId, companyId).Scan(&location.Id, &location.CompanyId, &location.Name, &location.Address)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &location, nil
}

func (r *Repository) GetCompanyLocations(ctx context.Context, companyId string) ([]Location, error) {
	rows, err := r.db.GetDB().Query(ctx, `
		SELECT id, company_id, name, address
		FROM locations
		WHERE company_id = $1 and deleted_at = 0
	`, companyId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.Id, &location.CompanyId, &location.Name, &location.Address); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return locations, nil
}

func (r *Repository) CreateLocation(ctx context.Context, req *CreateLocationRequest) (*CreateLocationResponse, error) {
	var locationId string

	err := r.db.GetDB().QueryRow(ctx, `
		INSERT INTO locations (id, company_id, name, address, type)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id
	`, req.CompanyId, req.Name, req.Address, req.Type).Scan(&locationId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_company_id_name_key" {
			return nil, ErrLocationAlreadyExists
		}
		return nil, err
	}

	return &CreateLocationResponse{LocationID: locationId}, nil
}

func (r *Repository) UpdateLocation(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error) {
	var locationId string

	err := r.db.GetDB().QueryRow(ctx, `
		UPDATE locations
		SET name = $1, address = $2, type = $3
		WHERE id = $4 AND company_id = $5
		RETURNING id
	`, req.Name, req.Address, req.Type, req.LocationId, req.CompanyId).Scan(&locationId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_company_id_name_key" {
			return nil, ErrLocationAlreadyExists
		}
		return nil, err
	}

	return &UpdateLocationResponse{LocationID: locationId}, nil
}
