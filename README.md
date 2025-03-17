# node-agent

A monitoring agent that exports network metrics for applications and their downstream dependencies, as well as application startup time metrics.

## Build

```bash
docker build -t node-agent:latest -f ./docker/Dockerfile .
```

## Deploy

```bash
docker run -d --rm \
  -e MY_NODE_NAME="xxxx" \
  -e MY_NODE_IP="192.168.1.xxx" \
  -v /proc:/proc:ro \
  --net=host --pid=host --privileged \
  node-agent:latest
```

## Configuration

Environment Variables:
- MY_NODE_NAME: Name of the node
- MY_NODE_IP: IP address of the node

For additional configuration options, please refer to the `config.yaml` file.