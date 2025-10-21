package serve

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

type WebServer struct {
	app    *echo.Echo
	logger *logrus.Logger
}

type EncodeRequest struct {
	Server   string `json:"server" validate:"required"`
	Protocol uint64 `json:"protocol" validate:"required"`
	Crc32    uint64 `json:"crc32" validate:"required"`
}

type EncodeResponse struct {
	Result uint64 `json:"result"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func newValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func setupWeb(encoders map[string]*encoder.Encoder, logger *logrus.Logger) *WebServer {
	app := echo.New()
	app.Validator = newValidator()
	// Middleware
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.WithFields(logrus.Fields{
					"uri":    v.URI,
					"status": v.Status,
				}).Info("request")
			} else {
				logger.WithFields(logrus.Fields{
					"uri":    v.URI,
					"status": v.Status,
					"err":    v.Error.Error(),
				}).Error("request error")
			}
			return nil
		},
	}))
	app.Use(middleware.Recover())
	app.Use(echoprometheus.NewMiddleware("encoder"))
	app.Pre(middleware.RemoveTrailingSlash())

	// Routes
	handler := encodeHandler(encoders, logger)
	app.POST("/", handler)
	return &WebServer{
		app:    app,
		logger: logger,
	}
}

func encodeHandler(encoders map[string]*encoder.Encoder, logger *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req EncodeRequest
		if err := c.Bind(&req); err != nil {
			logger.WithError(err).Warn("failed to bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request").SetInternal(err)
		}
		if err := c.Validate(&req); err != nil {
			logger.WithError(err).Warn("validation error")
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request").SetInternal(err)
		}

		enc, ok := encoders[req.Server]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "server not found")
		}
		result, err := enc.Encode(req.Protocol, req.Crc32)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to encode").SetInternal(err)
		}
		return c.JSON(http.StatusOK, EncodeResponse{Result: result.ReturnValue})
	}
}

func (s *WebServer) Serve(port string) error {
	return s.app.Start(port)
}

func (s *WebServer) Shutdown(timeout time.Duration) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.app.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Server forced to shutdown")
		return err
	}
	s.logger.Info("Server exited gracefully")
	return nil
}
