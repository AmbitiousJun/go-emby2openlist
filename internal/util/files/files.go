package files

import (
	"fmt"
	"os"
)

// ReleasePath 释放本地文件路径
//
// 如果 p 是目录, 则整个目录被删除
// 如果 p 是文件, 则文件被删除
func ReleasePath(p string) error {
	stat, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("获取路径文件状态失败 [%s]: %w", p, err)
	}

	if stat.IsDir() {
		if err = os.RemoveAll(p); err != nil {
			return fmt.Errorf("删除目录 %s 失败: %w", p, err)
		}
		return nil
	}

	if err = os.Remove(p); err != nil {
		return fmt.Errorf("删除文件 %s 失败: %w", p, err)
	}
	return nil
}
