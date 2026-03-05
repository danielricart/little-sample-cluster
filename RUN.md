# Execution

## Prerequirements

The connection to the database is defined using environment variables. 

The server port is defined as an environment variable.

Make sure the following environment variables exists. Some will take a default if specified. 

```
"DB_HOST="localhost"
DB_USERNAME="root"
DB_PASSWORD=""
DB_NAME=""
DB_PORT=3306
SERVER_PORT=8089
```

# Tests

## Prerequirements 

if you are using Rancher Desktop instead of Docker Desktop, make sure you have set the following environment variables:

```
export DOCKER_HOST=unix://$HOME/.rd/docker.sock
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
```

A similar pacth applies for Colima. Use `.colima` instead of `.rd`

## Run tests

```
go test ./... -v
```


