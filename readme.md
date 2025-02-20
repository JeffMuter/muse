This project is currently under development, and is not ready for production. At the bottom of the docs, you can see how to run the current version, even in its development state.

this project is broken into a few microservices.

the overall goal of the project is to have multiple security cameras that use rtsp to feed video to a server.

this server is named pelican. the pelican service is responsible for ingesting rtsp video & or audio feeds, and converting them into mp3 files, just the audio. i chose ffmpeg for this. after this, pelican sends these files to aws.

aws uses S3 to store files, then uses aws-transcribe to get a string of audio of anything said in the audio files. then sends these files to the next service, parrot.

parrot has the job of turning the json recieved from aws into just the strings we want. saving those into a database, then sends what we would call a 'conversation' to an AI to summarize, and categorize each conversation on certain metrics. Saving all of this to the database, and conditionally sending formatted information to the final service, pidgeon

pidgeon is responsible for taking information from parrot, and notifying a human for manual review. Formatting emails, and/or texts to have a human manually review transcriptions, and any files directly.

one other service exists for testing purposes, called simulator / needs changed to killdeer. this service is designed to mimic the bahavior of recieving multiple video streams, although currently, it only does one at a time until the mvp is completed. currently accomplishing this by feeding an mp4 file to the rtsp server.

to run, go to the muse/ directory, and run 'docker compose up', this will activate all the services in the correct order.

Technology:

Golang     // best language for concurrency and writing servers
Docker     // for deploying microservice applications in a more automated way
SQLite     // not enough concurrent writes to justify using anything more intense
AWS        // cloud services i am most familiar with, but primarily used for the transcribe service
S3         // used to store audio / video / json output from transcribe
Lambda     // used to activate aws-transcribe in an automated way, and send json from it back to the parrot service
Transcribe // used to get audio transcriptions from audio files, converted into json format
FFMPEG     // for working with audio/video
