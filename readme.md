simulator: when run, creates (currently) a single, fake video stream, looping an mp4 file over and over using mediamtx&ffmpeg

pelican: real-time streaming protocol (rtsp) server. ingests, currently, a video stream, then uses ffmpeg to turn stream into video. Then needs features added to send the mp3 to AWS S3, then S3 to AWS Transcribe, return the audio transcription, save to sqlite database, then use an ai to take large amounts of transcriptions and summarize this.

After built as a monolith, convert to grpc microservices.

run pelican first, opens a server to listen for streams.
