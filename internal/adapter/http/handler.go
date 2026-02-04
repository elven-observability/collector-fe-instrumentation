package http

import (
	"log/slog"
	"net/http"
	"regexp"

	"collector-fe-instrumentation/internal/domain"
	"collector-fe-instrumentation/internal/usecase"

	"github.com/gin-gonic/gin"
)

var tenantTokenRe = regexp.MustCompile(`^[\w\s\-.@]+$`)

// CollectorHandler handles Faro collect endpoint.
type CollectorHandler struct {
	svc *usecase.CollectorService
	log *slog.Logger
}

// NewCollectorHandler creates the HTTP handler for /collect/:tenant/:token.
func NewCollectorHandler(svc *usecase.CollectorService, log *slog.Logger) *CollectorHandler {
	if log == nil {
		log = slog.Default()
	}
	return &CollectorHandler{svc: svc, log: log}
}

// Collect is the POST /collect/:tenant/:token handler.
func (h *CollectorHandler) Collect(c *gin.Context) {
	tenantID := sanitizeParam(c.Param("tenant"))
	token := sanitizeParam(c.Param("token"))
	if tenantID == "" || token == "" {
		h.log.Warn("collect: missing tenant or token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant or token not provided"})
		return
	}

	var payload domain.Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.log.Warn("collect: invalid JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := h.svc.Collect(c.Request.Context(), tenantID, &payload); err != nil {
		switch {
		case err == domain.ErrInvalidPayload || err == domain.ErrMissingTenant:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		case err == domain.ErrEmptyPayload:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload, no data found"})
			return
		default:
			h.log.Error("collect: loki push failed", "error", err, "tenant", tenantID)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failure"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func sanitizeParam(s string) string {
	s = trimSpace(s)
	if s == "" || !tenantTokenRe.MatchString(s) {
		return ""
	}
	return s
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
