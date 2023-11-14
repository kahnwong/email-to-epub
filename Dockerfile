FROM golang:1.20-bookworm

# pre-reqs
RUN go install github.com/gonejack/email-to-epub@latest && \
    go install github.com/rclone/rclone@latest

# Download and install any required dependencies
WORKDIR /build
COPY . .
RUN go build -o main

WORKDIR /usr/local/bin

RUN cp /go/bin/* .
RUN cp /build/main .
RUN chmod +x *

CMD ["./main"]
