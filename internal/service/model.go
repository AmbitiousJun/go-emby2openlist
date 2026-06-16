package service

// ValidateSecretRequest 校验程序密钥的请求体
type ValidateSecretRequest struct {
	Secret string `json:"secret"`
}
