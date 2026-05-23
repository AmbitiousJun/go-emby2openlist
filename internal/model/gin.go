package model

// Response gin 通用响应结构
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
