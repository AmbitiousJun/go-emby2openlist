package ffmpeg

import "time"

// Info 记录文件元信息
type Info struct {
	Duration time.Duration
}

// Music 记录音乐元信息
type Music struct {
	Info
	Album       string // 专辑
	Artist      string // 艺术家
	Comment     string // 备注
	Date        string // 发布日期
	Lyrics      string // 歌词
	Title       string // 标题
	Track       string // 轨道 (track/tracktotal)
	Disc        string // 光盘 (disc/disctotal)
	Genre       string // 流派
	AlbumArtist string // 专辑艺术家
}
