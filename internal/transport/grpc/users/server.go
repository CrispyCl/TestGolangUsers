package usersgrpc

import (
	"context"
	"errors"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/repository"
	"github.com/CrispyCl/TestGolangUsers/internal/service"
	"github.com/CrispyCl/TestGolangUsers/internal/transport/grpc/converts"
	usersv1 "github.com/CrispyCl/testprotos/gen/go/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	usersv1.UnimplementedUsersServer
	userServ UserService
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

func (s *GRPCServer) Register(ctx context.Context, req *usersv1.RegisterRequest) (*usersv1.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if req.Role != usersv1.UserRole_USER_ROLE_USER {
		// TODO: add permissions check
	}

	id, err := s.userServ.Create(ctx, req.Email, req.Password, converts.RoleFromApiToModel(req.Role))
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &usersv1.RegisterResponse{UserId: id}, nil
}

func (s *GRPCServer) CheckPassword(ctx context.Context, req *usersv1.CheckPasswordRequest) (*usersv1.SuccessResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	ok, err := s.userServ.CheckPassword(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to check password")
	}

	return &usersv1.SuccessResponse{Success: ok}, nil
}

func (s *GRPCServer) GetUserByEmail(ctx context.Context, req *usersv1.GetUserByEmailRequest) (*usersv1.GetUserResponse, error) {
	// TODO: add permissions check
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := s.userServ.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user with this email not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user by email")
	}

	return &usersv1.GetUserResponse{UserId: user.ID, Email: user.Email, Role: converts.RoleFromModelToApi(user.Role)}, nil
}

func (s *GRPCServer) GetUserMy(ctx context.Context, req *usersv1.EmptyRequest) (*usersv1.GetUserResponse, error) {
	// TODO: implement GetUserMy method with getting current user from JWT
	return nil, status.Error(codes.Unimplemented, "method GetUserMy not implemented")
}

func (s *GRPCServer) UpdateLastSeen(ctx context.Context, req *usersv1.EmptyRequest) (*usersv1.EmptyResponse, error) {
	// TODO: implement UpdateLastSeen method with getting current user from JWT
	return nil, status.Error(codes.Unimplemented, "method UpdateLastSeen not implemented")
}
