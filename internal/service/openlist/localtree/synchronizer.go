package localtree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/files"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/trys"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/urls"
	"golang.org/x/sync/errgroup"
)

// Synchronizer 同步远程 openlist 信息为本地磁盘目录树
type Synchronizer struct {
	// snapshot 同步过程中, 实时维护快照信息
	snapshot Snapshot

	// baseDir 本地目录树生成根路径
	baseDir string

	// pageSize 分页请求每页大小
	pageSize int

	// ctx 控制子任务执行和及时退出
	ctx context.Context

	// eg 并发同步的执行组
	eg *errgroup.Group

	// toSyncTasks 每次同步时的待处理子任务存放通道
	toSyncTasks chan []FileTask

	// activeTaskCount 记录 BFS 遍历过程中的活跃任务数
	activeTaskCount int32

	// threadsSem 控制同步任务的并发线程数
	threadsSem chan struct{}

	// hasScanTotal 当前已扫描到的总任务数
	hasScanTotal int64

	// hasScanFinish 当前已处理完成的任务数
	hasScanFinish int64
}

// NewSynchronizer 指定目录树根路径 初始化一个同步器
func NewSynchronizer(baseDir string, pageSize int) *Synchronizer {
	return &Synchronizer{
		baseDir:     baseDir,
		pageSize:    pageSize,
		toSyncTasks: make(chan []FileTask, 1024),
	}
}

// Sync 触发一次同步操作
func (s *Synchronizer) Sync() (total, added, deleted int, err error) {
	if err := s.InitSnapshot(); err != nil {
		return 0, 0, 0, fmt.Errorf("初始化快照异常: %w", err)
	}

	// 初始化状态
	s.toSyncTasks = make(chan []FileTask, 1024)
	okTaskChan := make(chan FileTask, 1024)
	s.eg, s.ctx = errgroup.WithContext(context.Background())
	s.threadsSem = make(chan struct{}, config.C.Openlist.LocalTreeGen.Threads)
	s.hasScanFinish, s.hasScanTotal = 0, 0

	// 读取根目录放置到任务通道中
	s.activeTaskCount = 0
	if err := s.walkDir2SyncTasks("/"); err != nil {
		return 0, 0, 0, fmt.Errorf("获取 openlist 根目录异常: %w", err)
	}

	// 每隔固定时间输出一下当前的同步进度
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			var percent float64
			if s.hasScanTotal > 0 {
				percent = float64(s.hasScanFinish) / float64(s.hasScanTotal)
			}
			logf(colors.Purple, "预估同步进度 (已扫描/已发现任务数) => %d/%d (%.2f%%)", s.hasScanFinish, s.hasScanTotal, percent*100)
		}
	}()

	// 执行同步任务
	go s.handleSyncTasks(okTaskChan)

	// 更新快照和目录树
	s.eg.Go(func() error {
		s.updateLocalTree(okTaskChan, &total, &added, &deleted)
		return nil
	})

	// 等待任务完成
	if err := s.eg.Wait(); err != nil {
		return 0, 0, 0, fmt.Errorf("同步异常: %w", err)
	}
	return
}

// InitSnapshot 扫描本地磁盘 初始化快照
func (s *Synchronizer) InitSnapshot() error {
	ss := NewSnapshot()

	// 检查根目录
	stat, err := os.Stat(s.baseDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("根目录扫描异常: %w", err)
		}
		if err = os.MkdirAll(s.baseDir, os.ModePerm); err != nil {
			return fmt.Errorf("初始化根目录异常: %w", err)
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("根目录被占用: [%s]", s.baseDir)
	}

	// 递归扫描目录树
	err = filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("无法访问路径 %s: %w", path, err)
		}

		// 获取相对路径
		relPath, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return fmt.Errorf("无法获取相对路径 %s: %w", path, err)
		}
		if relPath == "." {
			return nil
		}
		relPath = urls.TransferSlash(relPath)

		// 跳过隐藏目录文件
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		ss.Put("/"+relPath, info.IsDir())
		return nil
	})
	if err != nil {
		return fmt.Errorf("磁盘扫描异常: %w", err)
	}

	s.snapshot = ss
	return nil
}

// walkDir2SyncTasks 分页遍历 openlist 指定前缀目录下的文件, 加入到任务通道中
func (s *Synchronizer) walkDir2SyncTasks(prefix string) error {
	walker := openlist.WalkFsList(prefix, s.pageSize)
	var page openlist.FsList
	var err error

	eof := false
	err = trys.Try(func() (innerErr error) {
		page, innerErr = walker.Next()
		if innerErr == openlist.ErrWalkEOF {
			eof = true
			return nil
		}
		return
	}, 3, time.Second*5)

	for err == nil && !eof {
		taskList := make([]FileTask, len(page.Content))
		for i, info := range page.Content {
			taskList[i] = FsGetTask(prefix, info)
		}
		s.toSyncTasks <- taskList
		atomic.AddInt32(&s.activeTaskCount, 1)
		atomic.AddInt64(&s.hasScanTotal, int64(len(taskList)))

		err = trys.Try(func() (innerErr error) {
			page, innerErr = walker.Next()
			if innerErr == openlist.ErrWalkEOF {
				eof = true
				return nil
			}
			return
		}, 3, time.Second*5)
	}
	return err
}

