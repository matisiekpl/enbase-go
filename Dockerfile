FROM golang AS build

RUN mkdir -p /go/src/enbase
WORKDIR /go/src/enbase
COPY . .
RUN go get -v enbase
RUN go build -ldflags "-linkmode external -extldflags -static" enbase

FROM scratch
COPY --from=build /go/src/enbase/enbase /enbase
CMD ["/enbase"]