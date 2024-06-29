package store

import (
	"github.com/diegodario88/sesamo/types"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) FindUserByEmail(email string) (*types.User, error) {
	args := m.Called(email)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserStore) InsertUser(user *types.User) (*types.User, error) {
	args := m.Called(user)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserStore) CountUsers() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
