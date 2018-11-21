FROM golang:1.10

RUN mkdir -p /go/src/enbase
WORKDIR /go/src/enbase
COPY . .
RUN curl -L -s https://github.com/golang/dep/releases/download/v0.3.1/dep-linux-amd64 -o $GOPATH/bin/dep
RUN chmod +x $GOPATH/bin/dep
RUN go build enbase

FROM scratch
COPY --from=0 /go/src/enbase /main
CMD ["/main/enbase"]