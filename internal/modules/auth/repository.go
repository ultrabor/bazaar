package auth

import (
	"bazaar/internal/platform/database"
	"bazaar/internal/platform/support/apperror"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCompany(ctx context.Context, name string) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, "INSERT INTO companies(id, label) VALUES(gen_random_uuid(), $1) RETURNING id",
		name).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create company", err)
	}

	return id, nil
}

func (r *Repository) CreateUser(
	ctx context.Context,
	companyID, roleID, firstName, lastName, phone, passwordHash string,
) (string, error) {
	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, `
        INSERT INTO users (
            id, company_id, user_role_id,
            first_name, last_name, phone, password_hash
        )
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
        RETURNING id
    `, companyID, roleID, firstName, lastName, phone, passwordHash).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "users_active_phone_unique" {
			return "", ErrPhoneTaken
		}
		return "", apperror.New("failed to create user", err)
	}

	return id, nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	var u User

	err := r.db.Conn(ctx).QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone, password_hash FROM users WHERE phone = $1 and deleted_at = 0", phone).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone, &u.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by phone", err)
	}

	return &u, nil
}

func (r *Repository) GetUserById(ctx context.Context, userId string) (*User, error) {
	var u User

	err := r.db.Conn(ctx).QueryRow(ctx, "SELECT id, company_id, user_role_id, first_name, last_name, phone FROM users WHERE id = $1 and deleted_at = 0", userId).Scan(&u.Id, &u.CompanyId, &u.UserRoleId, &u.FirstName, &u.LastName, &u.Phone)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInvalidCred
		}
		return nil, apperror.New("failed to get user by id", err)
	}

	return &u, nil
}

func (r *Repository) CreateSession(ctx context.Context, userId, tokenHash string, expiresAt time.Time, userAgent, ip string) (string, error) {
	query := `
		INSERT INTO sessions(user_id, token_hash, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, '')::inet)
		RETURNING id
	`

	var id string

	err := r.db.Conn(ctx).QueryRow(ctx, query, userId, tokenHash, expiresAt, userAgent, ip).Scan(&id)

	if err != nil {
		return "", apperror.New("failed to create session", err)
	}

	return id, nil
}

func (r *Repository) GetActiveSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	query := `
		SELECT s.id, s.user_id, s.token_hash, s.expires_at, s.revoked_at, s.user_agent, s.ip::text
		from sessions s
		join users u on s.user_id = u.id
		where s.token_hash = $1
		and s.revoked_at is null
		and s.expires_at > now()
		and u.deleted_at = 0
	`

	var s Session

	err := r.db.Conn(ctx).QueryRow(ctx, query, tokenHash).Scan(
		&s.Id,
		&s.UserId,
		&s.TokenHash,
		&s.ExpiresAt,
		&s.RevokedAt,
		&s.UserAgent,
		&s.Ip,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrInvalidCred
		}

		return nil, apperror.New("failed to get session", err)
	}

	return &s, nil
}

// RevokeSession returns ErrInvalidCred when the session is missing or
// already revoked.
func (r *Repository) RevokeSession(ctx context.Context, sessionId string) error {
	query := `
		UPDATE sessions
		SET revoked_at = now()
		where id = $1 and revoked_at is null
	`

	tag, err := r.db.Conn(ctx).Exec(ctx, query, sessionId)
	if err != nil {
		return apperror.New("failed to revoke session", err)
	}

	if tag.RowsAffected() == 0 {
		return apperror.ErrInvalidCred
	}

	return nil
}
