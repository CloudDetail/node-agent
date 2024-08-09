FROM golang:1.21.5-bullseye AS builder

WORKDIR /build
ENV GOPROXY https://goproxy.cn

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go build -v -o node-agent


FROM debian:bullseye-slim AS runner

WORKDIR /app
COPY --from=builder /build/node-agent /app

EXPOSE 9408
CMD ["/app/node-agent"]