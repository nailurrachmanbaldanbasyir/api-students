package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrUserNotFound  = errors.New("user tidak ditemukan")
	ErrUserDuplicate = errors.New("username atau email sudah dipakai")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	query := `
		INSERT INTO users (username, email, password, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password, role, is_active, created_at
	`
	var result model.User
	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role, u.IsActive).Scan(
		&result.ID, &result.Username, &result.Email, &result.Password, &result.Role, &result.IsActive, &result.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrUserDuplicate
		}
		return model.User{}, fmt.Errorf("membuat user: %w", err)
	}
	return result, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}
