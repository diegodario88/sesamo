package service_test

import "github.com/diegodario88/sesamo/types"

func (suite *ServiceTestSuite) TestAuthenticateUserByEmailPassword() {
	email := "test@example.com"
	password := "password123"
	hashedPassword := "$argon2id$v=19$m=65536,t=3,p=2$Woo1mErn1s7AHf96ewQ8Uw$D4TzIwGO4XD2buk96qAP+Ed2baMo/KbTRMqXX00wtsU"

	correctUser := &types.User{
		Email:        email,
		PasswordHash: &hashedPassword,
	}

	suite.mockUserStore.On("FindUserByEmail", email).Return(correctUser, nil)

	result, err := suite.userService.AuthenticateUserByEmailPassword(email, password)

	suite.NoError(err)
	suite.NotNil(result)
	suite.mockUserStore.AssertExpectations(suite.T())
}
