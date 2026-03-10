docker buildx build --platform linux/amd64,linux/arm64 \
  -t wordluc/sand-mmo-server:latest \
  -f server/Dockerfile . --push

docker buildx build --platform linux/amd64,linux/arm64 \
  -t wordluc/sand-mmo-web:latest \
  -f wasm_client/Dockerfile . --push
