#!/usr/bin/env bash
set -euo pipefail

# ─── Prerequisites check ────────────────────────────────────────────────────
for cmd in kind kubectl helm docker jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not found in PATH." >&2
    exit 1
  fi
done

read -r -s -p "Enter MySQL password: " INCLUSTER_MYSQL_PASSWORD
echo
if [[ -z "$INCLUSTER_MYSQL_PASSWORD" ]]; then
  echo "ERROR: MySQL password cannot be empty." >&2
  exit 1
fi

read -r -p "Enter ghcr.io username: " INCLUSTER_GHCR_USERNAME
echo
if [[ -z "$INCLUSTER_GHCR_USERNAME" ]]; then
  echo "ERROR: ghcr.io username cannot be empty." >&2
  exit 1
fi

read -r -s -p "Enter ghcr.io password/token: " INCLUSTER_GHCR_PASSWORD
echo
if [[ -z "$INCLUSTER_GHCR_PASSWORD" ]]; then
  echo "ERROR: ghcr.io password cannot be empty." >&2
  exit 1
fi

CLUSTER_NAME="little-sample-cluster"
NAMESPACE="kube-public"
RELEASE="little-cluster"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHART_DIR="$SCRIPT_DIR/../chart/little-sample-cluster"

# ─── kind cluster ───────────────────────────────────────────────────────────
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "=> kind cluster '${CLUSTER_NAME}' already exists, skipping creation."
else
  echo "=> Creating kind cluster '${CLUSTER_NAME}'..."
  kind create cluster --config "$SCRIPT_DIR/kind-cluster.yaml"
fi

kubectl config use-context "kind-${CLUSTER_NAME}"

# ─── Namespace ──────────────────────────────────────────────────────────────
echo "=> Ensuring namespace '${NAMESPACE}'..."
kubectl get namespace "$NAMESPACE" &>/dev/null || kubectl create namespace "$NAMESPACE"

# ─── ingress-nginx ──────────────────────────────────────────────────────────
echo "=> Labelling control-plane node for ingress-nginx scheduling..."
kubectl label node "${CLUSTER_NAME}-control-plane" ingress-ready=true --overwrite

echo "=> Installing ingress-nginx..."
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx --force-update 2>/dev/null || true
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --values "$SCRIPT_DIR/ingress-nginx.yaml" \
  --wait \
  --timeout 3m

# ─── Helm dependencies ──────────────────────────────────────────────────────
echo "=> Resolving Helm dependencies..."
helm repo add kubelauncher https://kubelauncher.github.io/charts --force-update 2>/dev/null || true
helm dependency build "$CHART_DIR"

# ─── Deploy ─────────────────────────────────────────────────────────────────
echo "=> Deploying Helm release '${RELEASE}' to namespace '${NAMESPACE}'..."
helm upgrade --install "$RELEASE" "$CHART_DIR" \
  --namespace "$NAMESPACE" \
  --values "$SCRIPT_DIR/values-override.yaml" \
  --set "global.mysql.auth.password=${INCLUSTER_MYSQL_PASSWORD}" \
  --set "registryCredentials.ghcr\\.io.username=${INCLUSTER_GHCR_USERNAME}" \
  --set "registryCredentials.ghcr\\.io.password=${INCLUSTER_GHCR_PASSWORD}" \
  --wait \
  --timeout 5m

# ─── Verify ─────────────────────────────────────────────────────────────────
echo ""
echo "=> Deployment status:"
kubectl get pods -n "$NAMESPACE"
echo ""
echo "=> Application is available at http://localhost:8888"
echo "   curl -X PUT -H 'Content-Type: application/json' --data '{\"dateOfBirth\": \"1999-12-01\"}' http://localhost:8888/hello/newuser"
echo "   curl http://localhost:8888/hello/newuser"
echo ""
echo "Done."
