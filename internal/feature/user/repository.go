package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, user *User) error {
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
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
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
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil

}

func (r *Repository) SearchByEmail(ctx context.Context, query string) ([]*UserResponse, error) {
	search := "%" + query + "%"
	searchQuery:=`
	SELECT
		id,
		email
	FROM users
	WHERE email ILIKE $1
	ORDER BY email
	LIMIT 10
	`

	rows, err := r.pool.Query(ctx, searchQuery, search)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	users := []*UserResponse{}

	for rows.Next() {
		user := &UserResponse{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}

	return users, nil
}