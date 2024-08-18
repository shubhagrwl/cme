package user

import (
	"context"
	"errors"
	"fmt"

	"cme/internal/app/db"
	users_DBModels "cme/internal/app/db/dto/users"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
)

type IUserRepository interface {
	GetUser(ctx context.Context, whr map[string]interface{}) (users_DBModels.Users, error)
	CreateUser(ctx context.Context, User users_DBModels.Users) error
}

type UserRepository struct {
	DBService *db.DBService
	table     table.Table
}

func NewUserRepository(dbService *db.DBService) IUserRepository {
	userTable := table.New(users_DBModels.UsersMetadata) // Ensure UsersMetadata is correctly defined
	return &UserRepository{
		DBService: dbService,
		table:     *userTable,
	}
}

func (u *UserRepository) CreateUser(ctx context.Context, User users_DBModels.Users) error {
	stmt, names := u.table.Insert()

	q := gocqlx.Query(u.DBService.GetSession().Query(stmt), names).BindStruct(User)
	if err := q.ExecRelease(); err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) GetUser(ctx context.Context, whr map[string]interface{}) (users_DBModels.Users, error) {
	var user users_DBModels.Users

	// Build the CQL statement with WHERE conditions
	stmt, names := qb.Select(u.table.Name()).Where(whrToEq(whr)...).ToCql()

	// If ALLOW FILTERING is necessary, ensure it is used only when justified
	stmt += " ALLOW FILTERING"

	// Prepare and execute the query
	q := gocqlx.Query(u.DBService.GetSession().Query(stmt), names).BindMap(whr)

	// Fetch the result
	if err := q.GetRelease(&user); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			// Explicitly return an error indicating that the user was not found
			return user, fmt.Errorf("user not found: %v", whr)
		}
		return user, err
	}

	return user, nil
}

func whrToEq(whr map[string]interface{}) []qb.Cmp {
	var cmps []qb.Cmp
	for key := range whr {
		cmps = append(cmps, qb.Eq(key))
	}
	return cmps
}
