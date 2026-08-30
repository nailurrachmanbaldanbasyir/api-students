package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("student tidak ditemukan")
	ErrDuplicate = errors.New("NIM sudah digunakan")
)

type StudentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{
		db: db,
	}
}

type ListParams struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

var allowedSort = map[string]string{
	"id":        "id",
	"nim":       "nim",
	"name":      "name",
	"grade":     "grade",
	"is_active": "is_active",
}

func (r *StudentRepository) List(
	ctx context.Context,
	params ListParams,
) ([]model.Student, int, error) {
	sortColumn, ok := allowedSort[params.Sort]
	if !ok {
		sortColumn = "id"
	}

	order := "ASC"
	if strings.ToLower(params.Order) == "desc" {
		order = "DESC"
	}

	offset := (params.Page - 1) * params.Limit

	args := []any{}
	where := []string{"1=1"}

	if params.Search != "" {
		args = append(args, "%"+params.Search+"%")
		where = append(where, fmt.Sprintf(
			"(name ILIKE $%d OR nim ILIKE $%d)",
			len(args),
			len(args),
		))
	}

	if params.IsActive != nil {
		args = append(args, *params.IsActive)
		where = append(where, fmt.Sprintf(
			"is_active = $%d",
			len(args),
		))
	}

	whereSQL := strings.Join(where, " AND ")

	countQuery := `
		SELECT COUNT(*)
		FROM students
		WHERE ` + whereSQL

	var total int

	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, params.Limit)
	limitParam := len(args)

	args = append(args, offset)
	offsetParam := len(args)

	query := fmt.Sprintf(`
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortColumn, order, limitParam, offsetParam)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		if err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *StudentRepository) GetByID(
	ctx context.Context,
	id int,
) (model.Student, error) {
	var student model.Student

	query := `
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		return model.Student{}, err
	}

	return student, nil
}

func (r *StudentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	query := `
		INSERT INTO students (
			nim,
			name,
			grade,
			is_active
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nim, name, grade, is_active, created_at
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return result, nil
}

func (r *StudentRepository) Update(
	ctx context.Context,
	id int,
	student model.Student,
) (model.Student, error) {
	query := `
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, created_at
	`

	var result model.Student

	err := r.db.QueryRow(
		ctx,
		query,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		id,
	).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return result, nil
}

func (r *StudentRepository) Patch(
	ctx context.Context,
	id int,
	fields map[string]any,
) (model.Student, error) {
	if len(fields) == 0 {
		return model.Student{}, errors.New("tidak ada field yang diubah")
	}

	setParts := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	fieldOrder := []string{"nim", "name", "grade", "is_active"}

	for _, field := range fieldOrder {
		value, exists := fields[field]
		if !exists {
			continue
		}

		args = append(args, value)
		setParts = append(
			setParts,
			fmt.Sprintf("%s = $%d", field, len(args)),
		)
	}

	if len(setParts) == 0 {
		return model.Student{}, errors.New("field tidak valid")
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE students
		SET %s
		WHERE id = $%d
		RETURNING id, nim, name, grade, is_active, created_at
	`, strings.Join(setParts, ", "), len(args))

	var result model.Student

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.NIM,
		&result.Name,
		&result.Grade,
		&result.IsActive,
		&result.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}

		return model.Student{}, err
	}

	return result, nil
}

func (r *StudentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM students WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
