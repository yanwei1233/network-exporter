FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./


RUN go env -w GOPROXY=https://goproxy.cn,direct ; go mod download
COPY . .

RUN CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build -o k8s-network-probe-agent cmd/agent/main.go

FROM alpine as runner
COPY --from=builder /app/k8s-network-probe-agent .
ENTRYPOINT [ "./k8s-network-probe-agent" ]

