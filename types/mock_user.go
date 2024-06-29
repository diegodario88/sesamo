package types

import "github.com/stretchr/testify/mock"

type MockUser struct {
	mock.Mock
}

func (m *MockUser) CheckPassword(password string) (bool, error) {
	args := m.Called(password)
	return args.Bool(0), args.Error(1)
}
