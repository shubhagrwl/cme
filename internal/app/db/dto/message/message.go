package messages

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
)

const (
	TABLE_NAME          = "messages"
	COLUMN_MESSAGE_ID   = "message_id"
	COLUMN_SENDER_ID    = "sender_id"
	COLUMN_MESSAGE_TEXT = "message_text"
	COLUMN_CREATED_AT   = "created_at"
)

type Messages struct {
	MessageID   gocql.UUID `json:"message_id" db:"message_id"`     // UUID for the message
	SenderID    gocql.UUID `json:"sender_id" db:"sender_id"`       // UUID of the sender (refers to Users.ID)
	MessageText string     `json:"message_text" db:"message_text"` // The actual message content
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`     // Timestamp of when the message was sent
}

// Metadata for gocqlx
var MessagesMetadata = table.Metadata{
	Name: TABLE_NAME,
	Columns: []string{
		COLUMN_MESSAGE_ID,
		COLUMN_SENDER_ID,
		COLUMN_MESSAGE_TEXT,
		COLUMN_CREATED_AT,
	},
	PartKey: []string{COLUMN_SENDER_ID},  // Partition key(s)
	SortKey: []string{COLUMN_CREATED_AT}, // Clustering key(s)
}

func (m Messages) ValidateMessageDetails() error {
	if m.SenderID == (gocql.UUID{}) {
		return fmt.Errorf("sender_id can't be empty")
	}
	if m.MessageText == "" {
		return fmt.Errorf("message_text can't be empty")
	}
	return nil
}

func CreateMessagesTableCQL() string {
	return `
    CREATE TABLE IF NOT EXISTS messages (
        message_id UUID,
        sender_id UUID,
        message_text text,
        created_at timestamp,
        PRIMARY KEY ((sender_id), created_at)
    ) WITH CLUSTERING ORDER BY (created_at ASC);
    `
}

// Optionally create indexes on message_id if querying by message_id directly is needed
func CreateMessageIndexesCQL() []string {
	return []string{
		"CREATE INDEX IF NOT EXISTS ON messages (message_id);",
	}
}
