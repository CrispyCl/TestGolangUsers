package usersgrpc

import (
	"context"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	usersv1 "github.com/CrispyCl/testprotos/gen/go/users"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	usersv1.UnimplementedUsersServer
	userServ UserService
	log      logger.Logger
}

type UserService interface {
	Create(ctx context.Context, email, password string, role models.UserRole) (int64, error)
	CheckPassword(ctx context.Context, email, password string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	UpdateLastSeen(ctx context.Context, id int64) error
}

func Register(gRPCServer *grpc.Server, users UserService) {
	usersv1.RegisterUsersServer(gRPCServer, &GRPCServer{userServ: users})
}
