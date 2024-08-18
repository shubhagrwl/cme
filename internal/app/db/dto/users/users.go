package users

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

const (
	TABLE_NAME          = "users"
	COLUMN_ID           = "id"
	COLUMN_FIRST_NAME   = "first_name"
	COLUMN_LAST_NAME    = "last_name"
	COLUMN_EMAIL        = "email"
	COLUMN_PASSWORD     = "password"
	COLUMN_COUNTRY_CODE = "country_code"
	COLUMN_PHONE        = "phone"
	COLUMN_ADDRESS      = "address"
	COLUMN_STATUS       = "status"
	COLUMN_CREATED_AT   = "created_at"
)

type Users struct {
	ID          gocql.UUID `json:"id" db:"id"` // Cassandra typically uses UUIDs for unique identifiers
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"password,omitempty" db:"password"`
	CountryCode string     `json:"country_code" db:"country_code"`
	Phone       string     `json:"phone" db:"phone"`
	Address     string     `json:"address" db:"address"`
	Status      bool       `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Metadata for gocqlx
var UsersMetadata = table.Metadata{
	Name: TABLE_NAME,
	Columns: []string{
		COLUMN_ID,
		COLUMN_FIRST_NAME,
		COLUMN_LAST_NAME,
		COLUMN_EMAIL,
		COLUMN_PASSWORD,
		COLUMN_COUNTRY_CODE,
		COLUMN_PHONE,
		COLUMN_ADDRESS,
		COLUMN_STATUS,
		COLUMN_CREATED_AT,
	},
	PartKey: []string{COLUMN_ID},         // Partition key(s)
	SortKey: []string{COLUMN_CREATED_AT}, // Clustering key(s), if needed
}

func (u Users) ValidateRegisterDetails() error {
	if u.FirstName == "" {
		return fmt.Errorf("first_name can't be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("email can't be empty")
	}
	if u.Password == "" {
		return fmt.Errorf("password can't be empty")
	}
	return nil
}

func CreateUserTableCQL() string {
	return `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY,
        first_name text,
        last_name text,
        email text,
        password text,
        country_code text,
        phone text,
        address text,
        status boolean,
        created_at timestamp
    );
    `
}

func CreateUserIndexesCQL() []string {
	return []string{
		"CREATE INDEX IF NOT EXISTS ON users (email);",
		"CREATE INDEX IF NOT EXISTS ON users (phone);",
	}
}
