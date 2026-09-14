# 阶段一：构建阶段
FROM golang:1.21-alpine AS builder

# 开启 CGO 禁用，构建完全静态的可执行文件（减小体积、提升启动速度）
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
# 如果国内拉取慢，可以设置代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 优先缓存依赖模块
COPY go.mod ./
# 如果有 go.sum 也需要 COPY go.sum ./
RUN go mod download

# 复制剩余代码
COPY . .

# 编译应用，去掉调试信息减小体积
RUN go build -ldflags="-s -w" -o main .

# 阶段二：运行阶段（使用极致精简的 alpine）
FROM alpine:latest

# 安装证书以支持 HTTPS 请求
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从 builder 阶段把编译好的二进制文件拿过来
COPY --from=builder /app/main .

# 暴露 80 端口（微信云托管规定）
EXPOSE 80

# 运行应用
CMD ["./main"]
