FROM golang

RUN apt update && apt install -y pulseaudio alsa-utils libasound2-dev

WORKDIR /go/src/github.com/tuarrep/sounddrop
COPY . .

RUN go get

EXPOSE 19416

CMD ["go", "run", "main.go"]
