# Usar una imagen de Go como imagen de construcción
FROM golang:1.22 AS builder

WORKDIR /app

# Copiar los archivos del módulo Go y descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto de los archivos del proyecto
COPY . .

# Construir la aplicación sin especificar el archivo principal
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp

# Usar Alpine para la imagen final debido a su tamaño reducido
FROM alpine:latest

# Agregar certificados CA y zona horaria (opcional)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar el ejecutable desde la etapa de construcción
COPY --from=builder /app/myapp .

# Exponer el puerto que utiliza la aplicación
EXPOSE 8080

# Ejecutar la aplicación
CMD ["./myapp"]