# Go api template

This repository contains a generic template for Go services.

## Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install)  
- [Docker](https://docs.docker.com/engine/install/)  
- [delve](https://github.com/go-delve/delve) for debugging  

### Setup

1. Clone the repository.

2. Build application executable:  
    ```sh
    make build
    ```  
    You can check all available make targets by running `make help`.  

3. Run compiled app:  
    ```sh
    ./bin/server
    ```  
    or `make run` for development, without the need to build.  

5. Run some test queries:  
    ```sh
    curl -v localhost:8080/ping

    curl -v localhost:8080/version

    curl -v -X POST localhost:8080/api/v1/users \
        -H "Content-Type: application/json" \
        -H "Cookie: visitor_id=visitor-123" \
        -d '{
          "userId": "12345",
          "itemId": "item-001"
        }'

    curl -v -X OPTIONS -H "Origin: https://example.com" -H "Access-Control-Request-Method: POST" localhost:8080/api/v1/users
    ```  

For debug, use [this](./.vscode/launch.json) vs code config which runs `delve`.  

### Configuration

The application can be configured using environment variables, for example:  

- `RUNTIME_ENVIRONMENT`: defines where app is running: `dev`, `staging` or `production`.

### Deploy

The `deploy/` folder is a placeholder for Kubernetes deployment assets such as Helm charts and environment manifests.

### Docker

The Dockerfile uses a multi-stage build: the first stage compiles a static Linux binary, and the final stage is `scratch` to keep the image minimal. This reduces attack surface and image size by shipping only the compiled server and config files.

- Build image: `make docker.build`
- Run image: `make docker.run`
