package message

import (
	"context"
	"encoding/json"
	"errors"

	"cme/internal/app/db"
	"cme/internal/app/db/cache"
	messages_DBModels "cme/internal/app/db/dto/message"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
)

type IMessageRepository interface {
	SendMessage(ctx context.Context, message messages_DBModels.Messages) error
	GetMessages(ctx context.Context, whr map[string]interface{}) ([]messages_DBModels.Messages, error)
	GetAllMessages(ctx context.Context) ([]messages_DBModels.Messages, error)
}

type MessageRepository struct {
	DBService  *db.DBService
	RedisCache *cache.RedisCache
	table      table.Table
}

func NewMessageRepository(dbService *db.DBService, redisCache *cache.RedisCache) IMessageRepository {
	messageTable := table.New(messages_DBModels.MessagesMetadata) // Ensure MessagesMetadata is correctly defined
	return &MessageRepository{
		DBService:  dbService,
		RedisCache: redisCache,
		table:      *messageTable,
	}
}

// Cache recent messages in Redis
func (m *MessageRepository) cacheMessages(key string, messages []messages_DBModels.Messages) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	return m.RedisCache.Set(key, data)
}

// Retrieve messages from cache
func (m *MessageRepository) getMessagesFromCache(key string) ([]messages_DBModels.Messages, error) {
	data, err := m.RedisCache.Get(key)
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	var messages []messages_DBModels.Messages
	err = json.Unmarshal([]byte(data), &messages)
	return messages, err
}

func (m *MessageRepository) SendMessage(ctx context.Context, message messages_DBModels.Messages) error {
	stmt, names := m.table.Insert()

	q := gocqlx.Query(m.DBService.GetSession().Query(stmt), names).BindStruct(message)
	if err := q.ExecRelease(); err != nil {
		return err
	}

	// Invalidate cache
	m.RedisCache.Delete("messages:" + message.SenderID.String())
	m.RedisCache.Delete("messages:" + message.MessageID.String())

	return nil
}

func (m *MessageRepository) GetMessages(ctx context.Context, whr map[string]interface{}) ([]messages_DBModels.Messages, error) {
	var messages []messages_DBModels.Messages

	var cacheKey string
	if _, ok := whr["sender_id"]; ok {
		cacheKey = "messages" + whr["sender_id"].(string)
	} else {
		cacheKey = "messages"
	}

	// Generate cache key based on the sender and recipient IDs

	// Try to get messages from cache
	cachedMessages, err := m.getMessagesFromCache(cacheKey)
	if err == nil && cachedMessages != nil {
		return cachedMessages, nil
	}

	// Build the CQL statement with WHERE conditions
	stmt, names := qb.Select(m.table.Name()).Where(whrToEq(whr)...).ToCql()

	// Append ALLOW FILTERING to the query
	stmt += " ALLOW FILTERING"

	// Prepare and execute the query
	q := gocqlx.Query(m.DBService.GetSession().Query(stmt), names).BindMap(whr)

	// Fetch the results
	if err := q.SelectRelease(&messages); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return messages, nil // Return an empty slice if no messages are found
		}
		return nil, err
	}

	// Cache the result
	m.cacheMessages(cacheKey, messages)

	return messages, nil
}

func (m *MessageRepository) GetAllMessages(ctx context.Context) ([]messages_DBModels.Messages, error) {
	var messages []messages_DBModels.Messages

	// Build the CQL statement to select all messages
	stmt, names := qb.Select(m.table.Name()).ToCql()

	// Prepare and execute the query
	q := gocqlx.Query(m.DBService.GetSession().Query(stmt), names)

	// Fetch all results
	if err := q.SelectRelease(&messages); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return messages, nil // Return an empty slice if no messages are found
		}
		return nil, err
	}

	return messages, nil
}
func whrToEq(whr map[string]interface{}) []qb.Cmp {
	var cmps []qb.Cmp
	for key := range whr {
		cmps = append(cmps, qb.Eq(key))
	}
	return cmps
}
