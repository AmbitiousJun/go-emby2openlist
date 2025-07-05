package ffmpeg_test

import (
	"log"
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
)

func TestInspectInfo(t *testing.T) {
	if err := ffmpeg.AutoDownloadExec("../../../.."); err != nil {
		t.Fatal(err)
		return
	}

	i, err := ffmpeg.InspectInfo("/Users/ambitious/Downloads/test.mp4")
	if err != nil {
		t.Fatal(err)
		return
	}
	log.Println(i)
}
