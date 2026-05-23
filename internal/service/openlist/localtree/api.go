package localtree

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/model"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
	"github.com/gin-gonic/gin"
)

// updatePrefixValidateRegex 校验请求的路径前缀
var updatePrefixValidateRegex = regexp.MustCompile(constant.Reg_OpenlistLocalTreeUpdatePrefix)

// UpdateManually 手动跟新本地目录树
func UpdateManually(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusNotFound, "404 not found")
		return
	}

	// 获取本地密钥
	localSecret := config.C.Openlist.LocalTreeGen.ApiSecret
	if localSecret = strings.TrimSpace(localSecret); localSecret == "" {
		c.JSON(http.StatusOK, model.Response{Message: "请先配置本地密钥"})
		return
	}

	// 确保模块已经初始化完毕
	if synchronizer == nil {
		c.JSON(http.StatusOK, model.Response{Message: "请先启用本地目录树模块"})
		return
	}

	// 转换请求体
	var reqData UpdateRequest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusOK, model.Response{Message: "请求参数错误"})
		return
	}

	// 校验本地密钥
	if reqData.Secret != localSecret {
		c.JSON(http.StatusOK, model.Response{Message: "密钥错误"})
		return
	}

	// 校验路径
	if reqData.Prefix = strings.TrimSpace(reqData.Prefix); reqData.Prefix != "" {
		if !updatePrefixValidateRegex.MatchString(reqData.Prefix) {
			c.JSON(http.StatusOK, model.Response{Message: "无效的路径前缀"})
			return
		}
	}

	// 执行同步
	errChan := make(chan error)
	go func() {
		if err := doSync(synchronizer, reqData.Prefix); err != nil {
			logf(colors.Red, "同步失败: %v", err)
			errChan <- err
		}
	}()

	timeoutTimer := time.NewTimer(time.Second * 2)
	select {
	case err := <-errChan:
		c.JSON(http.StatusOK, model.Response{Message: "同步失败: " + err.Error()})
	case <-timeoutTimer.C:
		msg := "调用成功, 开始全量扫描, 详情请查看容器日志..."
		if reqData.Prefix != "" {
			msg = "调用成功, 扫描路径: [" + reqData.Prefix + "], 详情请查看容器日志..."
		}
		c.JSON(http.StatusOK, model.Response{
			Success: true,
			Message: msg,
		})
	}

}
