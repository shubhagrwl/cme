package dto

import (
	messages "cme/internal/app/db/dto/message"
	"cme/internal/app/db/dto/users"
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

func RunTableMigration(session *gocql.Session) {

	// Run the table creation CQL
	tableCQL := users.CreateUserTableCQL()
	if err := session.Query(tableCQL).Exec(); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	fmt.Println("Table 'users' created successfully.")

	tableMessageCQL := messages.CreateMessagesTableCQL()
	if err := session.Query(tableMessageCQL).Exec(); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	fmt.Println("Table 'messages' created successfully.")

	// Run the index creation CQL
	indexesCQL := users.CreateUserIndexesCQL()
	for _, indexCQL := range indexesCQL {
		if err := session.Query(indexCQL).Exec(); err != nil {
			log.Fatalf("Error creating index: %v", err)
		}
	}

	// Run the index creation CQL
	messagesCQL := messages.CreateMessageIndexesCQL()
	for _, indexCQL := range messagesCQL {
		if err := session.Query(indexCQL).Exec(); err != nil {
			log.Fatalf("Error creating index: %v", err)
		}
	}
}
