# Usa Ubuntu 24.04
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Update
RUN 

# Instalar pacotes
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
            python3 python3-pip python3-uno \
            golang-go \
            libreoffice-core libreoffice-writer libreoffice-common fonts-liberation fonts-dejavu-core &&\
    rm -rf /var/lib/apt/lists/*

# Instalar unoserver
RUN pip install --no-cache-dir --break-system-packages unoserver

ENV PATH="/usr/local/go/bin:${PATH}"

# Copiar código
WORKDIR /app
COPY main.go .

# Build do Go app
RUN go mod init writer-converter && \
    go mod tidy && \
    go build -ldflags="-s -w" -o /app/converter main.go

# Copiar entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]