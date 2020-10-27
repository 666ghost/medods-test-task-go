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

func (h *Handler) Create(c echo.Context) error {
	log.Print("sdksdklfjsdlkfjsdf")
	u := &models.User{
		ID:   primitive.NewObjectID(),
		Guid: uuid.New().String(),
	}
	_, err := u.Insert(context.TODO())
	if err != nil {
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, map[string]string{
		"guid": u.Guid,
	})
}

type tokenReqBody struct {
	RefreshToken   string `json:"refresh_token"`
	RefreshTokenId string `json:"refresh_token_id"`
}

func (h Handler) Refresh(c echo.Context) error {
	log.Print("1")

	tokenReq := tokenReqBody{}
	err := c.Bind(&tokenReq)

	if err != nil {
		panic(err)
	}
	log.Print("2")
	rToken, err := jwt.Parse(tokenReq.RefreshToken, func(rToken *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		log.Print("3")
		if _, ok := rToken.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Print("err")
			return nil, fmt.Errorf("Unexpected signing method: %v", rToken.Header["alg"])
		}

		return []byte(config.New().TokenSecret), nil
	})
	log.Print("4")

	if err != nil || rToken == nil {
		panic(err)
	}

	aToken := c.Get("user").(*jwt.Token)
	aClaims, ok1 := aToken.Claims.(jwt.MapClaims)
	log.Print("5")
	if ok1 && rToken.Valid {
		log.Print("6")
		u, err := models.SelectUserByGuid(context.TODO(), aClaims["guid"].(string))
		if err != nil {
			panic(err)
		}
		log.Print("7")
		token, err := models.SelectUnusedTokenById(context.TODO(), aClaims["refresh_id"].(string))
		if err != nil || token == nil {
			panic(err)
		}
		log.Print("8")
		if bcrypt.CompareHashAndPassword([]byte(token.Refresh), []byte(tokenReq.RefreshToken)) == nil {
			tokenPair := getNewTokenPair(u)
			log.Print("9")
			_, err = token.Update(context.TODO(), bson.D{{"used", true}})
			if err != nil {
				panic(err)
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
		panic(err)
	}
	u, err := models.SelectUserByGuid(context.TODO(), m["guid"])
	if err == mongo.ErrNoDocuments || u == nil {
		return echo.ErrUnauthorized
	}
	_, err = models.InvalidateOldUserRefreshTokens(context.TODO(), u)
	if err != nil {
		panic(err)
	}

	tokenPair := getNewTokenPair(u)

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
		panic(err)
	}
	aToken := c.Get("user").(*jwt.Token)
	aClaims, _ := aToken.Claims.(jwt.MapClaims)

	u, err := models.SelectUserByGuid(context.TODO(), aClaims["guid"].(string))
	if err != nil {
		panic(err)
	}
	tokenId, err := primitive.ObjectIDFromHex(m["refresh_token_id"])
	if err != nil {
		panic(err)
	}

	_, err = models.RemoveToken(context.TODO(), bson.M{"_id": tokenId, "user_id": u.ID})

	if err == mongo.ErrNoDocuments {
		return echo.ErrNotFound
	} else if err != nil {
		panic(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Token deleted successfully",
	})
}
func (h *Handler) TruncateUserTokens(c echo.Context) error {
	m := make(map[string]string)
	err := c.Bind(&m)
	if err != nil {
		panic(err)
	}
	uIdObj, _ := primitive.ObjectIDFromHex(m["user_id"])
	_, err = models.RemoveToken(context.TODO(), bson.M{"user_id": uIdObj})

	if err == mongo.ErrNoDocuments {
		return echo.ErrNotFound
	} else if err != nil {
		panic(err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User tokens deleted successfully",
	})
}

func getNewTokenPair(u *models.User) map[string]string {
	cfg := config.New()

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["guid"] = u.Guid
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	rt, err := refreshToken.SignedString([]byte(cfg.TokenSecret))
	if err != nil {
		panic(err)
	}
	hashedRToken, err := bcrypt.GenerateFromPassword([]byte(rt), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	token := models.Token{
		ID:      primitive.NewObjectID(),
		UserId:  u.ID,
		Refresh: string(hashedRToken),
		Used:    false,
	}
	savedToken, err := token.Insert(context.TODO())
	if err != nil || savedToken == nil {
		panic(err)
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
		panic(err)
	}

	return map[string]string{"access_token": t, "refresh_token": rt, "refresh_id": refreshId}
}
