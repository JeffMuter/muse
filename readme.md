simulator: when run, creates (currently) a single, fake video stream, looping an mp4 file over and over using mediamtx&ffmpeg

pelican: stream ingestion server, cuts out video, sends via grpc to the next service... wip...

run pelican first, opens a server to listen for streams.
