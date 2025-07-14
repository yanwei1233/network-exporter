#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd "$SCRIPT_DIR"


kubectl -n monitoring create secret generic alertmanager-templates --from-file=vxTemplate.tmpl

# 检查命令是否成功
if [ $? -ne 0 ]; then
   echo "第一个命令失败，尝试运行第二个命令..."
   # 如果第一个命令失败，则运行第二个命令
   kubectl -n monitoring create secret generic alertmanager-templates --from-file=vxTemplate.tmpl -o yaml --dry-run=client | kubectl -n monitoring replace -f -
fi

kubectl patch alertmanager main -n monitoring --type='json' -p='[{"op": "replace", "path": "/spec/secrets", "value": ["alertmanager-templates"]}]'

if [ $? -ne 0 ]; then
  echo "k8s-main的secret打patch失败，退出脚本"
  exit 1
fi


kubectl -n monitoring create secret generic alertmanager-main --from-file=alertmanager.yaml -o yaml --dry-run=client | kubectl -n monitoring replace  -f  -

if [ $? -ne 0 ]; then
  echo "替换alertmanager文件失败，退出脚本。"
  exit 1
fi

echo "alertmanager文件部署完成！"