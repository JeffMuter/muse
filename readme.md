this project is still under development. while this service does work, it does require you to make an .env file in the /muse directory, you'll need to add a few .env files. I'll describe the process at the bottom of this readme.

this project is broken into a few microservices.

the overall goal of the project is to have multiple security cameras that use rtsp to feed video to a server.

Simulator: this is just a service for testing the actual project. When you run that golang service, it will effectively pretend to be 10 security cameras, sending an mp4 file on a loop using ffmpeg. I added a couple random video files for testing with some audio in it. Does nothing else, feel free to swap out the mp4 files with your own to test.

Pelican: this rtsp/ffmpeg using server is named pelican. the pelican service is responsible for ingesting rtsp video & or audio feeds, and converting them into mp3 files, just the audio. i chose ffmpeg for this. after this, pelican sends these files to aws S3 and repeats this for any number of feeds. i have tested a bit, and confirmed it can handle at least 10 video feeds at the same time.

Serverless: aws uses S3 to store files, then uses aws-transcribe to get a string of audio of anything said in the audio files. then sends these files to the next service, parrot.

Parrot: has the job of turning the json recieved from aws into just the strings we want. saving those into a local directory. Then sends what we would call a 'transcription' to an AI service called 'Anthropic' to summarize, and categorize each conversation on certain metrics. You can think of it like summarizing a long conversation into a single summary, then 'topics', and finally you can create a few categories I call 'alerts'. For example, if you add an alert for 'politics', the AI would add an alert to that array *only* if it detected a political conversation. You sort of have to use your own imagination for this. Add your name to it, then leave an rtsp device in a room, and if your name comes up, it'll be added as an 'alert'. 

Pigeon: is responsible for taking information from parrot, and notifying a human for manual review via email. Formatting emails using AWS SES(simple email service). This 'works', but be careful, gmail is a genuine pain to send/receive to. I happen to have some other emails on yahoo I use for this, to avoid the nonsense.

to run, go to the muse/ directory, and run 'docker compose up', this will activate all the services in the correct order.

Technology:

Golang     // best language for concurrency and writing servers
Docker     // for deploying microservice applications in a more automated way
SQLite     // not enough concurrent writes to justify using anything more intense
AWS        // cloud services i am most familiar with, but primarily used for the transcribe service
S3         // used to store audio / video / json output from transcribe
SQS        // simple queue service, saves parrot from checking S3, and basically 10% of the price.
SES        // simple email service. Sends email. Duh. Free until you start really pushing this project.
Lambda     // used to activate aws-transcribe in an automated way, and send json from it back to the parrot service
Transcribe // used to get audio transcriptions from audio files, converted into json format
FFMPEG     // for working with audio/video


you'll need to add env files to each of the following repos. and you'll need to populat the following variables:

pigeon .env:

MAIL_EMAIL=
GMAIL_PASSWORD=
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_SQS_QUEUE_URL=
ANTHROPIC_API_KEY=


parrot .env:

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_SQS_QUEUE_URL=
ANTHROPIC_API_KEY=


pelican .env:

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
