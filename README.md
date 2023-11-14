# Email to EPUB

Move top n emails to a temp folder, extract to eml files (done via this project), then convert to epub (done via [this project, which also shares the same name with this project*](https://github.com/gonejack/email-to-epub)).

*Because I don't know how many different ways you can say "email to epub."

## Prerequisites

```bash
go install github.com/gonejack/email-to-epub@latest
go install github.com/rclone/rclone@latest
```

## Build

```bash
go build -o main
```

## Usage

```bash
rclone config
make run
```

## Docker

```bash
docker build -t email-to-epub .
docker run --env-file .env -v "$HOME/.config/rclone:/root/.config/rclone" email-to-epub
```
