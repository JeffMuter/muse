package main

import (
	"context"
	"fmt"
	"log"
	"muse/pelican/audio"
	"os/exec"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
)

type StreamProcessor struct {
	inputStreams []string
	outputAddr   string
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processes    []*exec.Cmd
	mu           sync.Mutex
}

func main() {

	// rtsp listening
	client := gortsplib.Client{}

	// parse url
	url, err := base.ParseURL("rtsp://localhost:8554/stream")
	if err != nil {
		panic(err)
	}

	// connect server to stream
	err = client.Start(url.Scheme, url.Host)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	desc, _, err := client.Describe(url)
	if err != nil {
		panic(err)
	}

	err = client.SetupAll(desc.BaseURL, desc.Medias)
	if err != nil {
		panic(err)
	}

	audioHandler := audio.NewAudioHandler()

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			outputPath := fmt.Sprintf("audio_%d.wav", time.Now().Unix())
			if err := audioHandler.ConvertToWav(outputPath); err != nil {
				log.Printf("Failed to convert audio: %v", err)
				continue
			}
			// TODO: Send to gRPC service
		}
	}()

	client.OnPacketRTPAny(func(media *description.Media, forma format.Format, pkt *rtp.Packet) {
		log.Printf("RTP packet from media %v\n", media)
		if media.Type == "audio" {
			audioHandler.HandlePacket(media, forma, pkt)
		}
	})

	// play
	_, err = client.Play(nil)
	if err != nil {
		panic(err)
	}

	//	panic(client.Wait())

	//cut out the audio with ffmpeg into wav file

	// send audio to parrot microservice via grpc

	defer ticker.Stop() // stop ticker when program ends.
}
