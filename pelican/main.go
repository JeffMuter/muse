package main

import (
	"context"
	"os/exec"
	"sync"

	"github.com/aler9/gortsplib/v2"
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
	err := client.Start(url, opts)

	//cut out the audio with ffmpeg

	// send audio to parrot microservice via grpc

	return
}
