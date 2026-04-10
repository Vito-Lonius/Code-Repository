# 第一阶段：构建
FROM m.daocloud.io/docker.io/library/golang:1.23-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache git gcc musl-dev

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录为容器内的 /app，这里就是 go.mod 所在的位置
WORKDIR /app

# 1. 复制依赖文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 2. 复制整个 backend 文件夹的内容
COPY backend/ .

# 3. 编译命令：必须指向 cmd/server/main.go
# 注意：不需要写 backend/ 前缀，因为我们已经在 /app 下了
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/main.go


# 第二阶段：运行
FROM m.daocloud.io/docker.io/library/alpine:latest
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /root/

# 复制编译好的二进制文件
COPY --from=builder /app/server .

# 复制配置文件 (根据 main.go 逻辑，程序会找 /root/configs/)
COPY --from=builder /app/configs ./configs

ENV APP_ENV=docker
EXPOSE 8080

CMD ["./server"]