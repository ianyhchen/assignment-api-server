FROM golang:1.26.3 AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/task-api \
    ./cmd/api

FROM scratch

COPY --from=builder /out/task-api /task-api

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/task-api"]