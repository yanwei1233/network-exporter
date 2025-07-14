# 第一阶段：构建阶段
FROM harbor.shandy.com/nong360/golang:1.24.3-alpine3.21 AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum main.go ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /prometheus-wechat-webhook .
#-----------------------------------------------------------------------
# 第二阶段：运行阶段
FROM harbor.shandy.com/nong360/alpine:3.18
COPY --from=builder /prometheus-wechat-webhook /app/
WORKDIR /app
EXPOSE 5050
CMD ["/app/prometheus-wechat-webhook"]
