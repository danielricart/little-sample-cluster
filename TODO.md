# little-sample-cluster

- Username should only be letters.

self-made decisions: 
- GET response body will be:
```json
{ "dateOfBrith": "YYYY-MM-DD" }
```
- If GET username is not found, response will be 404
- if PUT payload does not match, response will be 400
- username can be passed either uppercase or lowercase. or mixed. it's always normalized in DB to lowercase. 


## DONE
### AI Usage
to speed up boilerplate for e2e tests and multi-case tests.
to speed up http method filtering. 

### meta
- build action for GHA on merge-to-main and PR to Main
- also run tests on GHA

### application
- logging in JSON output for easier parsing / indexing. It penalizes raw-reading logs
- `SERVER_PORT` envVar. defaults to 8089. 
- endpoint /health. always responds `HTTP200 OK`
- e2e tests for /health
- expose GET /hello/<username>
  - YYYY-MM-DD must be a date before today's date.
  - response content: { “dateOfBrith”: “YYYY-MM-DD” }
  - response: 200 OK
- expose PUT /hello/<username> { “dateOfBrith”: “YYYY-MM-DD” }
  - Save or updates a given username and date of birth in a database
  - response 204 No Content

## TODO

### application
- pending store and fetch from DB. some tests fail because of this
- DB respository for a simple struct with username and dateOfBirth
- DB client
- DB settings as env

- expose prometheus metrics. total for inserted Date of birth, histogram with 12 buckets
- 
- helm chart for application

### infrastructure
in a production ready env this would need:
- sealed-seacrets to ensure secrets can be part of the repository
- alternative: external-secrets to fetch secrets from 3rd party storages
- cert-manager to issue TLS certificates automatically
- VPA for ensuring pods have enough memory to serve designated volume of requests
- HPA to ensure there are enpough pods to serve the growth of requests over time. and reduce fleet when unused.
- some sort of cluster autoscaler to ensure new pods can be scheduled. Karpenter if using AWS/Azure is recommended. Cluster-autoscaler otherwise.
- argocd eases the management of all the cluster applications and client applications
- dedicated namespace for the application
- dedicated namespace for hte database (this is an optional pattern)
- prometheus for metricsd
- cluster-metrics for extraction of resource usage 
- 