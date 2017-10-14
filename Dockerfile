FROM golang:latest as builder
WORKDIR /go/src/github.com/DigitalAnswer/crypto-ticker-server/
RUN go get -d -v golang.org/x/net/html
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/DigitalAnswer/crypto-ticker-server/app .
CMD ["./app"]