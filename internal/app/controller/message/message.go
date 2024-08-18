package message

import (
	"encoding/json"
	"net/http"
	"time"

	"cme/internal/app/constants"
	"cme/internal/app/controller"
	messages_DBModels "cme/internal/app/db/dto/message"
	users_DBModels "cme/internal/app/db/dto/users"
	messageDB "cme/internal/app/db/repository/message"
	"cme/internal/app/service/correlation"
	"cme/internal/app/service/logger"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// IMessageController represents the interface for MessageController
type IMessageController interface {
	SendMessage(c *gin.Context)
	GetMessages(c *gin.Context)
}

// MessageController is the implementation of the IMessageController interface
type MessageController struct {
	MessageDBClient messageDB.IMessageRepository
}

// NewMessageController creates a new instance of MessageController
func NewMessageController(
	MessageClient messageDB.IMessageRepository) IMessageController {
	return &MessageController{
		MessageDBClient: MessageClient,
	}
}

// SendMessage handles sending messages between users
func (m MessageController) SendMessage(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	dataFromBody := messages_DBModels.Messages{}                 // Create an empty Messages struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Messages struct
	if err != nil {
		log.Errorf("Invalid request body", err)
		controller.RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	context, exist := c.Get(constants.CTK_CLAIM_KEY.String()) // Retrieve the user context from the request
	if !exist {
		log.Errorf("context not found")
		controller.RespondWithError(c, http.StatusUnauthorized, "Context Not Found")
		return
	}
	usr := context.(*users_DBModels.Users) // Type assertion to retrieve the user information

	dataFromBody.SenderID = usr.ID
	dataFromBody.MessageID = gocql.TimeUUID() // Generate a new UUID for the message
	dataFromBody.CreatedAt = time.Now()       // Set the current time as the creation time

	if err := dataFromBody.ValidateMessageDetails(); err != nil {
		// Validate the message details
		controller.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err = m.MessageDBClient.SendMessage(ctx, dataFromBody); err != nil {
		// Send the message using the MessageDBClient
		log.Errorf("Error sending message", err)
		controller.RespondWithError(c, http.StatusInternalServerError, "Error sending message")
		return
	}

	// Respond with success message and nil data
	controller.RespondWithSuccess(c, http.StatusAccepted, "Message Sent Successfully", nil)
}

// GetMessages handles retrieving messages between users
func (m MessageController) GetMessages(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	whr := make(map[string]interface{})

	// Retrieve sender_id and recipient_id from query parameters
	senderID := c.Query("sender_id")
	if senderID != "" {
		// Prepare the query filter
		whr[messages_DBModels.COLUMN_SENDER_ID] = senderID
	}

	messages, err := m.MessageDBClient.GetMessages(ctx, whr)
	if err != nil {
		log.Errorf("Error getting messages", err)
		controller.RespondWithError(c, http.StatusInternalServerError, "Error retrieving messages")
		return
	}

	// Respond with the list of messages
	controller.RespondWithSuccess(c, http.StatusOK, "Messages retrieved successfully", messages)
}
