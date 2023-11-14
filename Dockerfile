FROM golang:1.20-alpine AS builder

# pre-reqs
RUN go install github.com/gonejack/email-to-epub@latest && \
    go install github.com/rclone/rclone@latest

# Download and install any required dependencies
WORKDIR /build
COPY . .
RUN go build -o main

#--------------------------------------------
FROM alpine:latest as package

WORKDIR /usr/local/bin

COPY --from=builder /go/bin/* .
COPY --from=builder /build/main .
RUN chmod +x *

CMD ["./main"]
