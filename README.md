# node-agent

node-agent 用于导出应用和下游依赖的网络指标和应用程序启动时间指标

## 构建

```bash
docker build -t node-agent:latest -f ./docker/Dockerfile .
```

## 部署

```bash
docker run -d --rm \
  -e PROCESS_TYPE="java,python,node" \
  -e PROCESS_TIME="true" \
  -v /proc:/proc:ro \
  --net=host --pid=host --privileged \
  node-agent:latest
```

## 配置环境变量

- PROCESS_TYPE: 监控的应用启动名称白名单，如java,python,node
- K8S_NAMESPACE_WHITELIST: k8s命名空间白名单, 如default,go-auto
- LRU_CACHE_SIZE 指标缓存大小，默认为50000
- PROCESS_TIME: 是否监控应用启动时间
- PID_SCAN: 更新进程pid信息的间隔时间，默认为1分钟
- PING_SCAN: 更新进程ping信息的间隔时间，默认为5s
- MY_NODE_NAME: 节点名称
- MY_NODE_IP: 节点IP
- FETCH_SOURCE_ADDR: 连接 metadata 获取 kubernetes 信息
- AUTH_TYPE && KUBE_CONFIG: 直接连接 kubernetes 获取信息