"""
Architecture diagram for little-sample-cluster
Produces three output files:
  - architecture_layer1.png  (AWS region / VPC / subnets)
  - architecture_layer2.png  (request flow: internet → app → RDS)
  - architecture_layer3.png  (cluster internals + GitOps + observability)
"""

from diagrams import Diagram, Cluster, Edge
from diagrams.aws.network import (
    PublicSubnet, PrivateSubnet, NATGateway,
    ElasticLoadBalancing, Route53, ElbNetworkLoadBalancer
)
from diagrams.aws.compute import EKS, EC2, EC2AutoScaling, ECR
from diagrams.aws.database import RDS
from diagrams.aws.storage import S3
from diagrams.aws.security import ACM, SecretsManager
from diagrams.aws.management import Cloudwatch
from diagrams.k8s.compute import Deployment, Pod, ReplicaSet
from diagrams.k8s.network import Ingress, Service, Endpoint
from diagrams.k8s.rbac import Role, RoleBinding, ClusterRole
from diagrams.k8s.clusterconfig import HPA
from diagrams.k8s.podconfig import Secret, CM
from diagrams.k8s.others import CRD
from diagrams.onprem.gitops import ArgoCD
from diagrams.onprem.monitoring import Prometheus, Grafana
from diagrams.onprem.logging import Fluentbit
from diagrams.onprem.network import Istio, Internet
from diagrams.onprem.vcs import Github
from diagrams.onprem.certificates import CertManager, LetsEncrypt
from diagrams.generic.blank import Blank

GRAPH = {"fontsize": "17", "pad": "0.8", "ranksep": "0.9", "nodesep": "0.5"}


# ─────────────────────────────────────────────────────────────
# Layer 1 — AWS Region, VPC, Subnets
# ─────────────────────────────────────────────────────────────
with Diagram(
    "Layer 1 — AWS Infrastructure (eu-west-1)",
    filename="architecture_layer1",
    outformat="png",
    show=False,
    direction="TB",
    graph_attr=GRAPH,
):
    internet = Internet("Internet")

    with Cluster("AWS eu-west-1"):

        with Cluster("Regional Services"):
            route53 = Route53("Route 53")
            # acm     = ACM("ACM / cert-manager")
            # ecr     = ECR("ECR")
            s3      = S3("S3 (Velero)")
            cw      = Cloudwatch("CloudWatch Logs")
            sm      = SecretsManager("Secrets Manager")
            eks_cp  = EKS("EKS Control Plane\n(AWS Managed, multi-AZ)")


        with Cluster("VPC  10.x.0.0/16"):
            with Cluster("AZ  eu-west-1a"):
                with Cluster("Public Subnet"):
                    pub_a  = ElbNetworkLoadBalancer("Public (NLB)")
                    nat_a  = NATGateway("NAT GW")
                with Cluster("Private Subnet"):
                    boot_node1 = EC2AutoScaling("Bootstrap nodes")
                    worker_node1 = EC2("k8s node 1")
                    worker_node2 = EC2("k8s node 2")
                with Cluster("Data Subnet"):
                    db_node_1 = RDS("App Database")

            with Cluster("AZ  eu-west-1b"):
                with Cluster("Public Subnet"):
                    pub_b  = ElbNetworkLoadBalancer("Public (NLB)")
                    nat_b  = NATGateway("NAT GW")
                with Cluster("Private Subnet"):
                    boot_node2 = EC2AutoScaling("Bootstrap nodes")
                    worker_node4 = EC2("k8s node 4")
                    worker_node5 = EC2("k8s node 5")
                with Cluster("Data Subnet"):
                    db_node_2 = RDS("App Database read replica")

            with Cluster("AZ  eu-west-1c"):
                with Cluster("Public Subnet"):
                    pub_c  = ElbNetworkLoadBalancer("Public (NLB)")
                    nat_c  = NATGateway("NAT GW")
                with Cluster("Private Subnet"):
                    boot_node3 = EC2AutoScaling("Bootstrap nodes")
                    worker_node3 = EC2("k8s node 3")
                    worker_node6 = EC2("k8s node 6")
                with Cluster("Data Subnet"):
                    db_node_3 = Blank("")

        internet >> route53

        internet >> [pub_a, pub_b, pub_c]
        [nat_a, nat_b, nat_c] >> internet
        pub_a >> [boot_node1, worker_node1, worker_node2]
        pub_b >> [boot_node2, worker_node4, worker_node5]
        pub_c >> [boot_node3, worker_node3, worker_node6]

        [boot_node1, worker_node1, worker_node2] >> db_node_1
        [boot_node1, worker_node1, worker_node2] >> db_node_2
        #[boot_node1, worker_node1, worker_node2] >> db_node_3

        [boot_node2, worker_node4, worker_node5] >> db_node_1
        [boot_node2, worker_node4, worker_node5] >> db_node_2
        #[boot_node2, worker_node4, worker_node5] >> db_node_3

        [boot_node3, worker_node3, worker_node6] >> db_node_1
        [boot_node3, worker_node3, worker_node6] >> db_node_2
        #[boot_node3, worker_node3, worker_node6] >> db_node_3


        # eks_cp - [ecr, s3, cw, sm]
        eks_cp - [s3, cw, sm]


