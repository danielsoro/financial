resource "google_sql_database_instance" "finance" {
  name             = "finance-db"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier              = "db-f1-micro"
    edition           = "ENTERPRISE"
    availability_type = "ZONAL"
    disk_size         = 10

    backup_configuration {
      enabled = true
    }

    ip_configuration {
      ipv4_enabled = true
    }
  }

  deletion_protection = true

  depends_on = [google_project_service.apis]
}

resource "google_sql_database" "finance" {
  name     = "finance"
  instance = google_sql_database_instance.finance.name
}

resource "google_sql_user" "finance" {
  name     = "finance"
  instance = google_sql_database_instance.finance.name
  password = var.db_password
}
