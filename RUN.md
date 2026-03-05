Tests:

Prerequirements: 

if you are using Rancher Desktop instead of Docker Desktop, make sure you have set the following environment variables:

```
export DOCKER_HOST=unix://$HOME/.rd/docker.sock
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
```

A similar pacth applies for Colima. Use `.colima` instead of `.rd`

Run tests:

```
go test ./...
```


