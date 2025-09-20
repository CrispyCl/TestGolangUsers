package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
)

type UserRepository struct {
	data       map[int64]*models.User
	emailIndex map[string]int64
	mu         sync.RWMutex
}

func NewUserRepository() *UserRepository {

	return &UserRepository{
		data:       make(map[int64]*models.User),
		emailIndex: make(map[string]int64),
	}
}

func (r *UserRepository) Create(ctx context.Context, email string, passHash []byte, role models.UserRole) (int64, error) {
	const op = "repository.user.mock.Create"

	switch role {
	case models.UserRoleUser, models.UserRoleManager, models.UserRoleAdmin, models.UserRoleUnspecified:
		// available
	default:
		return 0, fmt.Errorf("%s: %w", op, repository.ErrInvalidUserRole)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.emailIndex[email]; exists {
		return 0, fmt.Errorf("%s: %w", op, repository.ErrUserExists)
	}

	id := int64(len(r.data)) + int64(1)
	r.emailIndex[email] = id
	r.data[id] = &models.User{
		ID:        id,
		Email:     email,
		PassHash:  passHash,
		Role:      role,
		CreatedAt: time.Now(),
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const op = "repository.user.mock.GetByID"
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.user.mock.GetByEmail"
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.emailIndex[email]
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
	}

	user, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
	}

	return user, nil
}

func (r *UserRepository) UpdateLastSeen(ctx context.Context, id int64) (*models.User, error) {
	const op = "repository.user.mock.UpdateLastSeen"
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
	}

	user.LastSeen = time.Now()
	r.data[id] = user

	return user, nil
}
