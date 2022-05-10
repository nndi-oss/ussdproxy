ARG IMAGE=golang:1.17-alpine

FROM golang:1.17-alpine AS builder
RUN apk add --no-cache git make
COPY . /go/src/github.com/nndi-oss/ussdproxy/
RUN cd /go/src/github.com/nndi-oss/ussdproxy/ && go build -o /dist/ussdproxy main.go

FROM $IMAGE
COPY --from=builder /dist/ussdproxy /bin/ussdproxy
RUN mkdir -p /ussdproxy/data
VOLUME /ussdproxy/data/
CMD ["/bin/ussdproxy", "--config", "/data/ussdproxy.yaml", "server"]
