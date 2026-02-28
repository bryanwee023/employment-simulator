# E2E Tests

> 💡 All commands listed below should be run from the monorepo root directory.

## Setting up

### On macOS

Docker on macOS runs inside a Linux VM, which cannot access Apple's Metal GPU.
Ollama runs natively to use Metal for hardware-accelerated inference.

If you don't have Ollama installed, install Ollama and pull the Llama 3.2:1b model

```bash
brew install ollama
ollama pull llama3.2:1b
```

Then, start the Ollama server and the test environment:

```bash
ollama serve
docker compose -f ./docker-compose.test.yml up --build -d
```

### On Linux

On Linux, Ollama runs as a container, so no manual activation is needed:

```bash
OLLAMA_URL=http://ollama:11434 \
docker compose -f ./docker-compose.test.yml --profile ollama up --build -d
```

## Running Tests

```bash
go test -tags=e2e -v -count=1 ./e2e_test/
```

To run a specific test

```bash
go test -tags=e2e -v ./e2e_test/ -run TestDoNothing
```

## Tearing Down

### On macOS

```bash
docker compose -f ./docker-compose.test.yml down
```

### On Linux

```bash
docker compose -f ./docker-compose.test.yml --profile ollama down
```