package localtree

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/music"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/bytess"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/mp4s"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/trys"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/urls"
)

// FileTask 包含同步必要信息的文件结构
type FileTask struct {
	// Path 文件绝对路径, 与 openlist 对应
	Path string

	// LocalPath 文件要存入本地的路径
	LocalPath string

	// IsDir 是否是目录
	IsDir bool

	// Container 标记文件的容器
	Container string

	// Sign openlist 文件签名
	Sign string

	// Modified 文件的最后修改时间
	Modified time.Time
}

func FsGetTask(prefix string, info openlist.FsGet) FileTask {
	container := strings.TrimPrefix(strings.ToLower(filepath.Ext(info.Name)), ".")
	fp := filepath.Join(prefix, info.Name)
	return FileTask{
		Path:      urls.TransferSlash(fp),
		LocalPath: fp,
		IsDir:     info.IsDir,
		Sign:      info.Sign,
		Container: container,
		Modified:  info.Modified,
	}
}

// TaskWriter 将 openlist 文件写入到本地文件系统
type TaskWriter interface {

	// Path 将 openlist 文件路径中的文件名
	// 转换为本地文件系统中的文件名
	Path(path string) string

	// Write 将文件信息写入到本地文件系统中
	Write(task FileTask, localPath string) error
}

var (
	vw = VirtualWriter{}
	sw = StrmWriter{}
	mw = MusicWriter{}
	rw = RawWriter{}
)

// LoadTaskWriter 根据文件容器加载 TaskWriter
func LoadTaskWriter(container string) TaskWriter {
	cfg := config.C.Openlist.LocalTreeGen
	if cfg.IsVirtual(container) {
		return &vw
	}
	if cfg.IsStrm(container) {
		return &sw
	}
	if cfg.IsMusic(container) {
		return &mw
	}
	return &rw
}

// VirtualWriter 写同名空文件, 尝试写入媒体时长
type VirtualWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (vw *VirtualWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (vw *VirtualWriter) Write(task FileTask, localPath string) error {
	// 默认写入时长 3 小时
	dftDuration := time.Hour * 3
	if !config.C.Openlist.LocalTreeGen.FFmpegEnable {
		return os.WriteFile(localPath, mp4s.GenWithDuration(dftDuration), os.ModePerm)
	}

	var info ffmpeg.Info
	err := trys.Try(func() (err error) {
		info, err = ffmpeg.InspectInfo(getRealDownloadUrl(task))
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("调用 ffmpeg 失败: %w", err)
	}

	if err := os.WriteFile(localPath, mp4s.GenWithDuration(info.Duration), os.ModePerm); err != nil {
		return err
	}

	abs, err := filepath.Abs(localPath)
	if err != nil {
		abs = localPath
	}
	logf(colors.Gray, "生成虚拟文件 [%s]: [时长: %v]", abs, info.Duration)
	return nil
}

// StrmWriter 写文件对应的 openlist strm 文件
type StrmWriter struct{}

// OpenlistPath 生成媒体的 openlist http 访问地址
func (sw *StrmWriter) OpenlistPath(task FileTask) string {
	segs := strings.Split(strings.TrimPrefix(task.Path, "/"), "/")
	for i, seg := range segs {
		segs[i] = url.PathEscape(seg)
	}

	return fmt.Sprintf(
		"%s/d/%s?sign=%s",
		config.C.Openlist.Host,
		strings.Join(segs, "/"),
		task.Sign,
	)
}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (sw *StrmWriter) Path(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext) + ".strm"
}

// Write 将文件信息写入到本地文件系统中
func (sw *StrmWriter) Write(task FileTask, localPath string) error {
	if err := os.WriteFile(localPath, []byte(sw.OpenlistPath(task)), os.ModePerm); err != nil {
		return err
	}

	abs, err := filepath.Abs(localPath)
	if err != nil {
		abs = localPath
	}
	logf(colors.Gray, "生成 strm: [%s]", abs)

	return nil
}

// MusicWriter 写同名空文件, 同时尝试写入时长和音乐标签元数据信息
type MusicWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (mw *MusicWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (mw *MusicWriter) Write(task FileTask, localPath string) error {
	if !config.C.Openlist.LocalTreeGen.FFmpegEnable {
		// 必须开启 ffmpeg 才能生成, 改用 strm 替代
		return sw.Write(task, localPath)
	}

	var meta ffmpeg.Music
	err := trys.Try(func() (err error) {
		meta, err = ffmpeg.InspectMusic(getRealDownloadUrl(task))
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("提取音乐元数据失败 [%s]: %w", filepath.Base(task.Path), err)
	}
	if meta.Duration == 0 {
		meta.Duration = time.Second
	}

	var pic []byte
	err = trys.Try(func() (err error) {
		pic, err = ffmpeg.ExtractMusicCover(getRealDownloadUrl(task))
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("提取音乐封面失败 [%s]: %w", filepath.Base(task.Path), err)
	}

	if err := music.WriteFakeMP3(localPath, meta, pic); err != nil {
		return err
	}

	abs, err := filepath.Abs(localPath)
	if err != nil {
		abs = localPath
	}
	logf(colors.Gray, "生成音乐虚拟文件 [%s]: [标题: %s] [艺术家: %s] [时长: %v]", abs, meta.Title, meta.Artist, meta.Duration)
	return nil
}

// RawWriter 请求 openlist 源文件写入本地
type RawWriter struct {
	mu sync.Mutex
}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (rw *RawWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (rw *RawWriter) Write(task FileTask, localPath string) error {
	// 防止并发访问网盘触发风控
	rw.mu.Lock()
	defer rw.mu.Unlock()

	header := http.Header{"User-Agent": []string{ffmpeg.DefaultUserAgent}}

	err := trys.Try(func() (err error) {
		logf(colors.Yellow, "尝试下载 openlist 源文件, 路径: [%s]", localPath)

		file, err := os.Create(localPath)
		if err != nil {
			return fmt.Errorf("创建文件失败 [%s]: %w", localPath, err)
		}
		defer file.Close()

		resp, err := https.Get(sw.OpenlistPath(task)).Header(header).Do()
		if err != nil {
			return fmt.Errorf("请求 openlist 直链失败: %w", err)
		}
		defer resp.Body.Close()

		if !https.IsSuccessCode(resp.StatusCode) {
			return fmt.Errorf("请求 openlist 直链失败, 响应状态: %s", resp.Status)
		}

		buf := bytess.CommonFixedBuffer()
		defer buf.PutBack()
		if _, err = io.CopyBuffer(file, resp.Body, buf.Bytes()); err != nil {
			return fmt.Errorf("写入 openlist 源文件到本地磁盘失败, 拷贝异常: %w", err)
		}

		logf(colors.Gray, "openlist 源文件 [%s] 已写入本地", filepath.Base(task.Path))
		return
	}, 3, time.Second*5)

	return err
}

// getRealDownloadUrl 获取真实的下载链接，跟随 302 跳转
func getRealDownloadUrl(task FileTask) string {
	// 构建 openlist 的 /d/ 路径
	openlistUrl := sw.OpenlistPath(task)

	// 发送请求并跟随重定向
	finalUrl, resp, err := https.
		Get(openlistUrl).
		AddHeader("User-Agent", ffmpeg.DefaultUserAgent).
		DoRedirect()
	if err != nil {
		logs.Warn("获取真实下载链接失败: %w", err)
		return openlistUrl
	}
	defer resp.Body.Close()

	if !https.IsSuccessCode(resp.StatusCode) {
		logs.Warn("获取真实下载链接失败: %s", resp.Status)
		return openlistUrl
	}

	return finalUrl
}
