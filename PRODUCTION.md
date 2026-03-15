# Going to production

---

a proper architecture for this application, when having AWS as a main infrastructure provider, requires some planning and decisions. 

For this application we will go with some managed services where they make sense while still keeping operational flexibility and ownership. 

# runtime
For the runtime, either ECS or EKS could work as container runtime for this specific application. 

ECS is a higher level abstraction that reduces the operational complexity of maintaining a container orchestration infrastructure. From a user's perspective you only see either just your workloads (Fargate serverless) or the ec2 instances (plain ECS) and everything else is orchestrated and managed. One of its biggest cons is the close tight of the observability tooling and debugging to AWS services (Cloudwatch) and AWS Support for infrastructure debugging. Also, it requires either custom tooling or AWS-provided integrations (i.e: there's no argocd or Flux for ECS). Most of the services required (certificates, DNS management, ingress, load balancing...) must use AWS services.

EKS rises the abstraction compared to self-deploying your own Kubernetes infrastructure, but it is below ECS. It still offers some managed integrations, and it's somewhat coupled to foundational AWS Services. It offers optionality on what of the offered services you want AWS to manage for you or if you prefer to operate them yourself. The opreational load is higher than when using ECS, or fargate, but can still manageable once deployed and properly dimensioned. AWS will take care of dimensioning the API and ETCD servers according to the load detected and cluster size. There's still some dependencies on AWS Cloudwatch but limited to API logs and some infrastructure metrics. 

Given the stability requirements and the operational work that comes with it, familiarity with the technology and operational flexibility that offers having access to a fully fledged K8S, I'll pick Amazon EKS as the runtime.

It's true that, in a production environment with a single high traffic application exactly like this exercise, picking ECS would be the most optimal solution even accounting for the added reliance on many AWS services and the lack of industry-standard tooling.

In terms of worker nodes, we only need one or two nodegroup-managed instances for bootstrapping the cluster. Cluster autoscaler activities can be managed and defined at runtime using Karpenter and its concept of nodepools. This autoscaler component replaces or can operate independently of upstream's CNCF cluster-autoscaler. It's aware of the instance definitions, equivalences and prices of AWS and Azure instances, and can pick the most suited instance to fit the amount of pending pods plus the estimated Daemonset overhead. 

For Karpenter to work it needs an instance where it's running before it takes over the instance provisioning activities. As we also need an instance to run a starter coreDNS, VPC CNI and kube-proxy, it generates little waste. we can use plain eks nodegroups for this.

There are few other controllers we need to ensure the cluster can be bootstrapped and operated. Many of them can be provided by AWS at installation time and their lifecycle is managed using AWS API/Terraform Module/AWS Console. Only a small subset is convenient enough to be managed and the rest offer more freedom when they are self-managed instead. Moving from managed addons to self-provided deployments and viceversa is not an easy two-way door. 

## AWS addons
To fully work with EKS and interacting we need to ensure we have some operators that help interacting with cloud-provider-spacific resources. 
- AWS Pod Identities/IRSA: to inject and expose AWS credentials to applications and controllers (required for ACK, EBS, AWS LB and others).
- AWS Controller for K8s (ACK): with some additional modules like IAM-controller, RDS-controller.
- AWS VPC CNI to interact with the network. we could use others, but this works well in our context.
- coreDNS
- kubeproxy

## other required controllers
- AWS Load balancer controller: to provision and treat AWS LBs as k8s objects.
- AWS EBS CSI driver: to offer persistent volumes to stateful applications in the cluster (if any, can be managed externally)

## Cost

Assumptions: 1 AWS month = 730 hours.

The costs to consider here are:
- EKS Control plane (0,10$/h)
- Bootstrap Nodegroup intances. (2x m8g.xlarge, 0,201$/h)
- Bootrap nodegroups EBS volumes. (2x 64GB GP3, 0,088$/GB-month)

Total costs of an empty cluster:
- Control plane: 73$
- Nodegroup EC2: 294$
- nodegroup EBS: 11$

*Total: 378$/month* 

This cost does not consider network traffic of any kind, AWS Load Balancers, external databases, nor the cost of the additional EC2 instances that Karpenter will provision to schedule all the pods that don't fit in the bootstrapping instances. 


# Database: 
- Amazon RDS
Given that the RDS offering is solid, and supports MySQL LTS versions, it's best if we go with the offering instead of hosting our own. This will free some cognitive load from the team and reduces the risk of the cluster operations as the state is hosted outside of it. 

