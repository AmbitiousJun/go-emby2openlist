package openlist

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrWalkEOF 标记分页遍历结束
var ErrWalkEOF = errors.New("walk EOF")

// Walker 分页遍历接口
type Walker[T any] struct {

	// curPage 当前页
	curPage int

	// Next 获取下一页数据
	Next func() (T, error)
}

// FetchFsList 请求 openlist "/api/fs/list" 接口, 支持分页
//
// 传入 path 与接口的 path 作用一致
func WalkFsList(path string, perPage int) *Walker[FsList] {
	w := Walker[FsList]{curPage: 1}

	w.Next = func() (FsList, error) {
		if w.curPage < 1 {
			return FsList{}, ErrWalkEOF
		}

		var res FsList
		err := Fetch("/api/fs/list", http.MethodPost, nil, map[string]any{
			"refresh":  false,
			"password": "",
			"path":     path,
			"page":     w.curPage,
			"per_page": perPage,
		}, &res)
		if err != nil {
			return FsList{}, fmt.Errorf("FsList 请求失败: %w", err)
		}
		w.curPage++

		if len(res.Content) == 0 {
			w.curPage = -1
			return FsList{}, ErrWalkEOF
		}

		return res, nil
	}

	return &w
}
