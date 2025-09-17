package pg_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/repository/pg"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	db       *sqlx.DB
	userRepo *pg.UserRepository
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "users",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}
	defer func() {
		if err := postgresC.Terminate(ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}()

	host, err := postgresC.Host(ctx)
	if err != nil {
		log.Fatalf("could not get container host: %v", err)
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("could not get mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/users?sslmode=disable", host, port.Port())
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	migrator, err := migrate.New(migrationsPath(), dsn)
	if err != nil {
		log.Fatalf("could not init migrator: %v", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run migrations: %v", err)
	}

	userRepo = pg.NewUserRepository(db)

	os.Exit(m.Run())
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	t.Run("create and get user", func(t *testing.T) {
		id, err := userRepo.Create(ctx, "user1@mail.com", []byte("hash1"), models.UserRoleUser)
		assert.NoError(t, err)
		assert.True(t, id > 0)

		user, err := userRepo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "user1@mail.com", user.Email)
		assert.Equal(t, []byte("hash1"), user.PassHash)
		assert.Equal(t, models.UserRoleUser, user.Role)
		assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Minute)
	})

	t.Run("get by email", func(t *testing.T) {
		user, err := userRepo.GetByEmail(ctx, "user1@mail.com")
		assert.NoError(t, err)
		assert.Equal(t, models.UserRoleUser, user.Role)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := userRepo.Create(ctx, "user1@mail.com", []byte("hash2"), models.UserRoleAdmin)
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserExists)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := userRepo.GetByEmail(ctx, "absent@mail.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := userRepo.Create(ctx, "user2@mail.com", []byte("hash3"), "INVALID")
		assert.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidUserRole)
	})

	t.Run("multiple users with different roles", func(t *testing.T) {
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

func migrationsPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory: %v", err)
	}

	for {
		migrations := filepath.Join(dir, "migrations")
		if fi, err := os.Stat(migrations); err == nil && fi.IsDir() {
			return "file://" + filepath.ToSlash(migrations)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("could not find migrations directory")
		}
		dir = parent
	}
}
