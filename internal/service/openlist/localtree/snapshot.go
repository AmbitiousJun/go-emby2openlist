package localtree

// Snapshot 记录本地目录树的快照
type Snapshot map[string]struct{ IsDir bool }

// NewSnapshot 创建一个新的快照实例
func NewSnapshot() Snapshot { return make(Snapshot) }

// Put 设置一个路径的信息
func (s Snapshot) Put(path string, dir bool) {
	s[path] = struct{ IsDir bool }{dir}
}

// Check 检查一个路径, 返回路径信息
func (s Snapshot) Check(path string) (isDir, exists bool) {
	res, ok := s[path]
	if !ok {
		return false, false
	}
	return res.IsDir, true
}
