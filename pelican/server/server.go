package server

import (
	"fmt"
	"log"
	"os/exec"
	"pelican/cloud"
	"strings"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
)

type ServerHandler struct {
	S         *gortsplib.Server
	mutex     sync.RWMutex
	streams   map[string]*streamState
	StreamMgr *StreamManager
}

type streamState struct {
	stream    *gortsplib.ServerStream
	publisher *gortsplib.ServerSession
	// Map to track active RTP handlers for each session
	activeHandlers map[*gortsplib.ServerSession]chan struct{}
}

func NewServerHandler(streamMgr *StreamManager) *ServerHandler {
	return &ServerHandler{
		streams:   make(map[string]*streamState),
		StreamMgr: streamMgr,
	}
}

// called when a connection is opened.
func (sh *ServerHandler) OnConnOpen(ctx *gortsplib.ServerHandlerOnConnOpenCtx) {
	// log.Printf("conn opened")
}

// called when a connection is closed.
func (sh *ServerHandler) OnConnClose(ctx *gortsplib.ServerHandlerOnConnCloseCtx) {
	log.Printf("conn closed (%v)", ctx.Error)
}

// called when a session is opened.
func (sh *ServerHandler) OnSessionOpen(ctx *gortsplib.ServerHandlerOnSessionOpenCtx) {
	// log.Printf("session opened")
}

// called when a session is closed.
func (sh *ServerHandler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	//	log.Printf("session closed")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// Check if this session is a publisher for any stream
	for path, state := range sh.streams {
		if state.publisher == ctx.Session {
			log.Printf("Publisher session closed for path: %s", path)
			state.publisher = nil
			break
		}

		// Cancel any packet handlers for this session
		if cancel, ok := state.activeHandlers[ctx.Session]; ok {
			close(cancel)
			delete(state.activeHandlers, ctx.Session)
			log.Printf("Packet handler removed for session on path: %s", path)
		}
	}
}

// called when receiving a DESCRIBE request.
func (sh *ServerHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	//	log.Printf("Received DESCRIBE request for path: %s", ctx.Path)

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// Check if we have a stream for this path
	if state, ok := sh.streams[ctx.Path]; ok {
		log.Printf("OnDescribe: Stream found for path: %s", ctx.Path)
		return &base.Response{
			StatusCode: base.StatusOK,
		}, state.stream, nil
	}

	log.Printf("OnDescribe: No stream available for path: %s", ctx.Path)
	return &base.Response{
		StatusCode: base.StatusNotFound,
	}, nil, nil
}

