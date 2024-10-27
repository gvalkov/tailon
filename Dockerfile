FROM golang:alpine as build

COPY . /go/src/github.com/gvalkov/tailon/
RUN apk add --no-cache --upgrade git
RUN <<EOF
set -xue
cd /go/src/github.com/gvalkov/tailon
go get 
go build -ldflags="-s -w"
EOF


FROM alpine:3.20

WORKDIR /tailon
COPY --from=build /go/src/github.com/gvalkov/tailon/tailon /usr/local/bin/tailon
CMD        ["--help"]
ENTRYPOINT ["/usr/local/bin/tailon"]
EXPOSE 8080
