package converts

import (
	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	usersv1 "github.com/CrispyCl/testprotos/gen/go/users"
)

var apiToModelRole = map[usersv1.UserRole]models.UserRole{
	usersv1.UserRole_USER_ROLE_ADMIN:       models.UserRoleAdmin,
	usersv1.UserRole_USER_ROLE_MANAGER:     models.UserRoleManager,
	usersv1.UserRole_USER_ROLE_USER:        models.UserRoleUser,
	usersv1.UserRole_USER_ROLE_UNSPECIFIED: models.UserRoleUnspecified,
}

var modelToApiRole = map[models.UserRole]usersv1.UserRole{
	models.UserRoleAdmin:       usersv1.UserRole_USER_ROLE_ADMIN,
	models.UserRoleManager:     usersv1.UserRole_USER_ROLE_MANAGER,
	models.UserRoleUser:        usersv1.UserRole_USER_ROLE_USER,
	models.UserRoleUnspecified: usersv1.UserRole_USER_ROLE_UNSPECIFIED,
}

func RoleFromApiToModel(role usersv1.UserRole) models.UserRole {
	if r, ok := apiToModelRole[role]; ok {
		return r
	}
	return models.UserRoleUnspecified
}

func RoleFromModelToApi(role models.UserRole) usersv1.UserRole {
	if r, ok := modelToApiRole[role]; ok {
		return r
	}
	return usersv1.UserRole_USER_ROLE_UNSPECIFIED
}
