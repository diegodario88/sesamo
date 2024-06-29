package store_test

import (
	"database/sql"
	"strings"

	"github.com/diegodario88/sesamo/store"
	"github.com/diegodario88/sesamo/types"
)

func (storeTestSuite *StoreTestSuite) TestInsertUser() {
	passwordHash := "test_pw"

	user := types.User{
		Email:        "insert@test.com",
		PasswordHash: &passwordHash,
	}

	userStore := store.NewUserStore(storeTestSuite.db)

	before, err := userStore.CountUsers()
	storeTestSuite.NoError(err)

	actual, err := userStore.InsertUser(&user)
	storeTestSuite.NoError(err)

	after, err := userStore.CountUsers()
	storeTestSuite.NoError(err)

	storeTestSuite.Greater(actual.ID, 0)
	storeTestSuite.Equal(user.Email, actual.Email)
	storeTestSuite.Equal(user.PasswordHash, actual.PasswordHash)
	storeTestSuite.Equal(before+1, after)
}

func (storeTestSuite *StoreTestSuite) TestFindUserByEmail() {
	passwordHash := "test_pw"

	newUser := types.User{
		Email:        "by-email@test.com",
		PasswordHash: &passwordHash,
	}

	userStore := store.NewUserStore(storeTestSuite.db)

	user, err := userStore.InsertUser(&newUser)
	storeTestSuite.NoError(err)

	arrange := []string{user.Email, strings.ToUpper(user.Email), "By-Email@TesT.coM"}

	for _, email := range arrange {
		actual, findErr := userStore.FindUserByEmail(email)
		storeTestSuite.NoError(findErr)
		storeTestSuite.Equal(user.ID, actual.ID)
	}

	actual, err := userStore.FindUserByEmail("non-existtent@test.com")
	storeTestSuite.ErrorIs(err, sql.ErrNoRows)
	storeTestSuite.Nil(actual)
}
