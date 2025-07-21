#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# 进入项目目录
cd "$SCRIPT_DIR" || exit 1

# 使用相对路径指定 Dockerfile，并以 k8s-network 目录作为上下文
docker build -t k8s-network-probe-server:v1.0 -f deploy/server.Dockerfile .
docker build -t k8s-network-probe-agent:v1.0 -f deploy/agent.Dockerfile .

# 推送镜像
#docker push xxx/k8s-network-probe-server:v1.0
#docker push xxx/k8s-network-probe-agent:v1.0
