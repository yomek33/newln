package handler

import (
	"fmt"
	"net/http"
	"newln/internal/services"

	"newln/internal/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Handlers struct {
	UserHandler *UserHandler
}

func NewHandler(services *services.Services) *Handlers {
	return &Handlers{
		UserHandler: NewUserHandler(services.UserService),
	}
}
func (h *Handlers) SetDefault(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to newln")
	})
}

func (h *Handlers) SetAPIRoutes(e *echo.Echo) {
	api := e.Group("/api")
	api.POST("/register", h.UserHandler.RegisterUser)
	api.POST("/login", h.UserHandler.LoginUser)
}

func Echo() *echo.Echo {
	e := echo.New()

	// Set up middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Secure())

	// Custom HTTP error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	return e
}
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := echo.Map{"message": "Internal Server Error"}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if message, ok = he.Message.(echo.Map); !ok {
			// he.Message was not of type echo.Map
			if messageStr, ok := he.Message.(string); ok {
				message = echo.Map{"message": messageStr}
			} else {
				message = echo.Map{"message": http.StatusText(code)}
			}
		}
		if he.Internal != nil {
			message = echo.Map{"message": fmt.Sprintf("%v, %v", message, he.Internal)}
		}
		log.Info("HTTP Error: ", code, message)
	}

	// Log the error
	c.Logger().Error(err)
	logger.Errorf("Error: %v", err)

	// Send JSON response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			c.NoContent(code)
		} else {
			c.JSON(code, message)
		}
	}
}
