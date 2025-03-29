package user

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	var newUserRepository = UserRepository{
		db: db,
	}

	return &newUserRepository
}

func (repo *UserRepository) InsertUser(user *UserEntity) (*UserEntity, error) {
	var insertResult UserEntity
	sqlQuery := `INSERT INTO users (first_name, last_name, email, password_hash) 
                          values ($1, $2, $3, $4) returning *`

	err := repo.db.Get(
		&insertResult,
		sqlQuery,
		user.FirstName,
		user.LastName,
		user.Email,
		user.PasswordHash,
	)

	if err != nil {
		return nil, fmt.Errorf("InsertUser: %w", err)
	}

	return &insertResult, nil
}

func (repo *UserRepository) CountUsers() (int, error) {
	var countResult int
	sqlQuery := `SELECT COUNT(*) FROM users`
	err := repo.db.QueryRow(sqlQuery).Scan(&countResult)

	return countResult, err
}

func (repo *UserRepository) FindUserByEmail(email string) (*UserEntity, error) {
	var foundResult UserEntity
	sqlQuery := `SELECT * FROM users u WHERE u.email = $1`

	err := repo.db.Get(&foundResult, sqlQuery, email)

	if err != nil {
		return nil, fmt.Errorf("FindUserByEmail: %w", err)
	}

	return &foundResult, nil
}

func (repo *UserRepository) FindUserById(id string) (*UserEntity, error) {
	var foundResult UserEntity
	sqlQuery := `SELECT * FROM users u WHERE u.id = $1`

	err := repo.db.Get(&foundResult, sqlQuery, id)

	if err != nil {
		return nil, fmt.Errorf("FindUserById: %w", err)
	}

	return &foundResult, nil
}

func (repo *UserRepository) FindAllUsers() ([]UserEntity, error) {
	var foundResult []UserEntity
	sqlQuery := `SELECT * FROM users`

	err := repo.db.Select(&foundResult, sqlQuery)

	if err != nil {
		return nil, fmt.Errorf("FindUserById: %w", err)
	}

	return foundResult, nil
}

func (repo *UserRepository) GetRoles(userId string) ([]string, error) {
	query := `
		SELECT r.name 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := repo.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (repo *UserRepository) HasAccess(userId string, permission string) (bool, error) {
	var hasPermission bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE u.id = $1 AND p.name = $2
		)
	`, userId, permission).Scan(&hasPermission)

	return hasPermission, err
}

func (repo *UserRepository) GetUserOrganizations(userId string) ([]OrganizationEntity, error) {
	organizations := []OrganizationEntity{}

	query := `
    SELECT DISTINCT
        o.id,
        o.external_company_id,
        o.external_head_office_id,
        o.name,
        o.description,
        o.created_at,
        o.updated_at
    FROM
        organizations o
        JOIN user_roles ur ON o.id = ur.organization_id
    WHERE
        ur.user_id = $1
    UNION
    SELECT DISTINCT
        o.id,
        o.external_company_id,
        o.external_head_office_id,
        o.name,
        o.description,
        o.created_at,
        o.updated_at
    FROM
        organizations o
    WHERE
        EXISTS (
            SELECT
                1
            FROM
                user_roles ur
                JOIN roles r ON ur.role_id = r.id
            WHERE
                ur.user_id = $1
                AND r.scope = 'global');`

	err := repo.db.Select(&organizations, query, userId)
	return organizations, err
}

func (repo *UserRepository) GetUserBranches(userId string, orgId string) ([]BranchEntity, error) {
	branches := []BranchEntity{}

	query := `
		SELECT DISTINCT 
			b.id, 
			b.external_office_id, 
			b.cnpj, 
			b.organization_id, 
			b.name, 
			b.description, 
			b.is_warehouse, 
			b.created_at, 
			b.updated_at
		FROM branches b
		WHERE b.organization_id = $1 AND (
			EXISTS (
				SELECT 1 FROM user_roles ur 
				WHERE ur.user_id = $2 AND ur.branch_id = b.id
			) OR
			EXISTS (
				SELECT 1 FROM user_roles ur
				JOIN roles r ON ur.role_id = r.id
				WHERE ur.user_id = $2 AND ur.organization_id = $1 AND r.scope = 'organization'
			) OR
			EXISTS (
				SELECT 1 FROM user_roles ur
				JOIN roles r ON ur.role_id = r.id
				WHERE ur.user_id = $2 AND r.scope = 'global'
			)
		);
	`
	err := repo.db.Select(&branches, query, orgId, userId)
	return branches, err
}

func (repo *UserRepository) FindOrganizationUsers(orgID string) ([]UserEntity, error) {
	users := []UserEntity{}

	query := `
        SELECT DISTINCT u.* 
        FROM users u
        JOIN user_roles ur ON u.id = ur.user_id
        WHERE ur.organization_id = $1 OR 
            EXISTS (
                SELECT 1 FROM user_roles ur2
                JOIN roles r ON ur2.role_id = r.id
                WHERE ur2.user_id = u.id AND r.scope = 'global'
            )
    `

	err := repo.db.Select(&users, query, orgID)
	return users, err
}

func (repo *UserRepository) FindOrganizationUserByID(
	orgID string,
	userID string,
) (*UserEntity, error) {
	var user UserEntity

	query := `
        SELECT u.* 
        FROM users u
        WHERE u.id = $1 AND (
            EXISTS (
                SELECT 1 FROM user_roles ur
                WHERE ur.user_id = u.id AND ur.organization_id = $2
            ) OR
            EXISTS (
                SELECT 1 FROM user_roles ur
                JOIN roles r ON ur.role_id = r.id
                WHERE ur.user_id = u.id AND r.scope = 'global'
            )
        )
    `

	err := repo.db.Get(&user, query, userID, orgID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepository) FindBranchUsers(orgID string, branchID string) ([]UserEntity, error) {
	users := []UserEntity{}

	query := `
        SELECT DISTINCT u.* 
        FROM users u
        WHERE 
            EXISTS (
                SELECT 1 FROM user_roles ur
                WHERE ur.user_id = u.id AND ur.branch_id = $2
            ) OR
            EXISTS (
                SELECT 1 FROM user_roles ur
                JOIN roles r ON ur.role_id = r.id
                WHERE ur.user_id = u.id AND ur.organization_id = $1 AND r.scope = 'organization'
            ) OR
            EXISTS (
                SELECT 1 FROM user_roles ur
                JOIN roles r ON ur.role_id = r.id
                WHERE ur.user_id = u.id AND r.scope = 'global'
            )
    `

	err := repo.db.Select(&users, query, orgID, branchID)
	return users, err
}

func (repo *UserRepository) GetOrganizationBranches(orgID string) ([]BranchEntity, error) {
	branches := []BranchEntity{}

	query := `
		SELECT 
			id, 
			external_office_id, 
			cnpj, 
			organization_id, 
			name, 
			description, 
			is_warehouse, 
			created_at, 
			updated_at
		FROM branches
		WHERE organization_id = $1
		ORDER BY name
	`

	err := repo.db.Select(&branches, query, orgID)
	return branches, err
}
