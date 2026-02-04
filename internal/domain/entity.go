package domain

import "encoding/json"

// Payload represents the Faro Web SDK v2.x transport payload.
// Aligned with Grafana Faro: logs, events, measurements, exceptions.
type Payload struct {
	Meta       Meta         `json:"meta"`
	Logs       []LogEntry   `json:"logs"`
	Events     []Event      `json:"events"`
	Measurements []Measurement `json:"measurements"`
	Exceptions []Exception  `json:"exceptions"`
}

type Meta struct {
	App      AppMeta      `json:"app"`
	Browser  BrowserMeta  `json:"browser"`
	View     ViewMeta     `json:"view"`
	Page     PageMeta     `json:"page"`
	Session  SessionMeta  `json:"session"`
	SDK      SDKMeta      `json:"sdk"`
	User     UserMeta     `json:"user"`
	Extra    map[string]interface{} `json:"-"`
}

type AppMeta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

type BrowserMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Mobile  bool   `json:"mobile"`
}

type ViewMeta struct {
	Name string `json:"name"`
}

type PageMeta struct {
	URL string `json:"url"`
}

type SessionMeta struct {
	ID string `json:"id"`
}

type SDKMeta struct {
	Version string `json:"version"`
}

type UserMeta struct {
	Username   string                 `json:"username"`
	Attributes map[string]interface{} `json:"attributes"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

type Event struct {
	Name       string                 `json:"name"`
	Domain     string                 `json:"domain"`
	Attributes map[string]interface{} `json:"attributes"`
	Timestamp  string                 `json:"timestamp"`
}

type Measurement struct {
	Type      string             `json:"type"`
	Values    map[string]float64 `json:"values"`
	Timestamp string             `json:"timestamp"`
}

type Exception struct {
	Type      string     `json:"type"`
	Value     string     `json:"value"`
	Timestamp string     `json:"timestamp"`
	Stacktrace Stacktrace `json:"stacktrace"`
}

type Stacktrace struct {
	Frames []StackFrame `json:"frames"`
}

type StackFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Lineno   int    `json:"lineno"`
	Colno    int    `json:"colno"`
}

// LokiStream is the push format for Loki (Faro collector → Loki).
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// UnmarshalJSON custom unmarshal to capture unknown meta fields (Faro SDK extensibility).
func (p *Payload) UnmarshalJSON(data []byte) error {
	type alias Payload
	aux := &struct {
		*alias
		Meta json.RawMessage `json:"meta"`
	}{
		alias: (*alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var metaMap map[string]interface{}
	if err := json.Unmarshal(aux.Meta, &metaMap); err != nil {
		return err
	}
	p.Meta.Extra = make(map[string]interface{})
	for k, v := range metaMap {
		fieldBytes, _ := json.Marshal(v)
		switch k {
		case "app":
			_ = json.Unmarshal(fieldBytes, &p.Meta.App)
		case "browser":
			_ = json.Unmarshal(fieldBytes, &p.Meta.Browser)
		case "view":
			_ = json.Unmarshal(fieldBytes, &p.Meta.View)
		case "page":
			_ = json.Unmarshal(fieldBytes, &p.Meta.Page)
		case "session":
			_ = json.Unmarshal(fieldBytes, &p.Meta.Session)
		case "sdk":
			_ = json.Unmarshal(fieldBytes, &p.Meta.SDK)
		case "user":
			_ = json.Unmarshal(fieldBytes, &p.Meta.User)
		default:
			p.Meta.Extra[k] = v
		}
	}
	return nil
}
