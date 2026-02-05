package http

import (
	"log/slog"
	"strings"
	"time"

	"collector-fe-instrumentation/internal/config"
	"collector-fe-instrumentation/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Router builds the Gin engine with CORS, JWT auth, and collect route.
func Router(cfg *config.Config, collector *usecase.CollectorService, opts ...RouterOption) *gin.Engine {
	o := routerOptions{}
	for _, fn := range opts {
		fn(&o)
	}
	// o.slog may be nil; handler uses slog.Default() then

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware(cfg))
	r.Use(requestHeadersCORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	collectorHandler := NewCollectorHandler(collector, o.slog)
	r.POST("/collect/:tenant/:token", JWTAuth(cfg), collectorHandler.Collect)

	return r
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.AllowOrigins
	// Check if wildcard is present
	allowAll := false
	for _, o := range origins {
		if o == "*" {
			allowAll = true
			break
		}
	}

	corsCfg := cors.Config{
		AllowMethods:     []string{"POST", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Scope-OrgID", "X-Faro-Session-Id", "Origin", "Accept", "Referer", "User-Agent"},
		ExposeHeaders:    []string{"Content-Length", "X-Kong-Request-ID", "X-Kong-Upstream-Latency"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if allowAll {
		corsCfg.AllowAllOrigins = true
	} else {
		allowFunc := func(origin string) bool {
			for _, allowed := range origins {
				if origin == allowed {
					return true
				}
				if strings.HasPrefix(allowed, "https://*.") {
					base := strings.TrimPrefix(allowed, "https://*.")
					if strings.HasSuffix(origin, base) {
						return true
					}
				}
			}
			return false
		}
		corsCfg.AllowOriginFunc = allowFunc
	}
	return cors.New(corsCfg)
}

func requestHeadersCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h := c.Request.Header.Get("Access-Control-Request-Headers"); h != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", h)
		}
		c.Next()
	}
}

// RouterOption configures the router.
type RouterOption func(*routerOptions)

type routerOptions struct {
	slog *slog.Logger
}

// WithLogger sets the logger for the collect handler.
func WithLogger(l *slog.Logger) RouterOption {
	return func(o *routerOptions) {
		o.slog = l
	}
}
