
if ! command -v docker-credential-osxkeychain &> /dev/null; then
  echo "Error: docker-credential-osxkeychain not found in PATH. If you are not using a Mac, edit the script appropiately. linux helper is docker-credential-secretservice"
  exit 1
fi
CREDS=$(echo "https://ghcr.io" | docker-credential-osxkeychain get)


USER=$(echo "$CREDS" | jq -r .Username)
PASS=$(echo "$CREDS" | jq -r .Secret)
AUTH=$(printf "%s:%s" "$USER" "$PASS" | base64)

jq -n \
  --arg u "$USER" \
  --arg p "$PASS" \
  --arg a "$AUTH" \
  '{
    auths: {
      "ghcr.io": {
        username: $u,
        password: $p,
        auth: $a
      }
    }
  }' | kubectl create secret generic regcred \
       --type=kubernetes.io/dockerconfigjson \
       --from-file=.dockerconfigjson=/dev/stdin

echo "Created secret regcred in current namespace and context"