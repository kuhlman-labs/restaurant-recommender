variable "location" {
  description = "The Azure region to deploy resources to."
  type        = string
  default     = "EastUS 2"
}

variable "resource_group_name" {
  description = "The name of the Resource Group."
  type        = string
  default     = "restaurant-recommender-rg"
}

variable "app_service_plan_name" {
  description = "The name of the App Service Plan."
  type        = string
  default     = "rr-asp00173"
}

variable "container_registry_name" {
  description = "The name of the Container Registry (must be globally unique)."
  type        = string
  default     = "rracr00173"
}

variable "webapp_name" {
  description = "The name of the Web App Service."
  type        = string
  default     = "rrwebapp00173"
}

variable "sql_server_name" {
  description = "The name of the Azure SQL Server (must be globally unique)."
  type        = string
  default     = "rrsqlserver00173"
}

variable "sql_database_name" {
  description = "The name of the Azure SQL Database."
  type        = string
  default     = "rrdb"
}

variable "client_secret" {
  description = "The client secret for the Azure AD Service Principal."
  type        = string
  }