This decision doesn't come for free. When setting up the DB Instance we will need to specify, at least, our own `Parameter Groups`. This will ease in the future customizing settings in the RDS database without rebuilding. 

We will need to evaluate replication, volume performance, costs,...  

Also, the performance of the database still needs to be monitored, by using prometheus-mysql-exporter (or any appropiate metrics extraction toool suited for our flavour of RDS).

In terms of resources and costs, we can go with db.m8g.large (2cores, 8gb) for ~122$/month or even lower to a db.t4g.medium (1c, 4gb) for ~50$/month.

These are safe starting points to ensure some stable performance. With the actual application running we could adjust better the instance type. 

The current helm chart of the application would need to be modified to remove the dependency on mysql chart. If we want to keep the approach of linking the application and ints dependencies,  we could replace the current dependency with a new in-chart `DBInstance.rds.services.k8s.aws/v1alpha1` object to provision and maintain the RDS instance as a kubernetes object. This approach requires the Amazon Controller K8S (ACK) installed with the RDS controller enabled. 

Alternate implementations could rely on AWS Aurora MySQL Serverless or OnDemand, but the starting pricing is similar or higher than classical RDS.

A different approach by only providing a secret could be taken if there's a centrally-provided big multitenant Database instance, with a team operating it as a service offering.

*Total cost: 122$/month* 

# infrastructure deployment
A good approach to managing the infrastructure would be using terragrunt for templating the environments and all variables terraform requires. Even if we have (for now) one single environment/cluster/account.

then, terraform takes the lead of provisioning the infrastructure based on the terragrunt run. 

Here, I'll separate what's an account or regional resource that's shared among every other component like, DNS Zones (generally, could be different), VPC and network foundations, Peering/TransitGateway management,... and any other kind of very shared resource. Some IAM roles may fall in here as well. 

Then, the cluster-related Terraform modules. In here we define all the required resources each cluster requires like Security groups, Instance roles, EKS cluster, eks addons, bootstrap nodegroups, secrets, certificates... It's also the moment to install the required bootstrapping components taht are self-managed like argocd and container registry credentials using the Kubernetes Provider's kubernetes_manifest. 

Also, it's time to apply the argocd applications for the critical components. These can be done using the same TF modules or externally after the TF section is done. 

## Cluster application deployment
Application deployment, can be deployed and maintained using different ways. One that's becoming a best practice and common in the industry is folowing a Gitops model. 

In a Gitops model, the deployment is initiated when a change is applied in a monitored repository. The detection can be pushed to the cluster externally and apply the manifests, or use dedicated tooling from inside the cluster like Flux or ArgoCD. 

Any can work, I'll pick ArgoCD. Eventho is fully CRD and CLI based, having a UI causes no harm and allows for checking at a quick glance what's the state of the applications.

As ArgoCD is one of the first applications provisioned and properly configured from its helm chart, it can be used to provision ad maintain all the components required. even before having ingress, DNS management, TLS Certificates... 

all the initial Argo applications objects will be provisioned using an app-of-apps argocd pattern. 

# application deployment 
Our little application is delivered as a helm chart and a container stored in a container registry. There's an argocd application available in the `application` folder. 

In a production environment, the argocd application would be updated either directly in the cluster by our CD system using valid system credentials or using an intermediate deployment tooling for advanced logic in the deployment.  

The ArgoCD Application YAML is available in the `application` folder. 

# Observability:
In here we could use datadog, prometheus stack, or other services. the most common approach uses prometheus and grafana at least in part of the flow. Datadog can be added instead of, in addition to, or on top of prometheus stack. 

## metrics
- Prometheus
Prometheus-style metrics is the de facto standard for kubernetes. It provides objects for metrics-endpoint discovery. 
The provided chart allows enabling the Monitor endpoints by setting `little_sample_cluster.podMonitor.enabled: true` and `mysql.metrics.enabled: true`

To manage prometheus instances we will install kube-prometheus-stack, which will aggregate several required exporters and components under the same installation. 

With such exporter we can provision at once several monitoring tools:
- node exporter, for gathering low-level metrics about each instance in the fleet. 
- several kubernetes core components (apiserver, kubeproxy, coredns...)
- grafana, for visualization

## logs
Logs are, in this context, highly operational. We are generating ingress logs, application logs, and cluster logs (api and audit logs, if enabled)

