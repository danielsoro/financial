resource "google_artifact_registry_repository" "finance" {
  location      = var.region
  repository_id = "finance"
  format        = "DOCKER"
  description   = "Finance app Docker images"

  depends_on = [google_project_service.apis]
}
