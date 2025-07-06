package localtree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/files"
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
}

// NewSynchronizer 指定目录树根路径 初始化一个同步器
func NewSynchronizer(baseDir string, pageSize int) *Synchronizer {
	return &Synchronizer{
		baseDir:     baseDir,
		pageSize:    pageSize,
		toSyncTasks: make(chan []FileTask, 1024),
	}
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

// Sync 触发一次同步操作
func (s *Synchronizer) Sync() (added, deleted int, err error) {
	// 初始化状态
	okTaskChan := make(chan FileTask, 1024)
	s.eg, s.ctx = errgroup.WithContext(context.Background())
	defer func() {
		s.eg, s.ctx = nil, nil
	}()

	// 读取根目录放置到任务通道中
	if err := s.walkDir2SyncTasks("/音乐1"); err != nil {
		return 0, 0, fmt.Errorf("获取 openlist 根目录异常: %w", err)
	}

	// 执行同步任务
	go s.handleSyncTasks(okTaskChan)

	// 更新快照和目录树
	s.eg.Go(func() error {
		s.updateLocalTree(okTaskChan, &added, &deleted)
		return nil
	})

	// 等待任务完成
	if err := s.eg.Wait(); err != nil {
		return 0, 0, fmt.Errorf("同步异常: %w", err)
	}
	return
}

// walkDir2SyncTasks 分页遍历 openlist 指定前缀目录下的文件, 加入到任务通道中
func (s *Synchronizer) walkDir2SyncTasks(prefix string) error {
	walker := openlist.WalkFsList(prefix, s.pageSize)
	page, err := walker.Next()
	for err == nil {
		taskList := make([]FileTask, len(page.Content))
		for i, info := range page.Content {
			taskList[i] = FsGetTask(prefix, info)
		}
		s.toSyncTasks <- taskList
		page, err = walker.Next()
	}
	if err != openlist.ErrWalkEOF {
		return err
	}
	return nil
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
				// 文件已存在, 不进行任何操作
				return nil
			}
			if err := os.RemoveAll(localAbsPath); err != nil {
				return fmt.Errorf("删除占用目录异常: %w", err)
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
				return nil
			default:
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
	wg := sync.WaitGroup{}
	for len(s.toSyncTasks) > 0 {
		num := len(s.toSyncTasks)
		for range num {
			tasks := <-s.toSyncTasks
			wg.Add(1)
			s.eg.Go(func() error {
				defer wg.Done()
				return handleTasks(tasks)
			})
		}
		wg.Wait()
	}

	close(okTaskChan)
}

// updateLocalTree 监听 okTaskChan 生成新快照, 并移除本地磁盘中的过期文件, 同时统计变更数
func (s *Synchronizer) updateLocalTree(okTaskChan <-chan FileTask, added, deleted *int) {
	if okTaskChan == nil ||
		added == nil ||
		deleted == nil ||
		s.ctx == nil ||
		s.snapshot == nil {
		return
	}
	current := NewSnapshot()
	*added, *deleted = 0, 0

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
			current.Put(task.LocalPath, task.IsDir)
			// 判断是否是新增
			if _, exists := s.snapshot.Check(task.LocalPath); !exists {
				*added++
			}
		}
	}

	// 统计并删除本地过期文件
	for path := range s.snapshot {
		if _, exists := current.Check(path); exists {
			continue
		}

		path = strings.TrimPrefix(path, "/")
		err := files.ReleasePath(filepath.Join(s.baseDir, path))
		if err != nil {
			logf(colors.Red, "删除过期文件失败: %v", err)
		} else {
			*deleted++
		}
	}

	s.snapshot = current
}
