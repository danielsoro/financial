#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$SCRIPT_DIR"

if [ ! -f terraform.tfvars ]; then
  echo "Error: terraform.tfvars not found. Copy terraform.tfvars.example and fill in values."
  exit 1
fi

PROJECT_ID=$(grep 'project_id' terraform.tfvars | sed 's/.*= *"\(.*\)"/\1/')
REGION=$(grep 'region' terraform.tfvars | sed 's/.*= *"\(.*\)"/\1/' || echo "southamerica-east1")
REGION="${REGION:-southamerica-east1}"
IMAGE_TAG="${1:-latest}"
REPO="${REGION}-docker.pkg.dev/${PROJECT_ID}/finance"
IMAGE="${REPO}/finance:${IMAGE_TAG}"

# Step 1: Ensure infrastructure exists (APIs, Artifact Registry, Cloud SQL, secrets)
echo "==> Provisioning infrastructure..."
terraform apply -var "image_tag=${IMAGE_TAG}"

# Step 2: Build Docker image
echo "==> Building Docker image..."
cd "$PROJECT_ROOT"
docker build -t "finance:${IMAGE_TAG}" .

# Step 3: Push to Artifact Registry
echo "==> Configuring Docker auth for Artifact Registry..."
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

echo "==> Tagging image: ${IMAGE}"
docker tag "finance:${IMAGE_TAG}" "${IMAGE}"

echo "==> Pushing image..."
docker push "${IMAGE}"

# Step 4: Update Cloud Run with the new image
echo "==> Updating Cloud Run service..."
cd "$SCRIPT_DIR"
terraform apply -var "image_tag=${IMAGE_TAG}" -auto-approve

echo ""
echo "==> Deploy complete!"
terraform output cloud_run_url
