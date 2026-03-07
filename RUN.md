# Execution

## Prerequirements

The connection to the database is defined using environment variables. 

The server port is defined as an environment variable.

Make sure the following environment variables exists. Some will take a default if specified. 

```
DB_HOST="localhost"
DB_USERNAME="root"
DB_PASSWORD=""
DB_NAME=""
DB_PORT=3306
SERVER_PORT=8089
```

## Github Container Registry Credentials
To allow the cluster to download the container you'll need a Github Token from your personal account. 

Go to:
- `github.com -> settings -> developer settings -> personal access tokens -> tokens (classic)`
- Create a new token.Select a duration and grant the permissions: `read:packages`
- Click on `Generate`
- Copy the token `ghp_...` to a safe place.
- Login into ghcr.io: `docker login ghcr.io`. it will ask for your github username and the password (the token created).
- make sure your kubernetes context points to the application namespace
- If you are using a Mac, run the helper provided `bash extract-docker-ghcr.io-credentials.sh`

Alternatively, [follow instructions here](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/) to create credentials for a private registry. Use `docker login ghcr.io` instead of plain `docker login`


## Helm dependencies
If applying the chart manually, you need to resolve the dependencies. 
from the project root, run:
```
helm repo add kubelauncher https://kubelauncher.github.io/charts
helm dependency build chart/little-sample-cluster
```


## Too many files errors
When running on a localhost development cluster, and any of the containers crash and log errors related to inotify, not enough file descriptors... this oculd be due to the local system ahving low file limits. 

In macOS, follow: https://superuser.com/a/1679740

Alternatively, could be because the Docker engine doesn't have enough file descriptors. Follow [this documentation page](https://docs.rancherdesktop.io/how-to-guides/increasing-open-file-limit/) to increase file descriptor for Rancher. 

The error could be because of a combination of both.

# Tests

## Prerequirements 

if you are using Rancher Desktop instead of Docker Desktop, make sure you have set the following environment variables:

```
export DOCKER_HOST=unix://$HOME/.rd/docker.sock
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
```

A similar patch applies for Colima. Use `.colima` instead of `.rd`

## Run tests

```
go test ./... -v
```


