package oauth

import (
	"encoding/json"
	"net/http"
	"time"

	"cme/internal/app/api/middleware/jwt"
	"cme/internal/app/constants"
	"cme/internal/app/controller"
	users_DBModels "cme/internal/app/db/dto/users"
	userDB "cme/internal/app/db/repository/user"

	"cme/internal/app/service/correlation"
	"cme/internal/app/service/logger"
	"cme/internal/app/service/util"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// IOAuthController represents the interface for OAuthController
type IOAuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

// OAuthController is the implementation of the IOAuthController interface
type OAuthController struct {
	UserDBClient userDB.IUserRepository
	JWT          jwt.IJwtService
}

// NewOAuthController creates a new instance of OAuthController
func NewOAuthController(
	UserClient userDB.IUserRepository,
	jwt jwt.IJwtService,
) IOAuthController {
	return &OAuthController{
		UserDBClient: UserClient,
		JWT:          jwt,
	}
}

// Register handles the sign-up functionality
func (u OAuthController) Register(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	dataFromBody := users_DBModels.Users{}                       // Create an empty Users struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Users struct
	if err != nil {
		log.Errorf(constants.BadRequest, err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.BadRequest)
		return
	}

	if err := dataFromBody.ValidateRegisterDetails(); err != nil {
		// Validate the sign-up details
		controller.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := util.GenerateHash(dataFromBody.Password) // Generate a hashed password
	if err != nil {
		log.Errorf(constants.InternalServerError, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.InternalServerError)
		return
	}

	dataFromBody.ID = gocql.TimeUUID()
	dataFromBody.Password = hashedPassword // Set the hashed password in the Users struct
	dataFromBody.Status = true             // Set the status to true
	dataFromBody.CreatedAt = time.Now()    // Set the current time as the creation time

	if err = u.UserDBClient.CreateUser(ctx, dataFromBody); err != nil {
		// Create a new user using the UserDBClient
		log.Errorf(constants.InternalServerError, err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.InternalServerError)
		return
	}

	// Respond with success message and nil data
	controller.RespondWithSuccess(c, http.StatusAccepted, "User Created Successfully", nil)
}

// Login handles the login functionality
func (u OAuthController) Login(c *gin.Context) {
	ctx := correlation.WithReqContext(c) // Get the request context
	log := logger.Logger(ctx)            // Get the logger

	dataFromBody := users_DBModels.Users{}                       // Create an empty Users struct
	err := json.NewDecoder(c.Request.Body).Decode(&dataFromBody) // Decode the request body into the Users struct
	if err != nil {
		log.Errorf(constants.BadRequest, err)
		controller.RespondWithError(c, http.StatusBadRequest, constants.BadRequest)
		return
	}

	whr := make(map[string]interface{})
	whr[users_DBModels.COLUMN_EMAIL] = dataFromBody.Email
	whr[users_DBModels.COLUMN_STATUS] = true

	user, err := u.UserDBClient.GetUser(ctx, whr)
	if err != nil {
		log.Errorf("Error getting user", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.InternalServerError)
		return
	}

	if !util.ValidatePassword(dataFromBody.Password, user.Password) {
		// Validate the provided password with the stored password
		log.Errorf("wrong credentials")
		controller.RespondWithError(c, http.StatusUnauthorized, "invalid email/password")
		return
	}

	token, err := u.JWT.CreateNewTokens(ctx, user) // Create new access tokens for the user
	if err != nil {
		log.Errorf("error while creating access token", err)
		controller.RespondWithError(c, http.StatusInternalServerError, constants.InternalServerError)
		return
	}

	// Respond with success message and the generated token
	controller.RespondWithSuccess(c, http.StatusAccepted, "Login Successfully", token)
}
