FROM golang:1.11.1-alpine3.8 as build
# All these steps will be cached
RUN mkdir /transport
WORKDIR /transport

RUN apk add --update --no-cache git

COPY go.mod .
COPY go.sum .

RUN go mod download

# COPY the source code as the last step
COPY main.go .

# Build the binary
RUN go build -o /go/bin/transport

# <- Second step to build minimal image
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/bin/transport /go/bin/transport
ENTRYPOINT ["/go/bin/transport"]
