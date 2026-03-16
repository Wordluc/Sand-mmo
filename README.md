# Sand MMO

An MMO sand simulator in Go. Players share the same map in real time, placing sand, water, fire, lava and smoke that all interact with each other.

![Go](https://img.shields.io/badge/Go-1.26-blue?logo=go)
![Docker](https://img.shields.io/badge/Docker-Swarm-blue?logo=docker)
![Redis](https://img.shields.io/badge/Redis-Snapshot-red?logo=redis)
![WASM](https://img.shields.io/badge/WASM-Go-purple)

<img width="1184" height="594" alt="Screenshot 2026-03-16 at 22 05 37" src="https://github.com/user-attachments/assets/9ac78600-94ba-4669-b167-dbc6c99a6c00" />

www.wordluc.it:8080/


## Features

- **Custom binary protocol over WebSocket** — 64-bit packets for client→server, 16-bit per cell for server→client.
- **Chunk-based world simulation** — the world is divided into chunks simulated to improve performance.
- **Web client via Go WASM** — the same encode/decode logic is compiled to WebAssembly and reused in the browser client, with no duplication
- **Client-side queue of chunks** — always renders the latest state, dropping older updates to avoid time distortion
- **World snapshots with Redis** — the server can save and restore the full world state
- **Docker Compose in Swarm mode** — container orchestration as a lightweight alternative to Kubernetes

---

## Cell Types

| Cell | Behavior |
|------|----------|
| Sand | Falls down, piles up |
| Water | Flows sideways and down, extinguishes fire |
| Fire | Spreads to wood and leaves, turns to smoke |
| Lava | Flows like water, ignites wood and leaf, vaporizes water |
| Smoke | Rises upward, fades over time |
| Wood | Burns when touched by fire or lava |
| Leaf | Falls like sand, ignites easily |
| Stone | Static, indestructible |
| Vacuum | Destroys surrounding cells |

- **Protocol:** 64-bit packets for client→server commands (coordinate, cell type, brush type), 16-bit per cell for server→client chunk updates to minimize bandwidth.
- **Chunk-based world simulation:** the world is divided into chunks simulated to improve performace.
- **Chunk system:** the world is divided into `5×5` chunks. Only active chunks (and their neighbors) are simulated each tick. Modified chunks are broadcast to all connected clients over WebSocket.

## Running

```bash
docker compose up
```

---

## TODO

- [ ] Per-chunk parallel simulation using checkerboard pattern.
- [ ] Mobile support.
- [ ] Map navigation (pan the viewport across the world).
