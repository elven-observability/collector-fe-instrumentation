package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"collector-fe-instrumentation/internal/domain"
)

// LokiWriter sends log streams to Loki (interface for clean architecture).
type LokiWriter interface {
	Push(ctx context.Context, tenantID string, streams []domain.LokiStream) error
}

// CollectorService implements the collect-logs use case (Faro → Loki).
type CollectorService struct {
	loki LokiWriter
	log  *slog.Logger
}

func NewCollectorService(loki LokiWriter, log *slog.Logger) *CollectorService {
	if log == nil {
		log = slog.Default()
	}
	return &CollectorService{loki: loki, log: log}
}

// Collect validates the payload, converts it to Loki streams, and pushes to Loki.
func (s *CollectorService) Collect(ctx context.Context, tenantID string, payload *domain.Payload) error {
	if payload == nil {
		return domain.ErrInvalidPayload
	}
	if tenantID == "" {
		return domain.ErrMissingTenant
	}

	streams := s.payloadToStreams(payload)
	if len(streams) == 0 {
		return domain.ErrEmptyPayload
	}

	if err := s.loki.Push(ctx, tenantID, streams); err != nil {
		s.log.ErrorContext(ctx, "loki push failed", "error", err, "tenant", tenantID)
		return fmt.Errorf("%w: %v", domain.ErrLokiSend, err)
	}
	return nil
}

func (s *CollectorService) payloadToStreams(p *domain.Payload) []domain.LokiStream {
	base := s.baseFields(p)
	var streams []domain.LokiStream

	for _, e := range p.Logs {
		fields := copyMap(base)
		fields["kind"] = logKind(e.Level)
		fields["level"] = e.Level
		fields["message"] = e.Message
		streams = append(streams, toLokiStream(fields))
	}
	for _, e := range p.Events {
		fields := copyMap(base)
		fields["kind"] = "event"
		fields["event_name"] = e.Name
		fields["event_domain"] = e.Domain
		fields["event_timestamp"] = e.Timestamp
		for k, v := range e.Attributes {
			fields[fmt.Sprintf("event_data_%s", k)] = v
		}
		streams = append(streams, toLokiStream(fields))
	}
	for _, m := range p.Measurements {
		fields := copyMap(base)
		fields["kind"] = "measurement"
		fields["measurement_type"] = m.Type
		fields["measurement_timestamp"] = m.Timestamp
		for k, v := range m.Values {
			fields[fmt.Sprintf("measurement_value_%s", k)] = v
		}
		streams = append(streams, toLokiStream(fields))
	}
	for _, ex := range p.Exceptions {
		fields := copyMap(base)
		fields["kind"] = "exception"
		fields["exception_type"] = ex.Type
		fields["exception_value"] = ex.Value
		fields["exception_timestamp"] = ex.Timestamp
		var stack []string
		for _, f := range ex.Stacktrace.Frames {
			stack = append(stack, fmt.Sprintf("%s:%s:%d:%d", f.Filename, f.Function, f.Lineno, f.Colno))
		}
		fields["exception_stacktrace"] = joinStrings(stack, " | ")
		streams = append(streams, toLokiStream(fields))
	}

	return streams
}

func (s *CollectorService) baseFields(p *domain.Payload) map[string]interface{} {
	fields := map[string]interface{}{
		"app":             p.Meta.App.Name,
		"app_version":     p.Meta.App.Version,
		"environment":     p.Meta.App.Environment,
		"browser_name":    p.Meta.Browser.Name,
		"browser_version": p.Meta.Browser.Version,
		"browser_os":      p.Meta.Browser.OS,
		"browser_mobile":  p.Meta.Browser.Mobile,
		"session_id":      p.Meta.Session.ID,
		"page_url":        p.Meta.Page.URL,
		"view_name":       p.Meta.View.Name,
		"sdk_version":     p.Meta.SDK.Version,
		"user_username":   p.Meta.User.Username,
	}
	for k, v := range p.Meta.User.Attributes {
		fields[fmt.Sprintf("user_attr_%s", k)] = v
	}
	for k, v := range p.Meta.Extra {
		if k == "user" {
			continue
		}
		switch val := v.(type) {
		case map[string]interface{}:
			for sk, sv := range val {
				fields[fmt.Sprintf("%s_%s", k, sk)] = sv
			}
		case []interface{}:
			fields[k] = val
		default:
			fields[k] = v
		}
	}
	return fields
}

func logKind(level string) string {
	switch level {
	case "error":
		return "error"
	case "warning":
		return "warning"
	case "debug":
		return "debug"
	default:
		return "info"
	}
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for i := 1; i < len(ss); i++ {
		out += sep + ss[i]
	}
	return out
}

func toLokiStream(fields map[string]interface{}) domain.LokiStream {
	labels := map[string]string{
		"app":         fmt.Sprint(fields["app"]),
		"kind":        fmt.Sprint(fields["kind"]),
		"level":       fmt.Sprint(fields["level"]),
		"environment": fmt.Sprint(fields["environment"]),
		"browser":     fmt.Sprint(fields["browser_name"]),
		"session_id":  fmt.Sprint(fields["session_id"]),
	}
	line := formatLogLine(fields)
	ts := fmt.Sprintf("%d", time.Now().UnixNano())
	return domain.LokiStream{
		Stream: labels,
		Values: [][]string{{ts, line}},
	}
}

func formatLogLine(fields map[string]interface{}) string {
	var s string
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			s += fmt.Sprintf("%s=%q ", k, val)
		case bool, float64, int, int64:
			s += fmt.Sprintf("%s=%v ", k, val)
		default:
			s += fmt.Sprintf("%s=%v ", k, v)
		}
	}
	return s
}
