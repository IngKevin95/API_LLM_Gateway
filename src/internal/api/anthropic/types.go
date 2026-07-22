package anthropic

import "encoding/json"

type MessageRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Stream    bool      `json:"stream,omitempty"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	var aux struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}
	aux.Alias = (*Alias)(m)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var str string
	if err := json.Unmarshal(aux.Content, &str); err == nil {
		m.Content = []Content{{Type: "text", Text: str}}
		return nil
	}
	var arr []Content
	if err := json.Unmarshal(aux.Content, &arr); err == nil {
		m.Content = arr
		return nil
	}
	return nil
}

type Content struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Id    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type MessageResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	Content      []Content `json:"content"`
	StopReason   string    `json:"stop_reason,omitempty"`
	StopSequence string    `json:"stop_sequence,omitempty"`
	Usage        Usage     `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Event struct {
	Type  string `json:"type"`
	Event string `json:"-"`
}

type MessageStartEvent struct {
	Type    string          `json:"type"`
	Message MessageResponse `json:"message"`
}

type ContentBlockDeltaEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
