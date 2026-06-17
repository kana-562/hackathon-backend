package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			elapsed := time.Since(start)
			req := c.Request()
			res := c.Response()
			log.Infof("[%s] %s %s %d %s",
				req.Method, req.URL.Path, req.RemoteAddr, res.Status, elapsed,
			)
			return err
		}
	}
}
