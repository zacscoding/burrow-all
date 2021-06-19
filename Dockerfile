FROM golang:1.14-alpine AS build

RUN mkdir -p /go/src/github.com/zacscoding/burrow-all ~/.ssh && \
    apk add --no-cache git openssh-client make gcc libc-dev
WORKDIR /go/src/github.com/zacscoding/burrow-all
COPY . .
RUN go build -a -o server

FROM alpine:3
COPY --from=build /go/src/github.com/zacscoding/burrow-all/server /bin/server
CMD /bin/server