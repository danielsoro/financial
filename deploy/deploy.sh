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
IMAGE_TAG="${1:-$(git -C "$PROJECT_ROOT" describe --tags --abbrev=0 2>/dev/null || echo latest)}"
REPO="${REGION}-docker.pkg.dev/${PROJECT_ID}/finance"
IMAGE="${REPO}/finance:${IMAGE_TAG}"

# Step 1: Build Docker image
echo "==> Building Docker image (tag: ${IMAGE_TAG})..."
cd "$PROJECT_ROOT"
docker build -t "finance:${IMAGE_TAG}" .

# Step 2: Push to Artifact Registry
echo "==> Configuring Docker auth for Artifact Registry..."
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

echo "==> Tagging image: ${IMAGE}"
docker tag "finance:${IMAGE_TAG}" "${IMAGE}"

echo "==> Pushing image..."
docker push "${IMAGE}"

# Step 3: Deploy to Cloud Run
echo "==> Deploying to Cloud Run..."
gcloud run deploy finance \
  --image "${IMAGE}" \
  --region "${REGION}" \
  --project "${PROJECT_ID}"

echo ""
echo "==> Deploy complete!"
