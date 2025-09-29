package usersgrpc_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository/mock"
	"github.com/CrispyCl/TestGolangUsers/internal/service/users"
	usersgrpc "github.com/CrispyCl/TestGolangUsers/internal/transport/grpc/users"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	usersv1 "github.com/CrispyCl/testprotos/gen/go/users"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ctx       = context.Background()
	zapLogger logger.Logger
)

func newServer(repo *mock.UserRepository) *usersgrpc.GRPCServer {
	serv := users.NewUserService(zapLogger, repo)
	return usersgrpc.NewServer(serv)
}

func TestMain(m *testing.M) {
	var err error
	zapLogger, err = logger.NewZapLogger("users_server_test", "test")
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	ctx = context.WithValue(ctx, logger.LoggerKey, zapLogger)

	os.Exit(m.Run())
}

func TestGRPCServer_Register(t *testing.T) {
	t.Run("empty email", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.Register(ctx, &usersv1.RegisterRequest{Email: "", Password: "111"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty password", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.Register(ctx, &usersv1.RegisterRequest{Email: "user@mail.com", Password: ""})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("user exists", func(t *testing.T) {
		m := mock.NewUserRepository()
		_, err := m.Create(ctx, "user@mail.com", []byte("pass1"), models.UserRoleUser)
		assert.NoError(t, err)

		s := newServer(m)
		_, err = s.Register(ctx, &usersv1.RegisterRequest{Email: "user@mail.com", Password: "pass1", Role: usersv1.UserRole_USER_ROLE_USER})
		assert.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.Register(ctx, &usersv1.RegisterRequest{Email: "user@mail.com", Password: "pass1", Role: usersv1.UserRole_USER_ROLE_USER})
		assert.NoError(t, err)
	})
}

func TestGRPCServer_CheckPassword(t *testing.T) {
	t.Run("empty email", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "", Password: "111"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("empty password", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "user@mail.com", Password: ""})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid email", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "absent@mail.com", Password: "wrong"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid password", func(t *testing.T) {
		m := mock.NewUserRepository()
		_, err := users.NewUserService(zapLogger, m).Create(ctx, "user@mail.com", "pass1", models.UserRoleUser)
		assert.NoError(t, err)

		s := newServer(m)
		_, err = s.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "user@mail.com", Password: "wrong"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.Register(ctx, &usersv1.RegisterRequest{Email: "user@mail.com", Password: "pass1", Role: usersv1.UserRole_USER_ROLE_USER})
		assert.NoError(t, err)

		resp, err := s.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "user@mail.com", Password: "pass1"})
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

func TestGRPCServer_GetByEmail(t *testing.T) {
	t.Run("empty email", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.GetUserByEmail(ctx, &usersv1.GetUserByEmailRequest{Email: ""})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found", func(t *testing.T) {
		s := newServer(mock.NewUserRepository())
		_, err := s.GetUserByEmail(ctx, &usersv1.GetUserByEmailRequest{Email: "user@mail.com"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		m := mock.NewUserRepository()
		_, err := m.Create(ctx, "user@mail.com", []byte("hash1"), models.UserRoleUser)
		assert.NoError(t, err)

		s := newServer(m)
		resp, err := s.GetUserByEmail(ctx, &usersv1.GetUserByEmailRequest{Email: "user@mail.com"})
		assert.NoError(t, err)
		assert.Equal(t, "user@mail.com", resp.Email)
		assert.Equal(t, usersv1.UserRole_USER_ROLE_USER, resp.Role)
	})
}
