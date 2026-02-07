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
	if a.LocalTreeGen == nil {
		a.LocalTreeGen = new(LocalTreeGen)
	}
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

	// MusicContainers 音乐媒体容器, 原始串, 以英文逗号分割
	MusicContainers string `yaml:"music-containers"`

	// AutoRemoveMaxCount 自动删除文件的最大数量
	AutoRemoveMaxCount int `yaml:"auto-remove-max-count"`

	// RefreshInterval 刷新间隔, 单位: 分钟
	RefreshInterval int `yaml:"refresh-interval"`

	// ScanPrefixes 指定扫描前缀
	ScanPrefixes []string `yaml:"scan-prefixes"`

	// AllowContainers 允许处理的容器,为空则允许所有
	AllowContainers string `yaml:"allow-containers"`

	// Threads 同步线程数
	Threads int `yaml:"threads"`

	// virtualContainers 虚拟媒体容器集合 便于快速查询
	virtualContainers map[string]struct{}

	// strmContainers strm 媒体容器集合 便于快速查询
	strmContainers map[string]struct{}

	// musicContainers 音乐媒体容器集合 便于快速查询
	musicContainers map[string]struct{}

	// allowContainers 允许处理的容器集合 便于快速查询
	allowContainers map[string]struct{}
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

	if ltg.AutoRemoveMaxCount < 0 {
		ltg.AutoRemoveMaxCount = 0
	}

	if ltg.RefreshInterval <= 0 {
		return fmt.Errorf("无效刷新间隔: %d", ltg.RefreshInterval)
	}

	if len(ltg.ScanPrefixes) == 0 {
		// 没有配置则全量扫描
		ltg.ScanPrefixes = append(ltg.ScanPrefixes, "/")
	}
	for _, prefix := range ltg.ScanPrefixes {
		prefix = strings.TrimSpace(prefix)
		if !strings.HasPrefix(prefix, "/") {
			return fmt.Errorf("无效扫描前缀: [%s], 必须以 / 开头", prefix)
		}
	}

	if ltg.Threads == 0 {
		// 默认线程数
		ltg.Threads = 8
	}
	if ltg.Threads < 0 {
		return fmt.Errorf("无效同步线程数: [%d], 必须配置为大于 0 的值", ltg.Threads)
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

	ss = strings.Split(strings.TrimSpace(ltg.MusicContainers), ",")
	ltg.musicContainers = make(map[string]struct{}, len(ss))
	for _, s := range ss {
		ltg.musicContainers[strings.ToLower(s)] = struct{}{}
	}

	ss = strings.Split(strings.TrimSpace(ltg.AllowContainers), ",")
	ltg.allowContainers = make(map[string]struct{}, len(ss))
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			ltg.allowContainers[strings.ToLower(s)] = struct{}{}
		}
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

// IsMusic 判断一个容器是否属于音乐容器
func (ltg *LocalTreeGen) IsMusic(container string) bool {
	container = strings.ToLower(container)
	_, ok := ltg.musicContainers[container]
	return ok
}

// IsAllowed 判断一个容器是否被允许处理
// 逻辑规则:
// 1. strm-containers, virtual-containers, music-containers 中的容器自动被允许
// 2. allow-containers 中明确指定的容器也被允许
// 3. 如果以上都未配置,则拒绝所有容器(避免意外下载所有文件)
func (ltg *LocalTreeGen) IsAllowed(container string) bool {
	container = strings.ToLower(container)
	
	// 检查是否在 strm/virtual/music 容器中
	if _, ok := ltg.strmContainers[container]; ok {
		return true
	}
	if _, ok := ltg.virtualContainers[container]; ok {
		return true
	}
	if _, ok := ltg.musicContainers[container]; ok {
		return true
	}
	
	// 检查是否在额外允许列表中
	if len(ltg.allowContainers) > 0 {
		_, ok := ltg.allowContainers[container]
		return ok
	}
	
	// 如果都没有配置,则拒绝所有(安全起见)
	return false
}

// IsValidPrefix 判断一个 openlist 路径是否在扫描前缀的范围中
func (ltg *LocalTreeGen) IsValidPrefix(path string) bool {
	for _, prefix := range ltg.ScanPrefixes {
		if strings.HasPrefix(path, prefix) || strings.HasPrefix(prefix, path) {
			return true
		}
	}
	return false
}