// handleSyncTasks 广度遍历 toSyncs 任务通道进行同步
//
// 将同步完成的任务写入 okTaskChan 中
// 所有任务同步完成后, 自动关闭 okTaskChan
func (s *Synchronizer) handleSyncTasks(okTaskChan chan<- FileTask) {
	if okTaskChan == nil {
		return
	}

	if s.ctx == nil || s.eg == nil {
		return
	}

	// handleDir 处理目录, 请求下一层级数据, 并写入任务通道
	handleDir := func(task FileTask) error {
		localAbsPath := filepath.Join(s.baseDir, strings.TrimPrefix(task.LocalPath, "/"))
		if err := os.MkdirAll(localAbsPath, os.ModePerm); err != nil {
			return fmt.Errorf("初始化目录异常 [%s]: %w", localAbsPath, err)
		}

		if err := s.walkDir2SyncTasks(task.Path); err != nil {
			return fmt.Errorf("扫描 openlist 目录异常 [%s]: %w", task.Path, err)
		}
		return nil
	}

	// handleFile 处理文件, 根据容器类型以不同方式写入本地
	handleFile := func(task *FileTask) error {
		if task == nil {
			return nil
		}

		// 获取适配容器的 writer
		writer := LoadTaskWriter(task.Container)

		// 将 openlist 路径转换为本地磁盘相应路径
		task.LocalPath = writer.Path(task.Path)
		localAbsPath := filepath.Join(s.baseDir, strings.TrimPrefix(task.LocalPath, "/"))

		// 如果路径被目录占用, 则删除目录
		stat, err := os.Stat(localAbsPath)
		if err == nil {
			if !stat.IsDir() {
				// 文件已存在
				// 根据本地文件的修改时间和远程文件的修改时间判断文件是否发生变更
				if stat.ModTime().After(task.Modified) {
					return nil
				}
			}
			if err := os.RemoveAll(localAbsPath); err != nil {
				return fmt.Errorf("删除占用路径异常 [%s]: %w", localAbsPath, err)
			}
		}

		if err := os.MkdirAll(filepath.Dir(localAbsPath), os.ModePerm); err != nil {
			return fmt.Errorf("初始化父目录异常 [%s]: %w", localAbsPath, err)
		}

		// 写入文件
		return writer.Write(*task, localAbsPath)
	}

	// handleTasks 处理任务, 将新增的文件写入本地, 任务处理完成后写入 okTaskChan
	handleTasks := func(tasks []FileTask) error {
		for _, task := range tasks {
			select {
			case <-s.ctx.Done():
				return s.ctx.Err()
			default:
				// 更新同步进度
				defer atomic.AddInt64(&s.hasScanFinish, 1)

				// 根据用户配置忽略特定文件和目录
				cfg := config.C.Openlist.LocalTreeGen
				if !cfg.IsValidPrefix(task.Path) {
					continue
				}
				if !task.IsDir && cfg.IsIgnore(task.Container) {
					continue
				}

				// 处理任务
				if task.IsDir {
					if err := handleDir(task); err != nil {
						return err
					}
				} else {
					if err := handleFile(&task); err != nil {
						return err
					}
				}

				// 当前任务写入 okTaskChan
				okTaskChan <- task
			}
		}
		return nil
	}

	// BFS
	for tasks := range s.toSyncTasks {
		s.eg.Go(func() error {
			// 获取线程信号量
			select {
			case s.threadsSem <- struct{}{}:
			case <-s.ctx.Done():
				return s.ctx.Err()
			}

			// 释放线程
			defer func() { <-s.threadsSem }()

			// 更新活跃任务数
			defer func() {
				if atomic.AddInt32(&s.activeTaskCount, -1) == 0 {
					close(s.toSyncTasks)
				}
			}()

			return handleTasks(tasks)
		})
	}

	close(okTaskChan)
}

// updateLocalTree 监听 okTaskChan 生成新快照, 并移除本地磁盘中的过期文件, 同时统计变更数
func (s *Synchronizer) updateLocalTree(okTaskChan <-chan FileTask, total, added, deleted *int) {
	if okTaskChan == nil ||
		added == nil ||
		deleted == nil ||
		s.ctx == nil ||
		s.snapshot == nil {
		return
	}
	current := NewSnapshot()
	*added, *deleted, *total = 0, 0, 0

	// 循环处理任务
	chanOpen := true
	for chanOpen {
		select {
		case <-s.ctx.Done():
			return
		case task, ok := <-okTaskChan:
			if !ok {
				chanOpen = false
				break
			}
			cleanLocalPath := urls.TransferSlash(task.LocalPath)
			current.Put(cleanLocalPath, task.IsDir)
			*total++
			// 判断是否是新增
			if _, exists := s.snapshot.Check(cleanLocalPath); !exists {
				*added++
			}
		}
	}

	toDelete := make([]string, 0, 1<<6)

	// 统计并删除本地过期文件
	for path := range s.snapshot {
		if _, exists := current.Check(path); exists {
			continue
		}

		path = strings.TrimPrefix(path, "/")
		toDelete = append(toDelete, filepath.Join(s.baseDir, path))
	}

	maxCount := config.C.Openlist.LocalTreeGen.AutoRemoveMaxCount
	if len(toDelete) > maxCount {
		logf(colors.Yellow, "过期文件数量 [%d] 超出最大限制 [%d], 跳过删除操作", len(toDelete), maxCount)
		return
	}

	for _, path := range toDelete {
		if err := files.ReleasePath(path); err != nil {
			logf(colors.Red, "删除过期文件失败: %v", err)
			continue
		}
		*deleted++
	}

	s.snapshot = nil
}
