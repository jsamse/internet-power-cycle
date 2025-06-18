FROM golang:1.24 AS builder

WORKDIR /app

COPY src/ .

RUN go mod init internet-power-cycle
RUN go mod tidy

RUN go build -o internet-power-cycle .

FROM gcr.io/distroless/base

COPY --from=builder /app/internet-power-cycle /internet-power-cycle

CMD ["/internet-power-cycle"]