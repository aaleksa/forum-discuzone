FROM golang:1.23-bookworm

# Вимикаємо перевірку підписів (тимчасово)
RUN echo 'Acquire::AllowInsecureRepositories "true";' > /etc/apt/apt.conf.d/99disable-verification && \
    echo 'Acquire::AllowDowngradeToInsecureRepositories "true";' >> /etc/apt/apt.conf.d/99disable-verification && \
    echo 'APT::Get::AllowUnauthenticated "true";' >> /etc/apt/apt.conf.d/99disable-verification

# Оновлюємо списки пакетів, очищуємо кеш, встановлюємо потрібні пакети
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/* && \
    apt-get update && \
    apt-get install -y \
      libssl-dev \
      libsqlcipher-dev \
      pkg-config \
      gcc \
      build-essential \
      sqlcipher && \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-DSQLITE_HAS_CODEC -I/usr/include"
ENV CGO_LDFLAGS="-L/usr/lib -lsqlcipher -lcrypto"

# Command line local env start service command
#  CGO_ENABLED=1 CGO_CFLAGS="-DSQLITE_HAS_CODEC -I/usr/include" CGO_LDFLAGS="-L/usr/lib -lsqlcipher -lcrypto" go build -o forum-app . && chmod +x forum-app && ./forum-app
RUN go build -o forum-app . && chmod +x forum-app

EXPOSE 8080

CMD ["./forum-app"]
