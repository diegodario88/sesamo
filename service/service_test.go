package service_test

import (
	"testing"

	"github.com/diegodario88/sesamo/service"
	"github.com/diegodario88/sesamo/store"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	mockUserStore *store.MockUserStore
	userService   service.UserService
}

func (serviceTestSuite *ServiceTestSuite) SetupTest() {
	serviceTestSuite.mockUserStore = new(store.MockUserStore)
	serviceTestSuite.userService = service.UserService{
		Store: serviceTestSuite.mockUserStore,
	}
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
