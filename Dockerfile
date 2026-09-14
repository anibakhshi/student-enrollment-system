FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/student-api \
    ./cmd/api


FROM scratch AS runtime

WORKDIR /app

COPY --from=builder /out/student-api /app/student-api

USER 65532:65532

EXPOSE 8081

ENTRYPOINT ["/app/student-api"]