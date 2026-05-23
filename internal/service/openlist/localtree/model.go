package localtree

// UpdateRequest 手动更新目录树接口请求封装
type UpdateRequest struct {
	Prefix string `json:"prefix"` // 扫描前缀, 为空时全量扫描
	Secret string `json:"secret"` // 内部密钥
}