package main

import (
	"log"
	"math"
	"os"
	"time"
)

func main() {
	const (
		duration      = time.Minute * 4
		bitrateKbps   = 32 // 比特率 32 kbps
		sampleRate    = 44100
		frameSize     = 105                 // 32kbps MPEG1 Layer3 44100Hz 计算出来的帧大小
		frameDuration = 1152.0 / sampleRate // 每帧时长
	)

	numFrames := int(math.Round(duration.Seconds() / frameDuration))

	log.Printf("Generating %d fake MP3 frames...\n", numFrames)

	f, err := os.Create("fake_3h.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 生成伪造帧
	for range numFrames {
		// 帧头（MPEG-1 Layer III, 32kbps, 44100Hz, 无CRC）
		f.Write([]byte{0xFF, 0xFB, 0x84, 0x64})
		// 填充假的帧数据
		padding := make([]byte, frameSize-4)
		f.Write(padding)
	}

	log.Println("fake_3h.mp3 created")
}
