package serve

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

type WebServer struct {
	app    *echo.Echo
	enc    *encoder.Encoder
	logger *logrus.Logger
}

type EncodeRequest struct {
	Protocol uint64 `json:"protocol" validate:"required"`
	Crc32    uint64 `json:"crc32" validate:"required"`
}

type EncodeResponse struct {
	Result uint64 `json:"result"`
}

func setupWeb(enc *encoder.Encoder, logger *logrus.Logger) *WebServer {
	app := echo.New()
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
	handler := encodeHandler(enc, logger)
	app.POST("/", handler)
	return &WebServer{
		app:    app,
		enc:    enc,
		logger: logger,
	}
}

func encodeHandler(enc *encoder.Encoder, logger *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req EncodeRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request").SetInternal(err)
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
