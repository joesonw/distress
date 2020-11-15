FROM golang:1.15 AS build
WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN make build

FROM debian:10
COPY --from=build /go/src/app/distress-agent /.
CMD ["/distress-agent"]

