package usersgrpc_integration_test

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/CrispyCl/TestGolangUsers/internal/repository/mock"
	"github.com/CrispyCl/TestGolangUsers/internal/service/users"
	usersgrpc "github.com/CrispyCl/TestGolangUsers/internal/transport/grpc/users"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	usersv1 "github.com/CrispyCl/testprotos/gen/go/users"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	grpcServer *grpc.Server
	serverAddr string
	zapLogger  logger.Logger
	client     usersv1.UsersClient
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
)

func TestMain(m *testing.M) {
	var err error
	zapLogger, err = logger.NewZapLogger("users_server_test", "test")
	if err != nil {
		log.Fatalf("could not create zap logger: %v", err)
	}

	cancelCtx, cancelFunc = context.WithCancel(context.Background())

	repo := mock.NewUserRepository()
	userServ := users.NewUserService(zapLogger, repo)

	unaryInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = context.WithValue(ctx, logger.LoggerKey, zapLogger)
		return handler(ctx, req)
	}

	grpcServer = grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))
	srv := usersgrpc.NewServer(userServ)
	usersv1.RegisterUsersServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	serverAddr = lis.Addr().String()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	client = usersv1.NewUsersClient(conn)

	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	conn.Close()
	grpcServer.Stop()
	cancelFunc()

	os.Exit(code)
}

func TestIntegrationGRPCServer_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("register users", func(t *testing.T) {
		usersToCreate := []struct {
			email string
			pass  string
			role  usersv1.UserRole
		}{
			{"user1@mail.com", "pass1", usersv1.UserRole_USER_ROLE_USER},
			{"manager@mail.com", "pass2", usersv1.UserRole_USER_ROLE_MANAGER},
			{"admin@mail.com", "pass3", usersv1.UserRole_USER_ROLE_ADMIN},
		}

		for _, u := range usersToCreate {
			resp, err := client.Register(ctx, &usersv1.RegisterRequest{Email: u.email, Password: u.pass, Role: u.role})
			assert.NoError(t, err)
			assert.NotZero(t, resp.UserId)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := client.Register(ctx, &usersv1.RegisterRequest{Email: "user1@mail.com", Password: "pass4", Role: usersv1.UserRole_USER_ROLE_USER})
		assert.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})
}

func TestIntegrationGRPCServer_CheckPassword(t *testing.T) {
	ctx := context.Background()

	_, err := client.Register(ctx, &usersv1.RegisterRequest{Email: "user_for_check_password@mail.com", Password: "pass111", Role: usersv1.UserRole_USER_ROLE_USER})
	assert.NoError(t, err)

	t.Run("wrong password", func(t *testing.T) {
		_, err = client.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "user_for_check_password@mail.com", Password: "wrong"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("wrong email", func(t *testing.T) {
		_, err = client.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "absent@mail.com", Password: "pass111"})
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		resp, err := client.CheckPassword(ctx, &usersv1.CheckPasswordRequest{Email: "user_for_check_password@mail.com", Password: "pass111"})
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})
}

func TestIntegrationGRPCServer_GetByEmail(t *testing.T) {
	ctx := context.Background()

	_, err := client.Register(ctx, &usersv1.RegisterRequest{Email: "user_for_get_by_email@mail.com", Password: "pass111", Role: usersv1.UserRole_USER_ROLE_USER})
	assert.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		_, err := client.GetUserByEmail(ctx, &usersv1.GetUserByEmailRequest{Email: "absent@mail.com"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		resp, err := client.GetUserByEmail(ctx, &usersv1.GetUserByEmailRequest{Email: "user_for_get_by_email@mail.com"})
		assert.NoError(t, err)
		assert.Equal(t, usersv1.UserRole_USER_ROLE_USER, resp.Role)
	})
}
