FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git make

# Copiar archivos de go
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Imagen final
FROM alpine:latest

WORKDIR /app

# Instalar certificados y runtime
RUN apk --no-cache add ca-certificates tzdata

# Copiar binario
COPY --from=builder /app/main .
COPY --from=builder /app/pkg /app/pkg

# Exponer puerto
EXPOSE 3000

# Ejecutar
CMD ["./main"]
