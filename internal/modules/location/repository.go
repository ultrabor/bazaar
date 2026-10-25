package location

import (
	"bazaar/internal/domain"
	"bazaar/internal/platform/database"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type repository struct {
	db *database.Database
}

var _ domain.LocationRepository = (*repository)(nil)

func NewRepository(db *database.Database) domain.LocationRepository {
	return &repository{db: db}
}

func (r *repository) GetLocationById(ctx context.Context, locationId, companyId string) (*domain.Location, error) {
	var location domain.Location

	err := r.db.Conn(ctx).QueryRow(ctx, `
		SELECT id, company_id, name, address, type, archived
		FROM locations
		WHERE id = $1 AND company_id = $2 AND archived = false
	`, locationId, companyId).Scan(
		&location.Id,
		&location.CompanyId,
		&location.Name,
		&location.Address,
		&location.Type,
		&location.Archived,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}

	return &location, nil
}

func (r *repository) GetCompanyLocations(ctx context.Context, companyId string) ([]domain.Location, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `
		SELECT id, company_id, name, address, type, archived
		FROM locations
		WHERE company_id = $1 AND archived = false
	`, companyId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]domain.Location, 0)
	for rows.Next() {
		var location domain.Location
		if err := rows.Scan(
			&location.Id,
			&location.CompanyId,
			&location.Name,
			&location.Address,
			&location.Type,
			&location.Archived,
		); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}

	return locations, nil
}

func (r *repository) CreateLocation(ctx context.Context, req *domain.CreateLocationRequest) (*domain.CreateLocationResponse, error) {
	var locationId string

	err := r.db.Conn(ctx).QueryRow(ctx, `
		INSERT INTO locations (id, company_id, name, address, type)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id
	`, req.CompanyId, req.Name, req.Address, req.Type).Scan(&locationId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_unique_active_name_idx" {
			return nil, domain.ErrLocationAlreadyExists
		}
		return nil, err
	}

	return &domain.CreateLocationResponse{LocationID: locationId}, nil
}

func (r *repository) UpdateLocation(ctx context.Context, req *domain.UpdateLocationRequest) (*domain.UpdateLocationResponse, error) {
	var locationId string

	err := r.db.Conn(ctx).QueryRow(ctx, `
		UPDATE locations
		SET name = $1, address = $2, type = $3, updated_at = now()
		WHERE id = $4 AND company_id = $5 AND archived = false
		RETURNING id
	`, req.Name, req.Address, req.Type, req.LocationId, req.CompanyId).Scan(&locationId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "locations_unique_active_name_idx" {
			return nil, domain.ErrLocationAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}

	return &domain.UpdateLocationResponse{LocationID: locationId}, nil
}

func (r *repository) ArchiveLocation(ctx context.Context, locationId, companyId string) error {
	cmdTag, err := r.db.Conn(ctx).Exec(ctx, `
		UPDATE locations
		SET archived = true, updated_at = now()
		WHERE id = $1 AND company_id = $2 AND archived = false
	`, locationId, companyId)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrLocationNotFound
	}

	return nil
}
