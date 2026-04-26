# ==========================================
# Fase 1: Build
# ==========================================
FROM golang:1.26.2-alpine AS builder

# Impostiamo la directory di lavoro dentro il container
WORKDIR /app

# Copiamo i file dei moduli e scarichiamo le dipendenze
# (Usare github.com/DaniFX/...)
COPY go.mod go.sum ./
RUN go mod download

# Copiamo tutto il resto del codice sorgente
COPY . .

# Compiliamo l'eseguibile. 
# CGO_ENABLED=0 e GOOS=linux garantiscono una build statica perfetta per Alpine/Cloud Run
RUN CGO_ENABLED=0 GOOS=linux go build -o registry-service ./cmd/registry/main.go

# ==========================================
# Fase 2: Produzione (Immagine finale snella)
# ==========================================
FROM alpine:latest

# Certificati CA necessari per le chiamate HTTPS (es. verso Firestore o Firebase)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiamo SOLO l'eseguibile dalla fase di build
COPY --from=builder /app/registry-service .

# Cloud Run inietta la variabile d'ambiente PORT (di default 8080)
EXPOSE 8080

# Avviamo il microservizio
CMD ["./registry-service"]