# ─────────────────────────────────────────────────────────────
# Layer 2 — Request Flow: Internet → Istio → App → RDS
# ─────────────────────────────────────────────────────────────
with Diagram(
    "Layer 2 — Request Flow",
    filename="architecture_layer2",
    outformat="png",
    show=False,
    direction="LR",
    graph_attr={**GRAPH, "ranksep": "1.2"},
):
    internet = Internet("Internet\nHTTPS")

    with Cluster("AWS eu-west-1"):

        route53  = Route53("Route 53\n(external-dns)")
        sm       = SecretsManager("AWS Secrets\nManager")
        nlb      = ElasticLoadBalancing("NLB\n(AWS LB Controller)")

        with Cluster("EKS Cluster"):

            with Cluster("ns: istio-system"):
                istio = Istio("Istio Gateway\n(HTTPRoute)")

            with Cluster("ns: app"):
                tls_cert = LetsEncrypt("TLS Cert")
                svc      = Service("Service\n(ClusterIP)")
                dep      = Deployment("Deployment")
                rs       = ReplicaSet("ReplicaSet")
                ep_slice = Endpoint("Endpoint Slice")
                pod1     = Pod("Pod")
                pod2     = Pod("Pod")
                pod3     = Pod("Pod")
                hpa      = HPA("HPA")
                vpa      = CRD("VPA")
                ext_sec  = CRD("ExternalSecret")
                secret   = Secret("Secret")

            with Cluster("ns: cert-manager"):
                certcrd = CertManager("Certificate CRD")

        rds = RDS("Amazon RDS\nMySQL (Multi-AZ)")

    internet >> route53 >> nlb
    tls_cert >> istio
    nlb >> istio >> svc >> ep_slice >> [pod1, pod2, pod3]
    dep - rs - [pod1, pod2, pod3]
    [pod1, pod2, pod3] >> rds
    vpa >> [pod1, pod2, pod3]
    hpa >> dep
    certcrd >> tls_cert
    sm >> ext_sec >> secret >> dep


