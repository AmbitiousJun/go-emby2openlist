# 第一阶段: 构建 web 产物包
FROM node:23-alpine AS web-builder

# 设置工作目录
WORKDIR /app/web/src

# 拷贝依赖声明 安装依赖
COPY web/src/package*.json ./
RUN npm ci

# 拷贝源码
COPY web/src .

# 执行构建
RUN npm run build

# 第二阶段：构建 Go 二进制文件
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置代理
RUN go env -w GOPROXY=https://goproxy.cn

# 复制源码
COPY cmd cmd
COPY main.go main.go
COPY go.mod go.mod
COPY go.sum go.sum
COPY internal internal
COPY web web

# 下载依赖
RUN go mod download

# 复制前端产物包
COPY --from=web-builder /app/web/src/build/client ./web/dist

# 编译源码成静态链接的二进制文件
RUN CGO_ENABLED=0 go build -tags=goexperiment.jsonv2 -a -installsuffix cgo -ldflags="-X main.ginMode=release" -o main .

# 第三阶段：运行阶段
FROM alpine:latest

# 设置时区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译后的二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8095
EXPOSE 8094

# 运行应用程序
CMD ["./main"]
