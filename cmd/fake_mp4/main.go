package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

func main() {
	var buf bytes.Buffer

	writeBox(&buf, "ftyp", func(b *bytes.Buffer) {
		b.WriteString("isom")
		b.Write([]byte{0x00, 0x00, 0x00, 0x01})
		b.WriteString("isom")
		b.WriteString("iso2")
	})

	writeBox(&buf, "moov", func(moov *bytes.Buffer) {
		// mvhd (全局 movie duration)
		writeBox(moov, "mvhd", func(b *bytes.Buffer) {
			b.WriteByte(0x00)                                     // version
			b.Write([]byte{0x00, 0x00, 0x00})                     // flags
			b.Write(make([]byte, 4))                              // creation_time
			b.Write(make([]byte, 4))                              // modification_time
			binary.Write(b, binary.BigEndian, uint32(1000))       // timescale
			binary.Write(b, binary.BigEndian, uint32(10800*1000)) // duration: 3小时
			binary.Write(b, binary.BigEndian, uint32(0x00010000)) // rate 1.0
			binary.Write(b, binary.BigEndian, uint16(0x0100))     // volume 1.0
			b.Write(make([]byte, 10))                             // reserved
			binary.Write(b, binary.BigEndian, [9]uint32{
				0x00010000, 0, 0,
				0, 0x00010000, 0,
				0, 0, 0x40000000,
			}) // unity matrix
			b.Write(make([]byte, 24))                    // pre-defined
			binary.Write(b, binary.BigEndian, uint32(2)) // next track ID
		})

		// trak (伪轨道)
		writeBox(moov, "trak", func(trak *bytes.Buffer) {
			// tkhd (track header)
			writeBox(trak, "tkhd", func(b *bytes.Buffer) {
				b.WriteByte(0x00)
				b.Write([]byte{0x00, 0x00, 0x07})                     // flags: track enabled, in movie, in preview
				b.Write(make([]byte, 4))                              // creation_time
				b.Write(make([]byte, 4))                              // modification_time
				binary.Write(b, binary.BigEndian, uint32(1))          // track_ID
				b.Write(make([]byte, 4))                              // reserved
				binary.Write(b, binary.BigEndian, uint32(10800*1000)) // duration
				b.Write(make([]byte, 8))                              // reserved
				binary.Write(b, binary.BigEndian, uint16(0))          // layer
				binary.Write(b, binary.BigEndian, uint16(0))          // alternate group
				binary.Write(b, binary.BigEndian, uint16(0))          // volume
				b.Write([]byte{0x00, 0x00})                           // reserved
				binary.Write(b, binary.BigEndian, [9]uint32{
					0x00010000, 0, 0,
					0, 0x00010000, 0,
					0, 0, 0x40000000,
				}) // matrix
				binary.Write(b, binary.BigEndian, uint32(0)) // width
				binary.Write(b, binary.BigEndian, uint32(0)) // height
			})

			// mdia
			writeBox(trak, "mdia", func(mdia *bytes.Buffer) {
				// mdhd
				writeBox(mdia, "mdhd", func(b *bytes.Buffer) {
					b.WriteByte(0x00)
					b.Write([]byte{0x00, 0x00, 0x00})
					b.Write(make([]byte, 4))                              // creation_time
					b.Write(make([]byte, 4))                              // modification_time
					binary.Write(b, binary.BigEndian, uint32(1000))       // timescale
					binary.Write(b, binary.BigEndian, uint32(10800*1000)) // duration
					binary.Write(b, binary.BigEndian, uint16(0x55c4))     // language = und (ISO-639-2/T code)
					b.Write([]byte{0x00, 0x00})                           // pre-defined
				})

				// hdlr (handler type: vide)
				writeBox(mdia, "hdlr", func(b *bytes.Buffer) {
					b.Write([]byte{0x00, 0x00, 0x00, 0x00}) // version + flags
					b.Write(make([]byte, 4))                // pre_defined
					b.Write([]byte("vide"))                 // handler_type
					b.Write(make([]byte, 12))               // reserved
					b.WriteString("Fake Video Handler")     // name
					b.WriteByte(0x00)                       // null terminator
				})

				// minf（可选，不加也能解析）
			})
		})
	})

	if err := os.WriteFile("fake_3h.mp4", buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
	log.Println("fake_3h.mp4 created successfully")
}

func writeBox(parent *bytes.Buffer, boxType string, writePayload func(*bytes.Buffer)) {
	var payload bytes.Buffer
	writePayload(&payload)

	size := uint32(8 + payload.Len())
	binary.Write(parent, binary.BigEndian, size)
	parent.WriteString(boxType)
	parent.Write(payload.Bytes())
}
