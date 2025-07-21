SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

cd "$SCRIPT_DIR"


echo "启动webhook程序"
docker-compose -f ../docker-compose.yaml up -d

if [ $? -ne 0 ]; then
  echo "webhook程序启动失败，退出脚本。"
  exit 1
fi