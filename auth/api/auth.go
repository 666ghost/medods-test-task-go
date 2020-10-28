package api

import (
	"context"
	"fmt"
	"github.com/666ghost/medods-test-task-go/auth/models"
	"github.com/666ghost/medods-test-task-go/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type Handler struct{}

type tokenReqBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Create(c echo.Context) error {
	log.Print("Creating user")
	u := &models.User{
		ID:   primitive.NewObjectID(),
		Guid: uuid.New().String(),
	}
	_, err := u.Insert(context.TODO())
	if err != nil {
		log.Print(err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]string{
		"guid": u.Guid,
	})
}

func (h Handler) Refresh(c echo.Context) error {
	log.Print("refreshing tokens... ")

	tokenReq := tokenReqBody{}
	err := c.Bind(&tokenReq)

	if err != nil {
		log.Print("Failed parsing request", err)
		return echo.ErrInternalServerError
	}
	rToken, err := jwt.Parse(tokenReq.RefreshToken, func(rToken *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := rToken.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v ", rToken.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", rToken.Header["alg"])
		}

		return []byte(config.New().TokenSecret), nil
	})

	if err != nil || rToken == nil {
		log.Print(err)
		return echo.ErrForbidden
	}

	aToken := c.Get("user").(*jwt.Token)
	aClaims, ok1 := aToken.Claims.(jwt.MapClaims)

	if ok1 && rToken.Valid {
		u, err := models.SelectUserByGuid(context.TODO(), aClaims["guid"].(string))
		if err != nil {
			log.Print("Failed to select user by guid ", err)
			return echo.ErrNotFound
		}
		token, err := models.SelectUnusedTokenById(context.TODO(), aClaims["refresh_id"].(string))
		if err != nil || token == nil {
			log.Print("Failed getting unused token by id ", err)
			return echo.ErrForbidden
		}
		if bcrypt.CompareHashAndPassword([]byte(token.Refresh), []byte(tokenReq.RefreshToken)) == nil {
			tokenPair, err := getNewTokenPair(u)
			if err != nil {
				log.Print("Failed generating token pair ", err)
				return echo.ErrInternalServerError
			}
			_, err = token.Update(context.TODO(), bson.D{{"used", true}})
			if err != nil {
				log.Print("Failed setting token used flag to true ", err)
				return echo.ErrInternalServerError
			}

			return c.JSON(http.StatusOK, map[string]string{
				"access_token":  tokenPair["access_token"],
				"refresh_token": tokenPair["refresh_token"],
				"refresh_id":    tokenPair["refresh_id"],
			})
		}
	}
	return echo.ErrUnauthorized
}

func (h *Handler) Login(c echo.Context) error {
	m := make(map[string]string)
	err := c.Bind(&m)
	if err != nil {
		log.Print("failed parsing request ", err)
		return echo.ErrInternalServerError
	}
	u, err := models.SelectUserByGuid(context.TODO(), m["guid"])
	if err == mongo.ErrNoDocuments || u == nil {
		return echo.ErrUnauthorized
	}
	_, err = models.InvalidateOldUserRefreshTokens(context.TODO(), u)
	if err != nil {
		log.Print("failed invalidate old user tokens ", err)
	}

	tokenPair, err := getNewTokenPair(u)
	if err != nil {
		log.Print("Failed generating token pair ", err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  tokenPair["access_token"],
		"refresh_token": tokenPair["refresh_token"],
		"refresh_id":    tokenPair["refresh_id"],
	})
}

func (h *Handler) RemoveToken(c echo.Context) error {
	m := make(map[string]string)
	err := c.Bind(&m)

	if err != nil {
		log.Print("Failed parsing request ", err)
		return echo.ErrInternalServerError
	}
	aToken := c.Get("user").(*jwt.Token)
	aClaims, _ := aToken.Claims.(jwt.MapClaims)

	u, err := models.SelectUserByGuid(context.TODO(), aClaims["guid"].(string))
	if err != nil {
		log.Print("Failed parsing request ", err)
		return echo.ErrInternalServerError
	}
	tokenId, err := primitive.ObjectIDFromHex(m["refresh_id"])
	if err != nil {
		log.Print("Failed parsing refresh token id ", err)
		return echo.ErrBadRequest
	}

	_, err = models.RemoveToken(context.TODO(), bson.M{"_id": tokenId, "user_id": u.ID})

	if err == mongo.ErrNoDocuments {
		return echo.ErrNotFound
	} else if err != nil {
		log.Print("Failed removing tokens ", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Token deleted successfully",
	})
}
func (h *Handler) TruncateUserTokens(c echo.Context) error {
	m := make(map[string]string)
	err := c.Bind(&m)
	if err != nil {
		log.Print("failed parsing request ", err)
		return echo.ErrInternalServerError
	}
	u, err := models.SelectUserByGuid(context.TODO(), m["guid"])
	if err != nil {
		log.Print("Failed to select user by guid ", err)
		return echo.ErrNotFound
	}

	_, err = models.RemoveToken(context.TODO(), bson.M{"user_id": u.ID})

	if err == mongo.ErrNoDocuments {
		return echo.ErrNotFound
	} else if err != nil {
		log.Print("Failed removing tokens ", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User tokens deleted successfully",
	})
}

func getNewTokenPair(u *models.User) (map[string]string, error) {
	cfg := config.New()

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["guid"] = u.Guid
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	rt, err := refreshToken.SignedString([]byte(cfg.TokenSecret))
	if err != nil {
		return nil, err
	}
	hashedRToken, err := bcrypt.GenerateFromPassword([]byte(rt), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	token := models.Token{
		ID:      primitive.NewObjectID(),
		UserId:  u.ID,
		Refresh: string(hashedRToken),
		Used:    false,
	}
	savedToken, err := token.Insert(context.TODO())
	if err != nil || savedToken == nil {
		return nil, err
	}
	refreshId := savedToken.InsertedID.(primitive.ObjectID).Hex()

	// Create token
	accessToken := jwt.New(jwt.SigningMethodHS512)
	claims := accessToken.Claims.(jwt.MapClaims)
	claims["refresh_id"] = refreshId
	claims["guid"] = u.Guid
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix() // Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID          works too)
	t, err := accessToken.SignedString([]byte(cfg.TokenSecret))
	if err != nil {
		return nil, err
	}

	return map[string]string{"access_token": t, "refresh_token": rt, "refresh_id": refreshId}, nil
}
