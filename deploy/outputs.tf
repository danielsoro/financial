output "cloud_run_url" {
  value       = google_cloud_run_v2_service.finance.uri
  description = "Finance app URL"
}
output "cloud_sql_connection_name" {
  value       = google_sql_database_instance.finance.connection_name
  description = "Cloud SQL instance connection name"
}

output "artifact_registry_repo" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.finance.repository_id}"
  description = "Artifact Registry repository URL"
}

output "github_actions_workload_identity_provider" {
  value       = google_iam_workload_identity_pool_provider.github.name
  description = "Workload Identity Provider resource name (use as GCP_WORKLOAD_IDENTITY_PROVIDER GitHub secret)"
}

output "github_actions_service_account" {
  value       = google_service_account.github_actions.email
  description = "GitHub Actions service account email (use as GCP_SERVICE_ACCOUNT GitHub secret)"
}

output "domain" {
  value       = var.domain
  description = "Root domain"
}

output "tenant_urls" {
  value       = { for name, record in cloudflare_record.tenant : name => "https://${record.hostname}" }
  description = "Tenant subdomain URLs"
}

