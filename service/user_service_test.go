package service_test

func (suite *ServiceTestSuite) TestAuthenticateUserByEmailPassword() {
	email := "test@example.com"
	password := "password123"

	suite.mockUser.On("GetEmail").Return(email)
	suite.mockUser.On("CheckPassword", password).Return(true, nil)
	suite.mockUserStore.On("FindUserByEmail", email).Return(suite.mockUser, nil)

	result, err := suite.userService.AuthenticateUserByEmailPassword(email, password)

	suite.NoError(err)
	suite.Equal(suite.mockUser, result)
	suite.mockUserStore.AssertExpectations(suite.T())
	suite.mockUser.AssertExpectations(suite.T())
}
