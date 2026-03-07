# Going to production

a proper architecture for this application, when having AWS as a main infrastructure provider, requires some planning and decisions. 

For this application we will go with some managed services where they make sense while still keeping operational flexibility and ownership. 

# runtime
For the runtime, either ECS or EKS could work as container runtime for this specific application. 

ECS is a higher level abstraction that reduces the operational complexity of maintaining a container orchestration infrastructure. From a user's perspective you only see either just your workloads (Fargate serverless) or the ec2 instances (plain ECS) and everything else is orchestrated and managed. One of its biggest cons is the close tight of the observability tooling and debugging to AWS services (Cloudwatch) and AWS Support for infrastructure debugging. 

EKS rises the abstraction compared to self-deploying your own Kubernetes infrastructure. It still offers some managed integrations, and it's tighly coupled to AWS Services. It offers optionality on what of the offered services you want AWS to manage for you or if you prefer to operate them yourself. The opreational load is higher than when using ECS, or fargate, but can still manageable once deployed and properly dimensioned. AWS will take care of dimensioning the API and ETCD servers according to the load detected and cluster size. There's still some dependencies on AWS Cloudwatch but limited to API logs and some infrastructure metrics. 

Given the stability requirements and the operational work that comes with it, familiarity with the technology and operational flexibility that offers having access to a fully fledged K8S, I'll pick Amazon EKS as the runtime.

It's true that, in a production environment with a single high traffic application exactly like this exercise, picking ECS would be the most optimal solution even accounting for the added reliance on many AWS services and the lack of industry-standard tooling.

In terms of worker nodes, we only need one or two nodegroup-managed instances for bootstrapping the cluster. Cluster autoscaler activities can be managed and defined at runtime using Karpenter and its concept of nodepools. This autoscaler component replaces or can operate independently of upstream's CNCF cluster-autoscaler. It's aware of the instance definitions, equivalences and prices of AWS and Azure instances, and can pick the most suited instance to fit the amount of pending pods plus the estimated Daemonset overhead. 

For Karpenter to work it needs an instance where it's running before it takes over the instance provisioning activities. As we also need an instance to run a starter coreDNS and kubelet, it generates little waste. we can use plain eks nodegroups for this.

There are few other controllers we need to ensure the cluster can be bootstrapped and operated. Many of them can be provided by AWS at instalation time and their lifecycle is managed using AWS API/Terraform Module/AWS Console. Only a small subset is convenient enough to be managed and the rest offer more freedom when they are self-managed instead. Moving from managed to self-provided and viceversa is not an easy two-way door. 

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



# Database: 
- Amazon RDS
Given that the RDS offering is solid, and supports MySQL LTS versions, it's best if we go with the offering instead of hosting our own. This will free some cognitive load from the team and reduces the risk of the cluster operations as the state is hosted outside of it. 

This decision doesn't come for free. When setting up the DB Instance we will need to specify, at least, our own `Parameter Groups`. This will ease in the future customizing settings in the RDS database without rebuilding. 

We will need to evaluate replication, volume performance, costs,...  

Also, the performance of the database still needs to be monitored, by using prometheus-mysql-exporter (or any appropiate metrics extraction toool suited for our flavour of RDS).

In terms of resources and costs, we can go with db.m8g.large (2cores, 8gb) for ~122$/month or even lower to a db.t4g.medium (1c, 4gb) for ~50$/month.

These are safe starting points to ensure some stable performance. With the actual application running we could adjust better the instance type. 

The current helm chart of the application would need to be modified to remove the dependency on mysql chart. If we want to keep the approach of linking the application and ints dependencies,  we could replace the current dependency with a new in-chart `DBInstance.rds.services.k8s.aws/v1alpha1` object to provision and maintain the RDS instance as a kubernetes object. This approach requires the Amazon Controller K8S (ACK) installed with the RDS controller enabled. 

A different approach by only providing a secret could be taken if there's a centrally-provided big multitenant Database instance, with a team operating it as a service offering.

# infrastructure deployment
A good approach to managing the infrastructure would be using terragrunt for templating the environments and all variables terraform requires. Even if we have (for now) one single environment/cluster/account.

then, terraform takes the lead of provisioning the infrastructure. 

Here, I'll separate what's an account or regional resource that's shared among every other component like, DNS Zones (generally, could be different), VPC and network foundations, Peering/TransitGateway management,... and any other kind of very shared resource. Some IAM roles may fall in here as well. 

Then, the cluster-related Terraform modules. In here we define all the required resources each cluster requires like Security groups, Instance roles, EKS cluster, eks addons, bootstrap nodegroups, secrets, certificates... It's also the moment to install the required bootstrapping components taht are self-managed like argocd and container registry credentials using the Kubernetes Provider's kubernetes_manifest. 

Also, it's time to apply the argocd applications for the critical components. These can be done using the same TF modules or externally after the TF section is done. 

## Cluster application deployment
Application deployment, can be deployed and maintained using different ways. One that's becoming a best practice and common in the industry is folowing a Gitops model. 

In a Gitops model, the deployment is initiated when a change is detected in a monitored repository. The detection can be pushed to the cluster externally and apply the manifests, or use dedicated tooling from inside the cluster like Flux or ArgoCD. 

Any can work, I'll pick ArgoCD. Eventho is fully CRD and CLI based, having a UI causes no harm and allows for checking at a quick glance what's the state of the applications.

As ArgoCD is one of the first applications provisioned and properly configured from its helm chart, it can be used to provision ad maintain all the components required. even before having ingress, DNS management, TLS Certificates... 

# application deployment 
Our little application is delivered as a helm chart and a container stored in a container registry. There's an argocd application available in the `application` folder. 

In a production environment, the argocd application would be updated either directly in the cluster by our CD system using valid system credentials or using an intermediate deployment tooling for advanced logic in the deployment.  

# Observability:
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

API and audit logs are available in CLoudwatch logs. we will need an extractor based on lambda SQS and SNS subscriptions to extract them. 

for general logs. Fluentbit is a defacto extractor. 


## visualization

## alerting

# namespace management

# isolation and security
namespace isolation, network policy, 

# ingress

# TLS Certificates

# DNS Management

# secret management 

