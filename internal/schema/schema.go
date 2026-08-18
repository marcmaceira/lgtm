// Package schema holds the JSON Schemas used to force structured output
// from the LLM backends.
package schema

const Commit = `{
  "type": "object",
  "properties": {
    "subject": {"type": "string"},
    "body": {"type": "string"}
  },
  "required": ["subject", "body"],
  "additionalProperties": false
}`

const PR = `{
  "type": "object",
  "properties": {
    "title": {"type": "string"},
    "body": {"type": "string"}
  },
  "required": ["title", "body"],
  "additionalProperties": false
}`

type CommitResult struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type PRResult struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
