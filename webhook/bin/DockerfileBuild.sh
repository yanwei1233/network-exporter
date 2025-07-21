#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd "$SCRIPT_DIR"
VERSION="v1.0"
DOCKERFILE_DIR="./hook_project"
IMAGE_NAME="wechat_webhook_image"

echo "正在构建 Docker 镜像..."
docker build -t ${IMAGE_NAME}:${VERSION} -f ${DOCKERFILE_DIR}/Dockerfile ${DOCKERFILE_DIR}

if [ $? -ne 0 ]; then
  echo "Docker 镜像构建失败，退出脚本。"
  exit 1
fi

echo "Docker 镜像构建成功！"

echo "password" | docker login --username username --password-stdin harbor.com  #这里要改为自己的harbor

docker push ${IMAGE_NAME}:${VERSION}

echo "镜像成功推送至harbor！"
