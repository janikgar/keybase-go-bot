FROM golang:1.18.5-alpine3.15 as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install github.com/goreleaser/goreleaser@latest && apk add git

COPY . .
RUN CGO_ENABLED=0 goreleaser build --snapshot --rm-dist --single-target && mv dist/*/keybase-go-bot /usr/src/app

FROM keybaseio/client:stable-slim
COPY --from=builder /usr/src/app/keybase-go-bot /home/keybase
WORKDIR /home/keybase

# USER keybase

CMD "/home/keybase/keybase-go-bot"