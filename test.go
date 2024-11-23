package main

import (
	"crypto/tls"
	"log"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func main() {
	server := "127.0.0.1:8000"

	// Establish a TLS connection
	conn, err := tls.Dial("tcp", server, &tls.Config{
		NextProtos: []string{"h2"},
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(http2.ClientPreface))
	if err != nil {
		log.Fatalf("Failed to send client preface: %v", err)
	}
	log.Printf("Sent HTTP2 preface\n")

	framer := http2.NewFramer(conn, conn)
	err = framer.WriteSettings(http2.Setting{
		ID:  http2.SettingMaxConcurrentStreams,
		Val: 100,
	})
	if err != nil {
		log.Fatalf("Failed to send SETTINGS frame: %v", err)
	}

	headerBlockBuf := &strings.Builder{}
	encoder := hpack.NewEncoder(headerBlockBuf)

	err = encoder.WriteField(hpack.HeaderField{Name: ":method", Value: "GET"})
	if err != nil {
		log.Fatalf("Failed to encode method header: %v", err)
	}
	err = encoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: "https"})
	if err != nil {
		log.Fatalf("Failed to encode scheme header: %v", err)
	}
	err = encoder.WriteField(hpack.HeaderField{Name: ":path", Value: "/"})
	if err != nil {
		log.Fatalf("Failed to encode path header: %v", err)
	}
	err = encoder.WriteField(hpack.HeaderField{Name: ":authority", Value: "example.com"})
	if err != nil {
		log.Fatalf("Failed to encode authority header: %v", err)
	}
	err = encoder.WriteField(hpack.HeaderField{Name: "x-test", Value: "custom-value"})
	if err != nil {
		log.Fatalf("Failed to encode custom header: %v", err)
	}
	err = encoder.WriteField(hpack.HeaderField{Name: "content-length", Value: "222"})
	if err != nil {
		log.Fatalf("Failed to encode custom header: %v", err)
	}

	err = framer.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      1,
		BlockFragment: []byte(headerBlockBuf.String()),
		EndStream:     false,
		EndHeaders:    true,
	})
	if err != nil {
		log.Fatalf("Failed to send HEADERS frame: %v", err)
	}

	body := "justatest"
	err = framer.WriteData(1, true, []byte(body))
	if err != nil {
		log.Fatalf("Failed to send DATA frame: %v", err)
	}

	loop:
	for {
		frame, err := framer.ReadFrame()
		if err != nil {
			log.Fatalf("Failed to read frame: %v", err)
		}
		log.Printf("Received frame: %v", frame)
		switch f := frame.(type) {
		case *http2.HeadersFrame:
			if f.StreamEnded() {
				log.Println("Stream ended with headers frame")
				break loop
			}
		case *http2.DataFrame:
			if f.StreamEnded() {
				log.Println("Stream ended with data frame")
				break loop
			}
		default:
			log.Printf("Received unexpected frame: %T\n", frame)
		}
	}
}