API and audit logs are available in CLoudwatch logs. we will need an exporter based on lambda SQS and SNS subscriptions to extract them. 

For pod-generated logs, fluentbit or fluentd are the defacto extractors. 

Once collected logs must be forwarded somewhere. It can be a 3rd party, an elasticsearch database or any log service like loki, graylog, or whatever is available.

## visualization
Grafana is one of the popular visualization tool that's used for observability data. It can be provisioned automatically along with Prometheus using kube-prometheus-stack. 

we can create Dashboards, keeping them in sync with a repository to ensure full traceability. Some of the components provision default dashboards.

## alerting
Alertmanager integrates with prometheus to trigger alerts and call an endpoint. Other services offer their own alerting system (datadog, grafanacloud). 

Routing alerts must be configured to post a wbhook call to the alerting system of choice (can be Slack, grafanacloud alerting, Pagerduty, or any alert router service available).

# autoscaling

## Vertical Autoscaling
**Kubernetes vertical pod autoscaler** is an operator that monitors resource consumption on pods and adds more memory or CPU resources when the usage is close to the threshold or the container dies due to resource exhaustion errors. 

Kubernetes VPA operator closely monitors the resource consumption trends and raises the required resources by a given % before it crashes on its own. 

Since kubernetes 1.35, In-place updates meaning container resources can be added without forcing a pod replacement if possible, graduated to Stable. This is a great improvement for pods which state is costly to rebuild or quorum to stabilize (think about kafka workers, large database clusters).

In a production cluster, all core components that are not scaling horizontally based on CPU/memory, should have an associated VPA to ensure they can deal the expected amount of load per replica. 

## horizontal Autoscaling
Horizontal autoscaling is an embedded kubernetes capability to add more replicas of a deployment based on k8s metrics or proportional to the cluster size.
Most common usages for this is to add more replicas based on the memory consumption or CPU usage.

If we prefer to scale based on more application-specific metrics (like amount of requests, kafka lag, GPU queue size) we can use KEDA as the autoscaler, which can size deployments based on metrics coming from multiple sources (prometheus, kafka, dynamodb...) including kubernetes own resource metrics. 

When using any sort of horizontal autoscaling it's best not to also scale vertically on the same metric, as this will create competing conditions as the averages can change when the number of replicas change.

It's perfectly ok to keep both autoscaling dimensions to ensure that each individual pod can cope with the load it has been dimensioned (thinking aobut ingress pods with enough CPU/Mem resources to process 1000 rps. )

## cluster autoscaling
To make sure that we can fit all the pods in the cluster, we need the ability to add more nodes as demand grows. This is also true the other way around, when the nodes are becoming unused because workloads running do not require all the capacity of the clusters. 

For this, two big projects are available: 
- cluster-autoscaler
- karpenter

### cluster autoscaler
This is a component that adds/removes nodes to the cluster. It monitors the amount of unused capacity of each nodegroup of the cluster and adds/retires nodes from each nodegroup. It interacts with multiple infrastructure providers to add and remove nodes following each providers' implementation. 

When using AWS, it's aware of the limitations and capabilities of autoscaling groups. One of those limitations is that you need to provide the autoscaling group with a static list of "equivalent" instance types; the ASG will fall back to any of those in case of capacity problems of the main instance type. 

This, in the long run, is problematic, as you cannot optimize your fleet for costs, or cpu/memory ratio. Also, if you need specialized instance types, it requires maintaining new nodegroups.

At the time of a cluster upgrade, or when adding finetunes to the AWS Autoscaling groups, that will trigger a full instance replacement of all the instances in the impacted nodegroups. This happens at deployment time and its difficult to pace and to halt. If errors occur, whose chances increase with the size, it requires involving AWS Support.

When using AWS, for very small fleets, or very specific nodegroups, like the defined bootstrap nodegroup above, it's good enough, tho. 

### Karpenter 
Karpenter is a more modern take on cluster autoscaling. It's very tailored to AWS and Azure clouds. It focus on maintaining individual nodes instead of the provided autoscaling groups of each provider. This give lots of flexibility on the variety of instances it can select for a given workload.

It's price-aware, label-based, and it monitors the amount of pods pending in the cluster. It's concept is still nodepools, where we define the specific selection criteria we want based on the intended usage and freedom we want to give ot the selection process. 

