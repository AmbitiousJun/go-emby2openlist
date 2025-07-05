package music

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/dhowden/tag"
)

// MusicNFO 音乐元数据
type MusicNFO struct {
	XMLName xml.Name `xml:"music"`
	Title   string   `xml:"title,omitempty"`
	Artist  string   `xml:"artist,omitempty"`
	Album   string   `xml:"album,omitempty"`
	Year    int      `xml:"year,omitempty"`
	Genre   string   `xml:"genre,omitempty"`
	Track   int      `xml:"track,omitempty"`
	Lyrics  string   `xml:"lyrics,omitempty"`
	Comment string   `xml:"comment,omitempty"`
}

// WriteNFO 写入音乐元数据到本地
func WriteNFO(filePath string, meta tag.Metadata) error {
	if meta == nil {
		return fmt.Errorf("元数据为空")
	}

	track, _ := meta.Track()
	info := MusicNFO{
		Title:   meta.Title(),
		Artist:  meta.Artist(),
		Album:   meta.Album(),
		Year:    meta.Year(),
		Genre:   meta.Genre(),
		Track:   track,
		Lyrics:  meta.Lyrics(),
		Comment: meta.Comment(),
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("无法创建文件: %w", err)
	}
	defer f.Close()

	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	return enc.Encode(info)
}

// WritePic 写入音乐封面到本地
func WritePic(filePath string, meta tag.Metadata) error {
	if meta == nil || meta.Picture() == nil {
		return fmt.Errorf("元数据为空")
	}
	return os.WriteFile(filePath, meta.Picture().Data, os.ModePerm)
}