// called when receiving an ANNOUNCE request.
func (sh *ServerHandler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("Stream announced on path: %s", ctx.Path)

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// If we already have a stream for this path
	if state, ok := sh.streams[ctx.Path]; ok {
		log.Printf("Replacing existing publisher for path: %s", ctx.Path)

		// If there's an old publisher, cancel its packet handlers
		if state.publisher != nil && state.publisher != ctx.Session {
			if cancel, ok := state.activeHandlers[state.publisher]; ok {
				close(cancel)
				delete(state.activeHandlers, state.publisher)
			}
		}

		// Update the publisher
		state.publisher = ctx.Session

		// Create a new stream with the new description
		// We need to keep track of the old stream to close it after creating a new one
		oldStream := state.stream
		state.stream = gortsplib.NewServerStream(sh.S, ctx.Description)

		// Close the old stream after creating the new one
		if oldStream != nil {
			oldStream.Close()
		}
	} else {
		// Create a new server stream
		stream := gortsplib.NewServerStream(sh.S, ctx.Description)

		// Store the new stream state
		sh.streams[ctx.Path] = &streamState{
			stream:         stream,
			publisher:      ctx.Session,
			activeHandlers: make(map[*gortsplib.ServerSession]chan struct{}),
		}

		// Start processing this stream path
		sh.StreamMgr.AddStream(ctx.Path)
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a SETUP request.
func (sh *ServerHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	// This is a client trying to read or publish to the stream
	//	log.Printf("Received SETUP request for path: %s", ctx.Path)

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// If this is a read request, return the stream if it exists
	state, ok := sh.streams[ctx.Path]
	if !ok {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, fmt.Errorf("no stream available at path %s", ctx.Path)
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, state.stream, nil
}

// called when receiving a PLAY request.
func (sh *ServerHandler) OnPlay(ctx *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("Client started playing stream: %s", ctx.Path)
	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a RECORD request.
func (sh *ServerHandler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	log.Printf("Client started recording to path: %s", ctx.Path)

	sh.mutex.Lock()

	// Ensure we have a stream for this path
	var state *streamState
	var ok bool

	if state, ok = sh.streams[ctx.Path]; !ok {
		// Create a new stream if it doesn't exist
		stream := gortsplib.NewServerStream(sh.S, ctx.Conn.Session().AnnouncedDescription())
		state = &streamState{
			stream:         stream,
			publisher:      ctx.Session,
			activeHandlers: make(map[*gortsplib.ServerSession]chan struct{}),
		}
		sh.streams[ctx.Path] = state
	} else {
		// Update the publisher of the existing stream

		// If there's an old publisher, cancel its packet handlers
		if state.publisher != nil && state.publisher != ctx.Session {
			if cancel, ok := state.activeHandlers[state.publisher]; ok {
				close(cancel)
				delete(state.activeHandlers, state.publisher)
			}
		}

		state.publisher = ctx.Session
	}

	// Create a cancel channel for this session's packet handlers
	cancelChan := make(chan struct{})
	state.activeHandlers[ctx.Session] = cancelChan

	// Store a local copy of stream for use in callback
	currentStream := state.stream
	sh.mutex.Unlock()

	// Called when receiving a RTP packet
	ctx.Session.OnPacketRTPAny(func(medi *description.Media, forma format.Format, pkt *rtp.Packet) {
		select {
		case <-cancelChan:
			// This handler has been cancelled
			return
		default:
			// Use the stored stream reference instead of accessing it through the map
			if currentStream != nil {
				currentStream.WritePacketRTP(medi, pkt)
			}
		}
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// begin setup for the 'streams'

type StreamManager struct {
	mu      sync.Mutex
	streams map[string]bool
	wg      sync.WaitGroup
	cancel  map[string]chan bool
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]bool),
		cancel:  make(map[string]chan bool),
	}
}

func (sm *StreamManager) AddStream(path string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[path]; exists {
		log.Printf("stream is already being processed on path: %s\n", path)
		return
	}

	fmt.Println("stream being added...")

	sm.streams[path] = true
	sm.cancel[path] = make(chan bool)

	sm.wg.Add(1)
	go sm.processStream(path, sm.cancel[path])
}

func (sm *StreamManager) RemoveStream(path string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if cancelChan, exists := sm.cancel[path]; exists {
		close(cancelChan)
		delete(sm.cancel, path)
		delete(sm.streams, path)
		fmt.Printf("stream successfully removed: %s\n", path)
	}
}

func (sm *StreamManager) processStream(path string, cancelChan chan bool) {
	defer sm.wg.Done()

	cleanPath := strings.ReplaceAll(path, "/", "-")
	if cleanPath == "" || cleanPath == "-" {
		cleanPath = "default"
	}
	if cleanPath[0] == '-' {
		cleanPath = cleanPath[1:]
	}

	for {
		select {
		case <-cancelChan:
			fmt.Printf("stopping process for stream: %s\n", path)
			return
		default:
			timestamp := time.Now().Format("20060102-150405")
			outputFileName := fmt.Sprintf("output-%s-%s.mp3", cleanPath, timestamp)

			// Build the RTSP URL
			rtspURL := fmt.Sprintf("rtsp://localhost:8554%s", path)

			log.Printf("Recording from %s to %s", rtspURL, outputFileName)

			// decided to make 600 seconds so that we get a whole conversation. But to try
			// and not have the server be abused in case I fall asleep for testing, or
			// something like that. Don't  judge me.
			cmd := exec.Command("ffmpeg",
				"-i", rtspURL,
				"-t", "10", // 600sec stream then restart
				"-vn", // cut out video, only audio added to file
				"-acodec", "libmp3lame",
				"-ab", "128k", // Audio bitrate explicitly set
				"-y",           // Overwrite output file if it exists
				outputFileName, // output file
			)

			if err := cmd.Run(); err != nil {
				fmt.Printf("Error recording from %s: %v, retrying in 5 seconds...", rtspURL, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Upload in a separate goroutine so it doesn't block the next recording
			go func(outputFileName string) {
				err := cloud.UploadFileToS3(outputFileName)
				if err != nil {
					fmt.Printf("Error uploading %s to S3: %v", outputFileName, err)
				} else {
					fmt.Printf("Successfully uploaded %s to S3", outputFileName)
				}
			}(outputFileName)
		}
	}
}
