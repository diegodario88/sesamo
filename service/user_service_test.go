package service_test

import "github.com/diegodario88/sesamo/types"

func (suite *ServiceTestSuite) TestAuthenticateUserByEmailPassword() {
	email := "test@example.com"
	password := "password123"

	correctUser := &types.User{
		Email: email,
	}

	encondedHash, errHash := correctUser.HashPassword(password)
	suite.NoError(errHash)

	correctUser.PasswordHash = &encondedHash

	suite.mockUserStore.On("FindUserByEmail", email).Return(correctUser, nil)

	result, err := suite.userService.AuthenticateUserByEmailPassword(email, password)

	suite.NoError(err)
	suite.NotNil(result)
	suite.mockUserStore.AssertExpectations(suite.T())
}
