FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct

COPY . .

WORKDIR /app/backend
RUN go build -o app_server ./main.go

FROM python:3.8-slim AS python-base

RUN echo "deb http://deb.debian.org/debian/ bookworm main" > /etc/apt/sources.list && \
    echo "deb http://security.debian.org/debian-security bookworm-security main" >> /etc/apt/sources.list && \
    echo "deb http://deb.debian.org/debian/ bookworm-updates main" >> /etc/apt/sources.list && \
    sed -i 's|http://deb.debian.org/debian|https://mirrors.aliyun.com/debian|g' /etc/apt/sources.list

RUN apt-get update && apt-get install -y \
    build-essential \
    gfortran \
    libopenblas-dev \
    liblapack-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/backend/app_server /app/app_server

COPY /codeBert_model /app/codeBert_model

WORKDIR /app/demo

COPY /demo /app/demo

RUN python3 -m venv /app/demo/venv \
    && /bin/sh -c ". /app/demo/venv/bin/activate && pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple && pip install --upgrade pip && pip --default-timeout=100 install --prefer-binary -r requirements.txt"

EXPOSE 50051

CMD ["/app/app_server","server"]
