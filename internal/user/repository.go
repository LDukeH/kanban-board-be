package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query:= 
	`INSERT INTO users (
			email,
			password_hash
		)
	VALUES ($1, $2)
	RETURNING id, created_at, updated_at`


	return r.pool.QueryRow(ctx,query,user.Email,user.PasswordHash).Scan(&user.ID,&user.CreatedAt,&user.UpdatedAt);
}

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {

	query := `
		SELECT
			id,
			email,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
query := `
		SELECT
			id,
			email,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}

	err := r.pool.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil

}