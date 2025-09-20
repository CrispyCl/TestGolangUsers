package users_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/repository/mock"
	"github.com/CrispyCl/TestGolangUsers/internal/service"
	"github.com/CrispyCl/TestGolangUsers/internal/service/users"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var (
	ctx      = context.Background()
	userServ *users.UserService
)

func TestMain(m *testing.M) {
	var log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	var userRepo = mock.NewUserRepository()

	userServ = users.NewUserService(log, userRepo)

	os.Exit(m.Run())
}

func TestUserService_CreateAndGet(t *testing.T) {
	t.Run("create and get", func(t *testing.T) {
		id, err := userServ.Create(ctx, "user_for_create1@mail.com", "pass1", models.UserRoleUser)
		assert.NoError(t, err)

		user, err := userServ.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "user_for_create1@mail.com", user.Email)
		assert.Equal(t, models.UserRoleUser, user.Role)
		assert.WithinDuration(t, time.Now(), user.CreatedAt, 10*time.Second)
		assert.NoError(t, bcrypt.CompareHashAndPassword(user.PassHash, []byte("pass1")))
	})

	t.Run("get by email", func(t *testing.T) {
		user, err := userServ.GetByEmail(ctx, "user_for_create1@mail.com")
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleUser, user.Role)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := userServ.Create(ctx, "user_for_create1@mail.com", "pass2", models.UserRoleManager)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserExists)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := userServ.GetByEmail(ctx, "absent@mail.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := userServ.Create(ctx, "user_for_create2@mail.com", "pass3", "INVALID")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidUserRole)
	})

	t.Run("multiple creation with different roles", func(t *testing.T) {
		id1, err := userServ.Create(ctx, "manager@mail.com", "pass4", models.UserRoleManager)
		assert.NoError(t, err)
		id2, err := userServ.Create(ctx, "admin@mail.com", "pass5", models.UserRoleAdmin)
		assert.NoError(t, err)

		user1, err := userServ.GetByID(ctx, id1)
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleManager, user1.Role)

		user2, err := userServ.GetByID(ctx, id2)
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleAdmin, user2.Role)
	})
}

func TestUserService_UpdateLastSeen(t *testing.T) {
	t.Run("update for existing user", func(t *testing.T) {
		id, err := userServ.Create(ctx, "user_for_update1@mail.com", "pass1", models.UserRoleUser)
		assert.NoError(t, err)

		_, err = userServ.GetByID(ctx, id)
		assert.NoError(t, err)

		time.Sleep(100 * time.Microsecond)

		err = userServ.UpdateLastSeen(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("update for non-existing user", func(t *testing.T) {
		err := userServ.UpdateLastSeen(ctx, 999_999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})
}

func TestUserService_CheckPassword(t *testing.T) {
	t.Run("create and check password", func(t *testing.T) {
		_, err := userServ.Create(ctx, "user_for_password_check1@mail.com", "pass1", models.UserRoleUser)
		assert.NoError(t, err)

		success, err := userServ.CheckPassword(ctx, "user_for_password_check1@mail.com", "pass1")
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("check for non-existing user", func(t *testing.T) {
		_, err := userServ.CheckPassword(ctx, "absent@mail.com", "pass1")
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	})

	t.Run("check with wrong pass", func(t *testing.T) {
		_, err := userServ.CheckPassword(ctx, "user_for_password_check1@mail.com", "pass2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrInvalidCredentials)
	})
}
