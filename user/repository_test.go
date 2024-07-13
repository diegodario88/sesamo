package user

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/diegodario88/sesamo/config"
	"github.com/diegodario88/sesamo/migrations"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByEmail(email string) (*UserEntity, error) {
	args := m.Called(email)
	return args.Get(0).(*UserEntity), args.Error(1)
}

func (m *MockUserRepository) InsertUser(user *UserEntity) (*UserEntity, error) {
	args := m.Called(user)
	return args.Get(0).(*UserEntity), args.Error(1)
}

func (m *MockUserRepository) CountUsers() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

type RepositoryTestSuite struct {
	suite.Suite
	db *sqlx.DB
}

func (repositoryTestSuite *RepositoryTestSuite) SetupTest() {
	testConnString := config.Variables.TestDatabaseUrl
	fmt.Println(testConnString)
	repositoryTestSuite.db = sqlx.MustConnect("postgres", testConnString)

	goose.SetBaseFS(migrations.Files)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(repositoryTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
	if err := goose.Reset(repositoryTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
	if err := goose.Up(repositoryTestSuite.db.DB, "migrations"); err != nil {
		panic(err)
	}
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (repositoryTestSuite *RepositoryTestSuite) TestInsertUser() {
	passwordHash := "test_pw"

	user := UserEntity{
		Email:        "insert@test.com",
		PasswordHash: &passwordHash,
	}

	userRepository := NewUserRepository(repositoryTestSuite.db)

	before, err := userRepository.CountUsers()
	repositoryTestSuite.NoError(err)

	actual, err := userRepository.InsertUser(&user)
	repositoryTestSuite.NoError(err)

	after, err := userRepository.CountUsers()
	repositoryTestSuite.NoError(err)

	repositoryTestSuite.Greater(actual.ID, 0)
	repositoryTestSuite.Equal(user.Email, actual.Email)
	repositoryTestSuite.Equal(user.PasswordHash, actual.PasswordHash)
	repositoryTestSuite.Equal(before+1, after)
}

func (repositoryTestSuite *RepositoryTestSuite) TestFindUserByEmail() {
	passwordHash := "test_pw"

	newUser := UserEntity{
		Email:        "by-email@test.com",
		PasswordHash: &passwordHash,
	}

	userRepository := NewUserRepository(repositoryTestSuite.db)

	user, err := userRepository.InsertUser(&newUser)
	repositoryTestSuite.NoError(err)

	arrange := []string{user.Email, strings.ToUpper(user.Email), "By-Email@TesT.coM"}

	for _, email := range arrange {
		actual, findErr := userRepository.FindUserByEmail(email)
		repositoryTestSuite.NoError(findErr)
		repositoryTestSuite.Equal(user.ID, actual.ID)
	}

	actual, err := userRepository.FindUserByEmail("non-existtent@test.com")
	repositoryTestSuite.ErrorIs(err, sql.ErrNoRows)
	repositoryTestSuite.Nil(actual)
}
