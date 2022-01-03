FROM golang:1.17 as build-env

WORKDIR /go/src/cave-canem
ADD . /go/src/cave-canem

RUN go get -d -v ./...

RUN go build -o /go/bin/cave-canem

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/cave-canem /

ENTRYPOINT ["/cave-canem"]