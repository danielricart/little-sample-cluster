# Dependencies
The application relies on a Database service. This service is provided in-cluster at the time of installing the helm chart. 

For a proper production setup this is probably not ideal, as the default MySQL settings won't make the cut under load. And the configuration of the dependency pollutes the values.yaml of the main application.

For a proper setup, it's best that the DB service is already provided, either by an independent application, or an external service. Alternative option could be to use AWS Controller for Kubernetes and provision and RDS database object. More on this is discussed in [`PRODUCTION.md`](./PRODUCTION.md)

# Data model
The datamodel of the application is clearly a key-value, using the username as the unique key of the data pair. With no further details on the usage of the service or additional information, I didn't envision the need of any sort of unique autoincremental ID for this usecase.

Usernames are considered case-insensitive, `oneuser` and `OneUser` are the same registry. To change this behaviour, the DB column must be created with a case-sensitive collation: `username VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_bin primary key` instead of the current one. 

## Database engine

I chose Mysql as i am familiar with it and I could craft a quick client in golang easily. For a simple application any engine (SQL or "noSQL") could work. The important details is that the team operating it is familiar enough with the quirks of operating the selected technology in production at a given volume. 

### Alternatives
Other similar approaches equally stable, and well-known in the SQL space could be MariaDB, postgreSQL or AWS Aurora. 

For a high performance pure key-value application we could move to Redis/ValKey with persistence enabled (both AOF Append-Only File and RDB) to ensure persistence with no dataloss (similar reliability to postgreSQL, according to their own docs). Without the highest degree of persistence we risk losing data in case of a failure of the database.

All the mentioned DB Engines have their corresponding cloud-managed services. 

If we are working with full Cloud services, we could change the repository to use AWS DynamoDB (if we have access to AWS), or its equivalent in other cloud providers. 

There are other DB Engines that we could use. 

# Testing
Given that this service has very little logic and is heavy on data retrieval and storing, I went with some openbox end to end approach. The database is integrated as part of the test suite using `testcontainers`. The sample data is also populated as part of the setup stage of the tests. 

The tests are heavily manually built as they rely very little on external steps or the same code it's trying to test.

# metrics
This application exposes two application metrics.
- "birthday_registered_valid_total": "Total number of valid birthdays registered"
- "birthday_invalid_total": "Total number of invalid birthdays attempted to register"

These counters are incremented in PUT actions. `invalid` attempts is anything that returns not an HTTP204. 

It's not a part of the assignment but a nice addition to expose also a podMonitor for monitoring this critical application.

In addition to the application metrics, it exposes the go engine metrics, with internal details about how the application is performing. Here, could be interesting keeping an eye on the Garbage Collection activity. If it's too frequent, the application may see a performance boost by adding more memory as a quick fix. 

# Helm chart
The helm chart includes the dependency of  the mysql database. I chose a simple upstream chart for it instead of going with the MySQL operator or any other fully fledge kubernetes-native approach to it. This decision helps scoping better the task. Also, it ensure it can be run in a `kind` cluster or similar with no additional requirements.

The chart includes support for: 
- Ingress API (not gateway). For a test application, Ingress CRDs are included as part of the core K8S API but Gateway requires installing the Gateway API CRDs  `kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.0/standard-install.yaml`
- Horizontal Pod Autoscaling: despite being disabled in the values file, it allows adding additional stateless replicas of the application when the average resource consumption (CPU, memory) of all the replicas crosses a given threshold. Finding the threshold may require some manual testing and observation. 
- PodMonitor: will instruct prometheus operator to start scraping all the metrics exposed. 
- Serviceaccount: This application does not require any specific access to any API. by using a dedicated Account we prevent granting to the binary any permissions that does not need to have. If the SA needs spcific access levels, it requires a proper RBAC management with some (Cluster)Roles and appropiate Bindings. 

