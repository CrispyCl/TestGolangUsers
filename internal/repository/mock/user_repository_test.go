package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/repository/mock"
	"github.com/stretchr/testify/assert"
)

var (
	userRepo = mock.NewUserRepository()
	ctx      = context.Background()
)

func TestUserMockRepository_CreateAndGet(t *testing.T) {
	t.Run("create and get user", func(t *testing.T) {
		id, err := userRepo.Create(ctx, "user_for_create1@mail.com", []byte("hash1"), models.UserRoleUser)
		assert.NoError(t, err)
		assert.True(t, id > 0)

		user, err := userRepo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "user_for_create1@mail.com", user.Email)
		assert.Equal(t, []byte("hash1"), user.PassHash)
		assert.Equal(t, models.UserRoleUser, user.Role)
		assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Minute)
	})

	t.Run("get by email", func(t *testing.T) {
		user, err := userRepo.GetByEmail(ctx, "user_for_create1@mail.com")
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleUser, user.Role)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := userRepo.Create(ctx, "user_for_create1@mail.com", []byte("hash2"), models.UserRoleManager)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserExists)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := userRepo.GetByEmail(ctx, "absent@mail.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := userRepo.Create(ctx, "user_for_create2@mail.com", []byte("hash3"), "INVALID")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidUserRole)
	})

	t.Run("multiple users with different role", func(t *testing.T) {
		id1, err := userRepo.Create(ctx, "manager@mail.com", []byte("hash4"), models.UserRoleManager)
		assert.NoError(t, err)
		id2, err := userRepo.Create(ctx, "admin@mail.com", []byte("hash5"), models.UserRoleAdmin)
		assert.NoError(t, err)

		user1, err := userRepo.GetByID(ctx, id1)
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleManager, user1.Role)

		user2, err := userRepo.GetByID(ctx, id2)
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleAdmin, user2.Role)
	})
}

func TestUserMockRepository_UpdateLastSeen(t *testing.T) {
	t.Run("update last seen for existing user", func(t *testing.T) {
		id, err := userRepo.Create(ctx, "user_for_update1@mail.com", []byte("hash1"), models.UserRoleUser)
		assert.NoError(t, err)

		userBefore, err := userRepo.GetByID(ctx, id)
		lastSeen := userBefore.LastSeen
		assert.NoError(t, err)

		time.Sleep(100 * time.Microsecond)

		updatedUser, err := userRepo.UpdateLastSeen(ctx, id)
		assert.NoError(t, err)

		assert.True(t, updatedUser.LastSeen.After(lastSeen))
		assert.WithinDuration(t, time.Now(), updatedUser.LastSeen, 10*time.Second)

		assert.Equal(t, userBefore.ID, updatedUser.ID)
		assert.Equal(t, userBefore.Email, updatedUser.Email)
		assert.Equal(t, userBefore.PassHash, updatedUser.PassHash)
		assert.Equal(t, userBefore.Role, updatedUser.Role)
		assert.Equal(t, userBefore.CreatedAt, updatedUser.CreatedAt)
	})

	t.Run("update last seen for non-existing user", func(t *testing.T) {
		_, err := userRepo.UpdateLastSeen(ctx, 999_999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("update last seen multiple times", func(t *testing.T) {
		id, err := userRepo.Create(ctx, "user_for_update2@mail.com", []byte("hash2"), models.UserRoleUser)
		assert.NoError(t, err)

		var lastSeen time.Time
		for i := range 5 {
			time.Sleep(10 * time.Microsecond)
			updatedUser, err := userRepo.UpdateLastSeen(ctx, id)
			assert.NoError(t, err)

			if i > 0 {
				assert.True(t, updatedUser.LastSeen.After(lastSeen))
			}
			lastSeen = updatedUser.LastSeen
		}
	})
}
