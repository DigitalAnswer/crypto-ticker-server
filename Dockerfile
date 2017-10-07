# Base Image
FROM golang:latest

ENV PROJECT_NAME=crypto-ticker-server
ENV SRC_PROJECT=DigitalAnswer/${PROJECT_NAME}
ENV SRC_DIR=/go/src/github.com/${SRC_PROJECT}

WORKDIR /app
# Add the source code:
ADD . $SRC_DIR
# Build it:
RUN cd $SRC_DIR; go build -o $PROJECT_NAME; cp $PROJECT_NAME /app/
ENTRYPOINT ["./crypto-ticker-server"]