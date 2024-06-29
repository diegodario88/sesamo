package store

import (
	"github.com/diegodario88/sesamo/types"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) FindUserByEmail(email string) (*types.MockUser, error) {
	args := m.Called(email)
	return args.Get(0).(*types.MockUser), args.Error(1)
}
