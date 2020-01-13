FROM golang:alpine as build
ADD . /go/src/github.com/gvalkov/tailon/
RUN apk add --upgrade git upx binutils
RUN cd /go/src/github.com/gvalkov/tailon && go get && go build && strip tailon && upx tailon

FROM alpine:3.7
WORKDIR /tailon
COPY --from=build /go/src/github.com/gvalkov/tailon/tailon /usr/local/bin/tailon

RUN unlink /usr/bin/awk && unlink /bin/grep && unlink /bin/sed
RUN apk add gawk grep sed

CMD        ["--help"]
ENTRYPOINT ["/usr/local/bin/tailon"]
EXPOSE 8080