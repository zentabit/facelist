FROM golang:1.14 as builder

WORKDIR /app

COPY go.* /app/
RUN go mod download

COPY *.go /app/
RUN CGO_ENABLED=0 go build -o facelist *.go

FROM alpine:latest
COPY --from=builder /app/facelist /facelist
COPY templates/ /templates/
CMD ["/facelist"]