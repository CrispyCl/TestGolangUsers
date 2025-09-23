package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/service"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, email string, passHash []byte, role models.UserRole) (id int64, err error)
	GetByID(ctx context.Context, id int64) (user *models.User, err error)
	GetByEmail(ctx context.Context, email string) (user *models.User, err error)
	UpdateLastSeen(ctx context.Context, id int64) (user *models.User, err error)
}

type UserService struct {
	log      logger.Logger
	userRepo UserRepository
}

func NewUserService(log logger.Logger, userRepo UserRepository) *UserService {

	return &UserService{
		log:      log,
		userRepo: userRepo,
	}
}

func (s *UserService) Create(ctx context.Context, email, password string, role models.UserRole) (int64, error) {
	const op = "service.user.Create"
	log := s.log.With(zap.String("op", op), zap.String("email", email), zap.String("role", string(role)))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(ctx, "failed to generate password hash", zap.Error(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.userRepo.Create(ctx, email, passHash, role)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			log.Debug(ctx, "user with this email is exists", zap.Error(err))
		} else if errors.Is(err, repository.ErrInvalidUserRole) {
			log.Error(ctx, "invalid user role", zap.Error(err))
		} else {
			log.Error(ctx, "failed to save user", zap.Error(err))
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *UserService) CheckPassword(ctx context.Context, email, password string) (bool, error) {
	const op = "service.user.CheckPassword"
	log := s.log.With(zap.String("op", op), zap.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Debug(ctx, "user with this email not found", zap.Error(err))
			return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
		}

		log.Error(ctx, "failed to get user", zap.Error(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Debug(ctx, "failed to compare password", zap.Error(err))
		return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
	}

	return true, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "service.user.GetByEmail"
	log := s.log.With(zap.String("op", op), zap.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Debug(ctx, "user with this email not found", zap.Error(err))
		} else {
			log.Error(ctx, "failed to get user", zap.Error(err))
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const op = "service.user.GetByID"
	log := s.log.With(zap.String("op", op), zap.Int64("id", id))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Debug(ctx, "user with this id not found", zap.Error(err))
		} else {
			log.Error(ctx, "failed to get user", zap.Error(err))
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) UpdateLastSeen(ctx context.Context, id int64) error {
	const op = "service.user.UpdateLastSeen"
	log := s.log.With(zap.String("op", op), zap.Int64("id", id))

	_, err := s.userRepo.UpdateLastSeen(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Debug(ctx, "user with this id not found", zap.Error(err))
		} else {
			log.Error(ctx, "failed to update user last seen", zap.Error(err))
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
