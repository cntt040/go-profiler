package profiler

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/alexcesaro/statsd.v2"
	"os"
	"strings"
)

type (
	ProfilerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		Address string
		Service string
	}
)

var (
	// DefaultBodyLimitConfig is the default Gzip middleware config.
	DefaultProfilerConfig = ProfilerConfig{
		Skipper: defaultSkipper,
		Address: ":8125",
		Service: "default",
	}
	client statsd.Client
)

func defaultSkipper(c echo.Context) bool {
	return false
}

func Profiler() echo.MiddlewareFunc {
	return ProfilerWithConfig(DefaultProfilerConfig)
}

func ProfilerWithConfig(config ProfilerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultProfilerConfig.Skipper
	}
	if config.Address == "" {
		config.Address = DefaultProfilerConfig.Address
	}
	if config.Service == "" {
		config.Service = DefaultProfilerConfig.Service
	}

	client, err := statsd.New(statsd.Address(config.Address))
	if err != nil {
		fmt.Println("Failed to initialized Statsd Client %s", err)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			t := client.NewTiming()
			if err = next(c); err != nil {
				c.Error(err)
			}
			s := strings.ToLower(fmt.Sprintf("response.%s.%s.%s.%d", config.Service, req.Method(), req.URL().Path(), res.Status()))
			if os.Getenv("LOG_LEVEL") == "debug" {
				fmt.Println(s)
			}
			t.Send(s)

			return
		}
	}
}
