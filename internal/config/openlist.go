package config

import (
	"fmt"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
)

type Openlist struct {
	// Token 访问 openlist 接口的密钥, 在 openlist 管理后台获取
	Token string `yaml:"token"`
	// Host openlist 访问地址（如果 openlist 使用本地代理模式, 则这个地址必须配置公网可访问地址）
	Host string `yaml:"host"`

	// LocalTreeGen 本地目录树生成相关
	LocalTreeGen *LocalTreeGen `yaml:"local-tree-gen"`
}

func (a *Openlist) Init() error {
	if err := a.LocalTreeGen.Init(); err != nil {
		return fmt.Errorf("openlist.local-tree-gen 配置错误: %w", err)
	}

	return nil
}

type LocalTreeGen struct {

	// Enable 是否启用
	Enable bool `yaml:"enable"`

	// FFmpegEnable 是否启用 ffmpeg
	FFmpegEnable bool `yaml:"ffmpeg-enable"`

	// VirtualContainers 虚拟媒体容器, 原始串, 以英文逗号分割
	VirtualContainers string `yaml:"virtual-containers"`

	// StrmContainers strm 媒体容器, 原始串, 以英文逗号分割
	StrmContainers string `yaml:"strm-containers"`

	// RefreshInterval 刷新间隔, 单位: 分钟
	RefreshInterval int `yaml:"refresh-interval"`

	// virtualContainers 虚拟媒体容器集合 便于快速查询
	virtualContainers map[string]struct{}

	// strmContainers strm 媒体容器集合 便于快速查询
	strmContainers map[string]struct{}
}

// Init 配置初始化
func (ltg *LocalTreeGen) Init() error {
	if !ltg.Enable {
		return nil
	}

	if ltg.FFmpegEnable {
		if err := ffmpeg.AutoDownloadExec(BasePath); err != nil {
			return fmt.Errorf("ffmpeg 初始化失败: %w", err)
		}
	}

	if ltg.RefreshInterval <= 0 {
		return fmt.Errorf("无效刷新间隔: %d", ltg.RefreshInterval)
	}

	ss := strings.Split(strings.TrimSpace(ltg.VirtualContainers), ",")
	ltg.virtualContainers = make(map[string]struct{}, len(ss))
	for _, s := range ss {
		ltg.virtualContainers[strings.ToLower(s)] = struct{}{}
	}

	ss = strings.Split(strings.TrimSpace(ltg.StrmContainers), ",")
	ltg.strmContainers = make(map[string]struct{}, len(ss))
	for _, s := range ss {
		ltg.strmContainers[strings.ToLower(s)] = struct{}{}
	}

	return nil
}

// IsVirtual 判断一个容器是否属于虚拟容器
func (ltg *LocalTreeGen) IsVirtual(container string) bool {
	container = strings.ToLower(container)
	_, ok := ltg.virtualContainers[container]
	return ok
}

// IsStrm 判断一个容器是否属于 strm 容器
func (ltg *LocalTreeGen) IsStrm(container string) bool {
	container = strings.ToLower(container)
	_, ok := ltg.strmContainers[container]
	return ok
}
