FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o OrbittoAuth cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/OrbittoAuth .

COPY --from=builder /app/pkg/api/swagger.swagger.json ./pkg/api/swagger.swagger.json

EXPOSE 50051 8080

CMD ["./OrbittoAuth"]