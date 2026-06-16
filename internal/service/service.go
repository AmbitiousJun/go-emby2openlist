package service

import (
	"net/http"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/model"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/gin-gonic/gin"
)

// ValidateApiSecret 校验 API Secret 是否正确
func ValidateApiSecret(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusOK, model.Response{Message: "不支持的请求方式"})
		return
	}

	var reqData ValidateSecretRequest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusOK, model.Response{Message: "请求参数错误"})
		logs.Error("校验程序密钥失败, 请求参数转换错误: %v", err)
		return
	}

	if reqData.Secret == config.C.Ge2o.ApiSecret {
		c.JSON(http.StatusOK, model.Response{
			Success: true,
			Message: "校验通过",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{Message: "校验失败"})
}
