package server

import (
	"cme/internal/app/api/middleware/auth"
	"cme/internal/app/api/middleware/jwt"
	"cme/internal/app/constants"
	"cme/internal/app/controller/healthcheck"
	messageController "cme/internal/app/controller/message"
	"cme/internal/app/controller/oauth"
	"cme/internal/app/db"
	"cme/internal/app/db/cache"
	messageDBClient "cme/internal/app/db/repository/message"
	userDBClient "cme/internal/app/db/repository/user"
	"cme/internal/app/service/logger"
	"context"
	"strings"
	"time"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Init(ctx context.Context, dbConnection *db.DBService) *gin.Engine {
	if strings.EqualFold(constants.Config.Environment, "prod") {
		gin.SetMode(gin.ReleaseMode)
	}
	return NewRouter(ctx, dbConnection)

}
func addCSPHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func addReferrerPolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func addPermissionsPolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Permissions-Policy", "default-src 'none'")
		c.Next()
	}
}

func addFeaturePolicyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Feature-Policy", "none")
		c.Next()
	}
}

func NewRouter(ctx context.Context, dbConnection *db.DBService) *gin.Engine {
	log := logger.Logger(ctx)

	log.Info("setting up service and controllers")

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(helmet.Default())
	//Content-Security-Policy
	router.Use(addCSPHeader())
	//Referrer-Policy
	router.Use(addReferrerPolicyHeader())
	//Permissions-Policy
	router.Use(addPermissionsPolicyHeader())
	//Feature-Policy
	router.Use(addFeaturePolicyHeader())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", constants.AUTHORIZATION, constants.CORRELATION_KEY_ID.String()},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(uuidInjectionMiddleware())

	// Cache
	redisCache := cache.NewRedisCache(constants.Config.REDIS)

	// DB Clients
	var (
		userDBClient    = userDBClient.NewUserRepository(dbConnection)
		messageDBClient = messageDBClient.NewMessageRepository(dbConnection, redisCache)
	)

	// SERVICES
	var (
		jwt = jwt.NewJwtService(userDBClient)
	)

	// Controller
	var (
		messageController = messageController.NewMessageController(messageDBClient)

		oAuthController       = oauth.NewOAuthController(userDBClient, jwt)
		healthCheckController = healthcheck.NewHealthCheckController()
	)

	v1 := router.Group("/cme/v1")
	{
		v1.GET(HEALTH_CHECK, healthCheckController.HealthCheck)

		v1.POST(REGISTER, oAuthController.Register)
		v1.POST(LOGIN, oAuthController.Login)

		authenticated := v1.Group("")
		{
			authenticated.Use(auth.Authentication(jwt))

			authenticated.POST(SEND, messageController.SendMessage)
			authenticated.GET(MESSAGE, messageController.GetMessages)
		}
	}

	return router
}

// uuidInjectionMiddleware injects the request context with a correlation id of type uuid
func uuidInjectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationId := c.GetHeader(string(constants.CORRELATION_KEY_ID))
		if len(correlationId) == 0 {
			correlationID, _ := uuid.NewUUID()
			correlationId = correlationID.String()
			c.Request.Header.Set(constants.CORRELATION_KEY_ID.String(), correlationId)
		}
		c.Writer.Header().Set(constants.CORRELATION_KEY_ID.String(), correlationId)

		c.Next()
	}
}
