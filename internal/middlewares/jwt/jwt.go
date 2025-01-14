package jwt // Service is an interface from which our api module can access our repository of all our models

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rilgilang/sticker-collection-api/config/dotenv"
	"github.com/rilgilang/sticker-collection-api/internal/api/presenter"
	"github.com/rilgilang/sticker-collection-api/internal/consts"
	"github.com/rilgilang/sticker-collection-api/internal/entities"
	"github.com/rilgilang/sticker-collection-api/internal/pkg/logger"
	"github.com/rilgilang/sticker-collection-api/internal/repositories"
	"strconv"
	"time"
)

type AuthMiddleware interface {
	GenerateToken(user *entities.User) (*string, error)
	ValidateToken() fiber.Handler
}

type authMiddlewares struct {
	userRepo repositories.UserRepository
	cfg      *dotenv.Config
}

type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(userRepo repositories.UserRepository, cfg *dotenv.Config) AuthMiddleware {
	return &authMiddlewares{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (m *authMiddlewares) GenerateToken(user *entities.User) (*string, error) {
	jwtKey := m.cfg.JWTKey
	expireMinute := m.cfg.JWTExpiredMin
	// Declare the expiration time of the token
	expirationTime := time.Now().Add(time.Duration(expireMinute) * time.Minute)
	// Create the JWT claims, which includes the email and expiry time
	claims := &Claims{
		ID:    user.ID,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		return nil, err
	}

	return &tokenString, nil
}

func (m *authMiddlewares) ValidateToken() fiber.Handler {
	return func(c *fiber.Ctx) error {

		var (
			jwtKey        = m.cfg.JWTKey
			authorization = ""
			log           = logger.NewLog("jwt_middleware_generate_token", m.cfg.LoggerEnable)
		)

		if len(authorization) != 2 {
			log.Error("authorization token not valid")
			c.Status(400)
			return c.JSON(presenter.ErrorResponse(errors.New("token not valid!")))
		}

		token := authorization[1]

		// Initialize a new instance of `Claims`
		claims := &Claims{}

		// Parse the JWT string and store the result in `claims`.
		// Note that we are passing the key in this method as well. This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match
		tkn, err := jwt.ParseWithClaims(strconv.Itoa(int(token)), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})

		if err != nil {
			log.Error(fmt.Sprintf(`authorization failed got %s`, err))
			if err == jwt.ErrSignatureInvalid {
				c.Status(401)
				return c.JSON(presenter.ErrorResponse(err))
			}
			c.Status(400)
			return c.JSON(presenter.ErrorResponse(err))
		}
		if !tkn.Valid {
			log.Error("authorization failed token invalid")
			c.Status(401)
			return c.JSON(presenter.ErrorResponse(err))
		}

		c.Locals(consts.UserId, claims.ID)
		return c.Next()
	}
}
