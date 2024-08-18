package jwt

import (
	"cme/internal/app/constants"
	users_DBModels "cme/internal/app/db/dto/users"
	userDB "cme/internal/app/db/repository/user"

	"cme/internal/app/service/logger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type IJwtService interface {
	CreateNewTokens(ctx context.Context, user users_DBModels.Users) (*TokenDetails, error)
	VerifyToken(ctx context.Context, tokenString string) (*users_DBModels.Users, bool)
}

type JwtService struct {
	UserDBClient userDB.IUserRepository
}

func NewJwtService(UserClient userDB.IUserRepository) *JwtService {
	return &JwtService{
		UserDBClient: UserClient,
	}
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func (j *JwtService) CreateNewTokens(ctx context.Context, user users_DBModels.Users) (*TokenDetails, error) {
	log := logger.Logger(ctx)
	log.Infof("Creating token for ", user)

	user.Password = ""
	var err error

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_ACCESS_EXP)).Unix()
	td.AccessUuid = uuid.NewString()

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user"] = user
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET))
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	td.RefreshUuid = td.AccessUuid + "++" + user.Email
	td.RtExpires = time.Now().Add(time.Minute * time.Duration(constants.Config.JwtConfig.JWT_REFRESH_EXP)).Unix()

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user"] = user
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(constants.Config.JwtConfig.JWT_REFRESH_SECRET))
	if err != nil {
		log.Errorf("error while generating access id ", err)
		return nil, err
	}
	return td, nil
}

func (j *JwtService) VerifyToken(ctx context.Context, tokenString string) (*users_DBModels.Users, bool) {
	log := logger.Logger(ctx)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(constants.Config.JwtConfig.JWT_ACCESS_SECRET), nil
	})
	if err != nil {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		user := users_DBModels.Users{}
		jsonString, err := json.Marshal(claims["user"])
		if err != nil {
			log.Errorf("unable to marshal user claims", err.Error())
			return nil, false
		}
		// convert json to struct
		err = json.Unmarshal(jsonString, &user)
		if err != nil {
			return nil, false
		}
		whr := make(map[string]interface{})
		whr[users_DBModels.COLUMN_EMAIL] = user.Email

		u, err := j.UserDBClient.GetUser(ctx, whr)
		if err != nil || !u.Status {
			return nil, false
		}

		return &user, true
	}
	return nil, false
}
