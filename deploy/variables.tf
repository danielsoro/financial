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

variable "github_repo" {
  type        = string
  description = "GitHub repository (owner/repo) allowed to authenticate via Workload Identity Federation"
}

variable "image_tag" {
  type        = string
  default     = "latest"
  description = "Docker image tag to deploy to Cloud Run"
}

variable "cloudflare_api_token" {
  type        = string
  sensitive   = true
  description = "Cloudflare API token"
}

variable "cloudflare_account_id" {
  type        = string
  description = "Cloudflare account ID"
}

variable "domain" {
  type        = string
  description = "Root domain (e.g., financeiro.app)"
}

variable "sendgrid_api_key" {
  type        = string
  sensitive   = true
  default     = ""
  description = "SendGrid API key (empty = LogSender in dev)"
}

variable "email_from" {
  type        = string
  default     = "noreply@dnafami.com.br"
  description = "Email sender address"
}
