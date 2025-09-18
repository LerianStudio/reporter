#!/bin/sh
set -eu

# Enforce non-empty JWT keys
if [ -z "${SEAWEEDFS_JWT_SIGNING_WRITE:-}" ] || [ -z "${SEAWEEDFS_JWT_SIGNING_READ:-}" ]; then
  echo "[FATAL] SEAWEEDFS_JWT_SIGNING_READ/WRITE must be set and non-empty to start Filer with JWT."
  exit 1
fi

# Generate the file /etc/seaweedfs/security.toml using the envs value
mkdir -p /etc/seaweedfs
cat > /etc/seaweedfs/security.toml <<EOF
[jwt.filer_signing]
key = "${SEAWEEDFS_JWT_SIGNING_WRITE:-}"

[jwt.filer_signing.read]
key = "${SEAWEEDFS_JWT_SIGNING_READ:-}"
EOF

echo "security.toml generated in /etc/seaweedfs/security.toml"

# Start Filer. security.toml will be auto-discovered from /etc/seaweedfs/
exec weed filer -ip=plugin-smart-templates-seaweedfs-filer -port=8888 -master=plugin-smart-templates-seaweedfs-master:9333


