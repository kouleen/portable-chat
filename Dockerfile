FROM kouleen/golang:1.25 AS builder
LABEL authors="kouleen.china@gmail.com"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o portable-chat ./cmd/api/main.go

FROM kouleen/alpine:latest
LABEL authors="kouleen.china@gmail.com"
RUN apk add --no-cache tzdata && rm -rf /var/cache/apk/*
WORKDIR /app
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 TZ=Asia/Shanghai
COPY --from=builder /app/portable-chat .
EXPOSE 9191
CMD ["./portable-chat"]