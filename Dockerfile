# ビルドステージ
FROM golang:1.22.5 AS builder

# パッケージインストール
RUN apt-get update && \
    apt-get install -y python3 python3-pip python3-venv && \
    rm -rf /var/lib/apt/lists/*

# Python仮想環境の作成とパッケージのインストール
RUN python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install PyMuPDF Pillow

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Goアプリのビルド
RUN go build -o main .

# 実行ステージ
FROM golang:1.22.5

# Python仮想環境をコピー
COPY --from=builder /opt/venv /opt/venv

# Goアプリをコピー
COPY --from=builder /app /app

# 環境変数を設定
ENV VIRTUAL_ENV=/opt/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"
ENV USER_ENV="user"
ENV PASSWORD_ENV="password"
ENV PORT_ENV=8080

WORKDIR /app

CMD ["/bin/sh", "-c", "./main -user $USER_ENV -password $PASSWORD_ENV -port $PORT_ENV"]
