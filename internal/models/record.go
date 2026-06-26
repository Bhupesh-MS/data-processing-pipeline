package models

type Record struct {
	Fields     []string       `json:"fields,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}
