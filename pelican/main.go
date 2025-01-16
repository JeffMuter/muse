package main

import (
	"context"
	"log"
	"os/exec"
	"sync"

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

	client.OnPacketRTPAny(func(media *description.Media, forma format.Format, pkt *rtp.Packet) {
		log.Printf("RTP packet from media %v\n", media)
	})

	_, err = client.Play(nil)
	if err != nil {
		panic(err)
	}

	panic(client.Wait())

	//cut out the audio with ffmpeg

	// send audio to parrot microservice via grpc

}
