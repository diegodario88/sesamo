package user

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockUserRepository *MockUserRepository
	userService        UserService
}

func (serviceTestSuite *ServiceTestSuite) SetupTest() {
	serviceTestSuite.mockUserRepository = new(MockUserRepository)
	serviceTestSuite.userService = UserService{
		Repository: serviceTestSuite.mockUserRepository,
	}
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) TestAuthenticateUserByEmailPassword() {
	loginUserPayload := LoginUserPayload{
		Password: "password123",
		Email:    "test@example.com",
	}

	correctUser := &UserEntity{
		Email: loginUserPayload.Email,
	}

	encondedHash, errHash := correctUser.HashPassword(loginUserPayload.Password)
	suite.NoError(errHash)

	correctUser.PasswordHash = &encondedHash

	suite.mockUserRepository.On("FindUserByEmail", loginUserPayload.Email).Return(correctUser, nil)

	result, err := suite.userService.authenticateUserByEmailPassword(loginUserPayload)

	suite.NoError(err)
	suite.NotNil(result)
	suite.mockUserRepository.AssertExpectations(suite.T())
}
