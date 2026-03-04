# ai-core Daemon

A Go-based microservice for managing local LLM (Ollama) model loading/unloading.

## Overview

ai-core acts as a middleman between your applications and Ollama, providing:
- RAM management by controlling when models are loaded/unloaded
- REST API for chat interactions
- Health monitoring of Ollama connection

## Requirements

- Go 1.21+
- Ollama running on localhost:11434

## Building

```bash
go mod tidy
go build -o ai-core .
```

## Configuration

Edit `config.json`:

| Field | Description | Default |
|-------|-------------|---------|
| `ollama_host` | Ollama API address | localhost:11434 |
| `server_host` | API server binding | localhost |
| `server_port` | API server port | 8080 |
| `preferred_models` | List of preferred models | [] |
| `model_timeout_seconds` | Auto-unload timeout | 300 |

## API Endpoints

### GET /health
Returns health status and loaded models.

### POST /load-model?model=<name>
Loads a model into memory. If no model specified, uses first in preferred_models.

### POST /unload-model?model=<name>
Unloads a model from memory to free RAM.

### POST /chat
Chat endpoint (non-streaming). Requires loaded model.

```json
{
  "model": "llama2",
  "messages": [{"role": "user", "content": "Hello!"}]
}
```

### POST /chat/stream
Streaming chat endpoint. Same payload as /chat but with `stream: true`.

## Usage

```bash
./ai-core                    # Uses default config.json
./ai-core /path/to/config.json
```

## Systemd Service

Copy `ai-core.service` to `~/.config/systemd/user/` and run:

```bash
systemctl --user daemon-reload
systemctl --user enable ai-core
systemctl --user start ai-core
```

## Workflow

1. Start ai-core daemon
2. Call `/load-model?model=llama2` to load model into RAM
3. Use `/chat` or `/chat/stream` for inference
4. Call `/unload-model?model=llama2` when done to free RAM
