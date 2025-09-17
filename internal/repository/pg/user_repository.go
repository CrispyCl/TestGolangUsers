package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	uniqueViolation = pq.ErrorCode("23505")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {

	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, email string, passHash []byte, role models.UserRole) (int64, error) {
	const op = "repository.user.postgres.Create"

	query := sq.Insert("users").
		Columns("email", "pass_hash", "role").
		Values(email, passHash, role).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("%s: build query: %w", op, err)
	}

	var id int64
	if err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == uniqueViolation {
				return 0, fmt.Errorf("%s: %w", op, repository.ErrUserExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const op = "repository.user.postgres.GetByID"

	return r.getOne(ctx, sq.Eq{"id": id}, op)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.user.postgres.GetByEmail"

	return r.getOne(ctx, sq.Eq{"email": email}, op)
}

func (r *UserRepository) getOne(ctx context.Context, where sq.Eq, op string) (*models.User, error) {
	query := sq.Select("id", "email", "pass_hash", "role", "created_at", "last_seen").
		From("users").
		Where(where).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: build query: %w", op, err)
	}

	var user models.User
	if err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&user.ID, &user.Email, &user.PassHash, &user.Role, &user.CreatedAt, &user.LastSeen); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
