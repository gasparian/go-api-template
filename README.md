# Go service template

This repository contains a generic template for Go services.

## Getting Started

### Prerequisites

- Docker  
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