# ─────────────────────────────────────────────────────────────
# Layer 3 — Cluster Internals, GitOps, Observability
# ─────────────────────────────────────────────────────────────
with Diagram(
    "Layer 3 — Cluster Internals & GitOps",
    filename="architecture_layer3",
    outformat="png",
    show=False,
    direction="LR",
    graph_attr={**GRAPH, "ranksep": "1.0", "nodesep": "0.6"},
):
    git = Github("Git Repository\n(app-of-apps)")
    s3  = S3("S3\n(Velero backups)")
    cw  = Cloudwatch("CloudWatch\n(API & audit logs)")
    sm  = SecretsManager("AWS Secrets\nManager")

    with Cluster("EKS Cluster — eu-west-1"):

        with Cluster("ns: kube-system"):
            vpc_cni   = CRD("VPC CNI + \nNetPol Controller")
            coredns   = CRD("CoreDNS")
            kubeproxy = CRD("kube-proxy")
            pod_id    = CRD("Pod Identity Agent")
            karpenter  = CRD("Karpenter")
            awslb      = CRD("AWS LB Controller")
            ack        = CRD("AWS Cloud Controller")
            csi        = CRD("AWS CSI EBS Driver")

        with Cluster("Bootstrap NodeGroup  (2x m8g.xlarge)"):
            boot_nodes = EC2("Static EC2 Nodes")

        with Cluster("Karpenter NodePool"):
            karp_nodes = EC2AutoScaling("Dynamic EC2\n(on-demand / spot)")

        with Cluster("ns: argocd"):
            argocd = ArgoCD("ArgoCD\n(app-of-apps)")

        with Cluster("ns: platform"):
            certmgr    = CRD("cert-manager")
            extdns     = CRD("external-dns")
            ext_sec    = CRD("external-secrets")
            sealed_sec = CRD("sealed-secrets")
            vpa        = CRD("VPA Operator")
            keda       = CRD("KEDA")
            reloader   = CRD("Stakater Reloader")
            noe        = CRD("Noe\n(arch webhook)")
            velero_op  = CRD("Velero")
            np_ctrl    = CRD("AWS NP Controller")

        with Cluster("ns: istio-system"):
            istio_gw   = Istio("Istio Gateway")
            istio_hpa      = HPA("HPA")
            istio_vpa      = CRD("VPA")


        with Cluster("ns: monitoring"):
            prometheus = Prometheus("Prometheus")
            grafana    = Grafana("Grafana")
            alertmgr   = CRD("Alertmanager")
            fluentbit  = Fluentbit("Fluentbit\n(DaemonSet)")
            node_exp   = CRD("node-exporter\n(DaemonSet)")
            mysql_exp  = CRD("mysql-exporter")

        with Cluster("ns: app"):
            with Cluster("<team>  (per namespace)"):
                ns_role   = Role("Role")
                ns_rb     = RoleBinding("RoleBinding")
                ns_cr     = ClusterRole("ClusterRole\n(reader)")
                ns_netpol = CRD("NetworkPolicy")
                ns_quota  = CRD("ResourceQuota\nLimitRange")
            with Cluster("App resources"):
                app_ing = Ingress("HTTPRoute")
                app_tls_cert = LetsEncrypt("TLS Cert")
                app_svc      = Service("Service\n(ClusterIP)")
                app_dep      = Deployment("Deployment")
                app_rs       = ReplicaSet("ReplicaSet")
                app_ep_slice = Endpoint("Endpoint Slice")
                app_pod1     = Pod("Pod")
                app_pod2     = Pod("Pod")
                app_pod3     = Pod("Pod")
                app_hpa      = HPA("HPA")
                app_vpa      = CRD("VPA")
                app_ext_sec  = CRD("ExternalSecret")
                app_secret   = Secret("Secret")
                app_pm = CRD("PodMonitor")
                rds_crd = CRD("RDS MySQL")

    # ── GitOps flow ──
    git >> argocd
#     argocd >> [
#         karpenter, awslb, ack, certmgr, extdns,
#         ext_sec, sealed_sec, istio_gw, vpa, keda,
#         reloader, noe, velero_op, np_ctrl
#     ]
#     argocd >> app_dep
#     argocd >> [ns_role, ns_rb, ns_netpol, ns_quota]

    # ── Node provisioning ──
    karpenter >> karp_nodes
    boot_nodes - [coredns, karpenter, vpc_cni]

    # ── Observability ──
    app_pm  >> prometheus
    node_exp >> prometheus
    mysql_exp >> prometheus
    prometheus >> grafana
    prometheus >> alertmgr
    fluentbit >> cw

    # ── ACK → RDS CRD ──
    ack >> rds_crd

    # ── Secrets ──
    sm >> ext_sec >> app_ext_sec
    sealed_sec >> app_secret
    reloader >> app_dep

    # ── Velero backup ──
    velero_op >> s3

    # istio
    istio_hpa >> istio_gw
    istio_vpa >> istio_gw
    istio_gw >> app_ing

    # ── App wiring ──
    app_ing >> app_svc
    app_hpa >> app_dep
    app_dep - karp_nodes

    app_dep >> app_svc >> app_ep_slice >> [app_pod1, app_pod2, app_pod3]
    [app_pod1, app_pod2, app_pod3] - app_vpa
    app_dep >> app_rs >> [app_pod1, app_pod2, app_pod3]

print("Generated:")
print("  architecture_layer1.png")
print("  architecture_layer2.png")
print("  architecture_layer3.png")