Opossed to cluster-autoscaler, karpenter replaces nodes at run time, hence a change in the nodepool definition that would trigger a replacement of the nodes, will not block the deployment and it will be handled asyncronously.

In general, if using compatible cloud providers, *Karpenter* is a better choice for cluster autoscaling operations. 


# namespace management
Namespaces need to be managed in some way. The bare minimum is that namespaces are not created automatically (like argoCD's `sync-option.createNamespace: true` or Helm's `--create-namespace`) and provisioned to the cluster using the deployment pipeline or a gitops repository. 

If there's additional objects that need to be applied to each namespace this approach helps with ensuring that the required objects are present (like netpols, resourcelimits, service accounts, namespace annotations...) and that in case of change, it's tracked where to apply the changes. 

# isolation and security
namespace isolation, network policy,

By default, k8s does not block communication between services in different namespaces. Also, it does not enforce any kind of limit on the amount of resources any pod can request or set any default on this. 

Also, it doesn't impose any permission restriction to what command any user can run anywhere in the cluster. Any kind of user management for kubernetes must be implemented. 

## Workload isolation

Isolation has multiple parts. A part of it will avoid or mitigate noisy neighbours, as in other workloads running on the same node that cause disturbances to the observed pod. 

LimitRanges, ResourceQuotas will prevent bestEffort pods in the cluster which will harm the ability to detect cluster scaling thresholds, and ResourceQuotas will ensure that no pod can take over a whole node, grow beyond the available limits or any other kind of size limits for a pod or a namespace. 

