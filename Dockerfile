FROM golang:1.11

RUN apt update && apt install -y pulseaudio alsa-utils

RUN go get github.com/sirupsen/logrus
RUN go get github.com/thejerf/suture
RUN go get github.com/syncthing/syncthing/lib/rand
RUN go get github.com/shibukawa/configdir
RUN go get github.com/golang/protobuf/proto
RUN go get github.com/faiface/beep
RUN go get github.com/pkg/errors
RUN go get github.com/bkaradzic/go-lz4
RUN go get github.com/gogo/protobuf/gogoproto
RUN go get github.com/gogo/protobuf/proto
RUN go get github.com/minio/sha256-simd
RUN go get golang.org/x/text

RUN apt install -y libasound2-dev
RUN go get github.com/hajimehoshi/oto

RUN go get github.com/xtaci/kcp-go

WORKDIR /go/src/sounddrop
COPY . .

EXPOSE 19416

CMD ["go", "run", "main.go"]
