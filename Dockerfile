FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/cmd/shortener/shortener cmd/shortener/main.go

FROM scratch
WORKDIR /app

COPY . .
COPY --from=builder /app/cmd/shortener/shortener cmd/shortener/shortener

CMD [ "./cmd/shortener/shortener" ]