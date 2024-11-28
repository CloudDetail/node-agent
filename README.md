# node-agent

node-agent 用于导出应用和下游依赖的网络指标和应用程序启动时间指标

## 构建

```bash
docker build -t node-agent:latest -f ./docker/Dockerfile .
```

## 部署

```bash
docker run -d --rm \
  -e MY_NODE_NAME="xxxx" \
  -e MY_NODE_IP="192.168.1.xxx" \
  -v /proc:/proc:ro \
  --net=host --pid=host --privileged \
  node-agent:latest
```

## 配置环境变量

- MY_NODE_NAME: 节点名称
- MY_NODE_IP: 节点IP

其他配置参数参考 `config.yaml`文件