NetworkPolicies will ensure that the pods in a namespace cannot talk to pods or services in other namespaces, with the required exceptions for ingress, monitoring and such. NetworkPolicies require a Controller that implements them (VPC CNI+amazon network policy controller, Cilium, VPC CNI+Calico's Policy-only).

Compared to Cilium, VPC CNI relies on AWS VPC IPAM, requiring "real" VPC IPs attached to each EC2 Network interface. One must ensure that those IPs are not routable through the network to ensure the network isolation is not broken. 

OTOH, using Cilium as a full CNI replacement introduces complexities and limitations in terms of attaching webhooks to host IPs and keeping them available from within the cluster. 

VPC CNI is a sane default that comes with EKS and provides the required features and integration with AWS Network fabric.

## User permissions

To manage "user permissions" in kubernetes we need an external identiy provider that will provide a form of mapping of each individual wanting to access resources in the cluster and the namespaces where those actions can be applied. 

we need an Identity Provider that will maintain the correspondence between user personas, their teams and the namespaces they can manage. 

Each namespace will have Roles, ClusterRoles and RoleBindings that will define what resources they can manipulate.

In EKS, we can rely on AWS EKS Access Entries integration (static IAM roles), maintaining the list of users in AWS, or integrate with an OIDC provider of some sort (okta, AWS IAM Identiy Center, Dex, keycloak...).

AWS EKS Access Entries, hence relying on a static set of AWS Roles that users can assume, is simpler but requires maintaining the list of roles, and the users knowing which IAM Role to assume. Offboarding requires ensuring that the AWS Credentials the users use, are disabled. 

Using an OIDC Provider, it is more complex to initially integrate, but management is integrated with the rest of the company. Onboarding and Offboarding are automated and tied to the general user lifecycle.  

For a small team on a single cluster IAM Access Entries is the obvious choice. Less to operate, good enough security, fits the AWS stack.

# ingress

Ingress, as in the ability to serve Requests from the cluster is a core functionality of any web application. 

ingress-nginx was the defacto server for the HTTP Ingress use-case. It is being retired this month of March 2026, making it no longer an option for new projects. 

In terms of CRDs, Ingress and Gateway API overlap in terms of features. Kubernetes development considers Ingress API frozen and feature complete and directs users towards Gateway API. 

For new projects, the recommendation would be to move towards Gateway API implementations or ingress controllers that manages both APIs. A good default OSS implementation that's backed by the CNCF Foundation, is Istio. 

*Istio* and Gateway API CRDs. This combination allows for using both classic Ingress API objects and Gateway API CRDs, along with custom implementations to cover for complex use-cases not yet offered by GatewayAPI.

## Why not traefik? 
Traefik is also a renowned contender for this realm. Compared to Istio, Traefik governance is managed by a private company that holds full ownership of the project, its licensing, and direction. This can cause a significant clash between the company priorities for theis portfolio and the community usage of their opensource offering. 

Istio is, OTOH, a CNCF graduated project. Which ensures a degree of transparency in their government and expectations comparable to other CNCF-backed projects (argocd, kubernetes, prometheus, ... )

In a technical plane, online documentation and support for traefik, can, IMO, be improved significantly.

# TLS Certificates

To expose the application using TLS encryption to final users, we need to be able to issue certificates for the domains of the applications managed in the cluster. 

A popular option nowadays is issuing certificates dinamically using the ACME protocol. This solution improves the TLS Certificates management and recreation. Also, it automates the renewals. 

Let's encrypt is a popular vendor for free short-lived valid TLS Certificates. At the time of writing, their certificates expire in 90 days. In the future, the duration will shrink to 45 days or lower, increasing the need for a managed renewal process.

A popular ACME protocol implementation in kubernetes is *cert-manager*. when using supported certificate Issuers and DNS providers, it automatically manages the full lifecycle of a Certificate and the domain ownership validation. 

*Cert-manager* can issue certificates using the Certificate CRD, or react to annotated Ingress and GatewayAPI objects. This integration simplifies the amount of objects required to provision.  

# DNS Management

When creating a new Ingress in the cluster, we usually assign a DNS Name to it. THese DNS names must be created in form of the appropiate DNS Records in a publicly available DNS Server. 

A good practice here for operation security and scope-reduction, is to have a cluster-managed DNS Zone where the applications have a valid DNS record. Application maintainers will create CNAMEs on their respective DNS Zones to the application DNS record. This reduces the potential risk it can come with a credential leak of a large publicly branded DNS Zone.

*external-dns* manages the DNS Records on our behalf. It reacts to annotated Ingresses, Gateways, Services of `type=LoadBalancer` or `DNSEndpoint` objects. 

It requires valid credentials/API Keys in the DNS provider of choice with enough permissions to manage the DNS Zone or Zones assigned to the application.

# secret management 

Secrets are of many kinds: repository credentials, DNS Zones, Certificate Providers, Container registries, application-specific secret values... 

There are two approaches possible, that are non-exclusive. Each can work for different usecases. 

## sealed-secrets

The cluster provides a public key (can be shared liberally) that cluster users will use along the sealed-secrets' CLI `kubeseal` tool. The resulting kubernetes object can be safely stored in repositories with no problem, makin the secrets an integral part of the application.

By keeping a known set of certificates we ensure the readability of the screts for the time being. 

All secrets stored as SealedSecrets will be decrypted into generic kubernetes Secrets when _unsealed_ by the *sealed-secrets* operator.

## external-secrets

When the secrets are stored in a specialized service of some kind (AWS Secret Manager, Vault, 1password,...) external-secrets allow to fetch secrets from these services and expose them as kubernetes Secrets. The operator will take care of reloading them periodically to ensure they are fresh. 



# Additional niceties

## Stakater reloader
When working with custom configurations, or refreshed secrets, we need a way to ensure the change is picked up by the consumers. 

`Reloader` will monitor secrets and configmap changes and refresh the Pods when the associated configmap or secret changes.

## Noe
When using mixed-architecture clusters, by default and without specific helpers in the pod definition (`nodeAffinity` , `nodeSelectors`), kubernetes scheduler does not know if the container wil run on the selected node. And for that to run in both, one need to have all the images published as multiarch, to ensure they will run regardless of the scheduled node. 

Noe is a Kubernetes mutating webhook that dynamically assigns node architectures to match the requirements of container images within a Pod. It simplifies mixed-architecture deployments (e.g. ARM and x86) by ensuring that Pods are scheduled on nodes capable of executing all their images.

## well-known API ingress endpoint 
When using EKS or other managed control plane service, the URL of the cluster is autogenerated and under provider DNS domains. This resulting URL is hard to identify as one's. Also, the network access rules may be limiting, or not suited for the networking layout we need. 

By having a custom ingress class with an Ingress object pointing to the API we can have a fully-controlled API endpoint that's subjected to the general networking rules we have in place. 

Technically speaking, is a specific ingress class with an Ingress object pointing to the kubernetes Service that exposes the cluster IP. 

## Velero
Eventho the cluster for this application is mainly stateless and can be redeployed and being up to speed with not much hassle, having backups of the objects in there can be useful in case of selective recovery or auditing. 

Velero is a fully fledged backup operator for kubernetes. It can back up Persistent Volumes and k8s objects following the defined rules and policies. Its state is stored in the actual backup storages.  