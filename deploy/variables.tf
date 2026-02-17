variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  default     = "southamerica-east1"
  description = "GCP region (default: São Paulo)"
}

variable "db_password" {
  type        = string
  sensitive   = true
  description = "Cloud SQL database password"
}

variable "jwt_secret" {
  type        = string
  sensitive   = true
  description = "JWT signing secret"
}

variable "image_tag" {
  type        = string
  default     = "latest"
  description = "Docker image tag for Cloud Run"
}

variable "github_repo" {
  type        = string
  description = "GitHub repository (owner/repo) allowed to authenticate via Workload Identity Federation"
}
