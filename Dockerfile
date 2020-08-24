FROM golang:alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git

WORKDIR /home/balanceServer

ADD go.mod go.sum ./
RUN go mod download

ADD . .
ENV PORT=9090

RUN go build -o balance .
CMD ["./balance"]

