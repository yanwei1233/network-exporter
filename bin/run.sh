#!/bin/bash


SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd "$SCRIPT_DIR"

# 先创建namespace
NAMESPACE="k8s-network-probe"

# 检查命名空间是否存在
if kubectl get ns "$NAMESPACE" &>/dev/null; then
  echo "命名空间 '$NAMESPACE' 已存在，跳过创建"
else
  # 创建命名空间
  kubectl create ns "$NAMESPACE"
  if [ $? -eq 0 ]; then
    echo "成功创建命名空间 '$NAMESPACE'"
  else
    echo "创建命名空间 '$NAMESPACE' 失败" >&2
    exit 1
  fi
fi


#部署时需要注意nodeselector

#部署client
kubectl apply -f ${SCRIPT_DIR}/deploy/daemonset_gray.yaml

#部署server
kubectl apply -f ${SCRIPT_DIR}/deploy/deployment.yaml

#部署alertmanager
chmod +x ./alertmanager/run.sh
./run.sh