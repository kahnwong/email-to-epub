FROM golang:1.20-alpine AS builder

# pre-reqs
RUN go install github.com/gonejack/email-to-epub@latest && \
    go install github.com/rclone/rclone@latest

# Download and install any required dependencies
WORKDIR /build
COPY . .
RUN go build -o main

# -----------------------
FROM alpine:latest

# install locales
ENV MUSL_LOCPATH="/usr/share/i18n/locales/musl"

RUN apk --no-cache add \
    musl-locales \
    musl-locales-lang

WORKDIR /opt/app

COPY --from=builder /go/bin/* /usr/local/bin/
COPY --from=builder /build/main .
RUN chmod +x main

CMD ["./main"]
