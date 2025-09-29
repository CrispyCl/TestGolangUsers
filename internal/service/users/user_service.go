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
	log := s.getLogger(ctx).With(zap.String("op", op), zap.String("email", email), zap.String("role", string(role)))

	log.Debug(ctx, "user creation start")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to generate password hash: %w", op, err)
	}
	log.Debug(ctx, "password hash generated")

	id, err := s.userRepo.Create(ctx, email, passHash, role)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Debug(ctx, "user created in repository", zap.Int64("id", id))

	return id, nil
}

func (s *UserService) CheckPassword(ctx context.Context, email, password string) (bool, error) {
	const op = "service.user.CheckPassword"
	log := s.getLogger(ctx).With(zap.String("op", op), zap.String("email", email))

	log.Debug(ctx, "password check start")

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			log.Debug(ctx, "user with this email not found", zap.Error(err))
			return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Debug(ctx, "user found")

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Debug(ctx, "password and hash mismatched", zap.Error(err))
			return false, fmt.Errorf("%s: %w", op, service.ErrInvalidCredentials)
		}
		return false, fmt.Errorf("%s: failed to compare password: %w", op, err)
	}
	log.Debug(ctx, "password checked")

	return true, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "service.user.GetByEmail"
	log := s.getLogger(ctx).With(zap.String("op", op), zap.String("email", email))

	log.Debug(ctx, "getting by email start")

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Debug(ctx, "user found")

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const op = "service.user.GetByID"
	log := s.getLogger(ctx).With(zap.String("op", op), zap.Int64("id", id))

	log.Debug(ctx, "getting by id start")

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Debug(ctx, "user found")

	return user, nil
}

func (s *UserService) UpdateLastSeen(ctx context.Context, id int64) error {
	const op = "service.user.UpdateLastSeen"
	log := s.getLogger(ctx).With(zap.String("op", op), zap.Int64("id", id))

	log.Debug(ctx, "updating last seen start")

	_, err := s.userRepo.UpdateLastSeen(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Debug(ctx, "last seen updated")

	return nil
}

func (s *UserService) getLogger(ctx context.Context) logger.Logger {
	log, ok := logger.GetLoggerFromContext(ctx)
	if !ok || log == nil {
		return s.log
	}
	return log
}
