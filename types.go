package xiangxinai

// Message Message model
type Message struct {
	Role    string      `json:"role"`    // Message role: user, assistant
	Content interface{} `json:"content"` // Message content, can be string or []interface{} (multimodal)
}

// NewMessage Create new message
func NewMessage(role, content string) *Message {
	return &Message{
		Role:    role,
		Content: content,
	}
}

// GuardrailRequest Guardrail detection request model
type GuardrailRequest struct {
	Model    string     `json:"model"`    // Model name
	Messages []*Message `json:"messages"` // Message list
}

// ComplianceResult Compliance detection result
type ComplianceResult struct {
	RiskLevel  string   `json:"risk_level"` // Risk level: no_risk, low_risk, medium_risk, high_risk
	Categories []string `json:"categories"` // Risk category list
}

// SecurityResult Security detection result
type SecurityResult struct {
	RiskLevel  string   `json:"risk_level"` // Risk level: no_risk, low_risk, medium_risk, high_risk
	Categories []string `json:"categories"` // Risk category list
}

// DataSecurityResult Data security detection result
type DataSecurityResult struct {
	RiskLevel  string   `json:"risk_level"` // Risk level: no_risk, low_risk, medium_risk, high_risk
	Categories []string `json:"categories"` // Sensitive data category list
}

// GuardrailResult Guardrail detection result
type GuardrailResult struct {
	Compliance *ComplianceResult   `json:"compliance"` // Compliance detection result
	Security   *SecurityResult     `json:"security"`   // Security detection result
	Data       *DataSecurityResult `json:"data"`       // Data leakage prevention result
}

// GuardrailResponse Guardrail API response model
type GuardrailResponse struct {
	ID                string           `json:"id"`                  // Request unique identifier
	Result            *GuardrailResult `json:"result"`              // Detection result
	OverallRiskLevel  string           `json:"overall_risk_level"`  // Overall risk level: no_risk, low_risk, medium_risk, high_risk
	SuggestAction     string           `json:"suggest_action"`      // Suggested action: pass, reject, replace
	SuggestAnswer     *string          `json:"suggest_answer"`      // Suggested answer content
	Score             *float64         `json:"score"`               // Detection confidence score
}

// IsSafe Check if the content is safe
func (r *GuardrailResponse) IsSafe() bool {
	return r.SuggestAction == "pass"
}

// IsBlocked Check if the content is blocked
func (r *GuardrailResponse) IsBlocked() bool {
	return r.SuggestAction == "reject"
}

// HasSubstitute Check if there is a substitute
func (r *GuardrailResponse) HasSubstitute() bool {
	return r.SuggestAction == "replace" || r.SuggestAction == "reject"
}

// GetAllCategories Get all risk categories
func (r *GuardrailResponse) GetAllCategories() []string {
	categorySet := make(map[string]bool)
	var categories []string

	if r.Result != nil {
		if r.Result.Compliance != nil {
			for _, category := range r.Result.Compliance.Categories {
				if !categorySet[category] {
					categorySet[category] = true
					categories = append(categories, category)
				}
			}
		}
		if r.Result.Security != nil {
			for _, category := range r.Result.Security.Categories {
				if !categorySet[category] {
					categorySet[category] = true
					categories = append(categories, category)
				}
			}
		}
		if r.Result.Data != nil {
			for _, category := range r.Result.Data.Categories {
				if !categorySet[category] {
					categorySet[category] = true
					categories = append(categories, category)
				}
			}
		}
	}

	return categories
}

// ClientConfig Client configuration
type ClientConfig struct {
	APIKey     string // API key
	BaseURL    string // API base URL
	Timeout    int    // Request timeout (seconds)
	MaxRetries int    // Maximum retry count
}