package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/service"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, email string, passHash []byte, role models.UserRole) (id int64, err error)
	GetByID(ctx context.Context, id int64) (user *models.User, err error)
	GetByEmail(ctx context.Context, email string) (user *models.User, err error)
	UpdateLastSeen(ctx context.Context, id int64) (user *models.User, err error)
}

type UserService struct {
	log      *slog.Logger
	userRepo UserRepository
}

func NewUserService(log *slog.Logger, userRepo UserRepository) *UserService {

	return &UserService{
		log:      log,
		userRepo: userRepo,
	}
}

func (s *UserService) Create(ctx context.Context, email, password string, role models.UserRole) (int64, error) {
	const op = "service.user.Create"
	log := s.log.With(slog.String("op", op), slog.String("email", email), slog.String("role", string(role)))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.userRepo.Create(ctx, email, passHash, role)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			log.Info("user with this email is exists", logger.Err(err))
		} else if errors.Is(err, repository.ErrInvalidUserRole) {
			log.Error("invalid user role", logger.Err(err))
		} else {
			log.Error("failed to save user", logger.Err(err))
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *UserService) CheckPassword(ctx context.Context, email, password string) (bool, error) {
	const op = "service.user.CheckPassword"
	log := s.log.With(slog.String("op", op), slog.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Info("user not found", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
		}

		log.Error("failed to get user", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
	}

	return true, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "service.user.GetByEmail"
	log := s.log.With(slog.String("op", op), slog.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Info("user not found", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to get user", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const op = "service.user.GetByID"
	log := s.log.With(slog.String("op", op), slog.Int64("id", id))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Info("user not found", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to get user", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) UpdateLastSeen(ctx context.Context, id int64) error {
	const op = "service.user.UpdateLastSeen"
	log := s.log.With(slog.String("op", op), slog.String("id", strconv.FormatInt(id, 10)))

	_, err := s.userRepo.UpdateLastSeen(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Error("user with this id not found", logger.Err(err))
		} else {
			log.Error("failed to update user last seen", logger.Err(err))
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
