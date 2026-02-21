resource "google_service_account" "cloud_run" {
  account_id   = "finance-cloud-run"
  display_name = "Finance Cloud Run Service Account"
}

resource "google_project_iam_member" "cloud_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloud_run.email}"
}

locals {
  image_url          = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.finance.repository_id}/finance:${var.image_tag}"
  db_connection_name = google_sql_database_instance.finance.connection_name
}

resource "google_cloud_run_v2_service" "finance" {
  name                = "finance"
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    service_account = google_service_account.cloud_run.email

    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [local.db_connection_name]
      }
    }

    containers {
      image = local.image_url

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "DATABASE_URL"
        value = "postgres://finance:${var.db_password}@/${google_sql_database.finance.name}?host=/cloudsql/${local.db_connection_name}"
      }

      env {
        name = "JWT_SECRET"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.jwt_secret.secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "STATIC_DIR"
        value = "./static"
      }

      env {
        name  = "ALLOWED_ORIGIN"
        value = var.domain
      }

      env {
        name = "SENDGRID_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.sendgrid_api_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "EMAIL_FROM"
        value = var.email_from
      }

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      ports {
        container_port = 8080
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 5
        period_seconds        = 5
        failure_threshold     = 5
      }
    }
  }

  depends_on = [
    google_project_service.apis,
    google_secret_manager_secret_version.jwt_secret,
    google_secret_manager_secret_version.sendgrid_api_key,
  ]

}

# Allow public access
resource "google_cloud_run_v2_service_iam_member" "public" {
  name     = google_cloud_run_v2_service.finance.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}