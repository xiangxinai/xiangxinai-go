package xiangxinai

// Message 消息模型
type Message struct {
	Role    string `json:"role"`    // 消息角色: user, system, assistant
	Content string `json:"content"` // 消息内容
}

// NewMessage 创建新的消息
func NewMessage(role, content string) *Message {
	return &Message{
		Role:    role,
		Content: content,
	}
}

// GuardrailRequest 护栏检测请求模型
type GuardrailRequest struct {
	Model    string     `json:"model"`    // 模型名称
	Messages []*Message `json:"messages"` // 消息列表
}

// ComplianceResult 合规检测结果
type ComplianceResult struct {
	RiskLevel  string   `json:"risk_level"` // 风险等级: 无风险, 低风险, 中风险, 高风险
	Categories []string `json:"categories"` // 风险类别列表
}

// SecurityResult 安全检测结果
type SecurityResult struct {
	RiskLevel  string   `json:"risk_level"` // 风险等级: 无风险, 低风险, 中风险, 高风险
	Categories []string `json:"categories"` // 风险类别列表
}

// GuardrailResult 护栏检测结果
type GuardrailResult struct {
	Compliance *ComplianceResult `json:"compliance"` // 合规检测结果
	Security   *SecurityResult   `json:"security"`   // 安全检测结果
}

// GuardrailResponse 护栏API响应模型
type GuardrailResponse struct {
	ID                string           `json:"id"`                  // 请求唯一标识
	Result            *GuardrailResult `json:"result"`              // 检测结果
	OverallRiskLevel  string           `json:"overall_risk_level"`  // 综合风险等级: 无风险, 低风险, 中风险, 高风险
	SuggestAction     string           `json:"suggest_action"`      // 建议动作: 通过, 阻断, 代答
	SuggestAnswer     *string          `json:"suggest_answer"`      // 建议回答内容
}

// IsSafe 判断内容是否安全
func (r *GuardrailResponse) IsSafe() bool {
	return r.SuggestAction == "通过"
}

// IsBlocked 判断内容是否被阻断
func (r *GuardrailResponse) IsBlocked() bool {
	return r.SuggestAction == "阻断"
}

// HasSubstitute 判断是否有代答
func (r *GuardrailResponse) HasSubstitute() bool {
	return r.SuggestAction == "代答" || r.SuggestAction == "阻断"
}

// GetAllCategories 获取所有风险类别
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
	}
	
	return categories
}

// ClientConfig 客户端配置
type ClientConfig struct {
	APIKey     string // API密钥
	BaseURL    string // API基础URL
	Timeout    int    // 请求超时时间（秒）
	MaxRetries int    // 最大重试次数
}