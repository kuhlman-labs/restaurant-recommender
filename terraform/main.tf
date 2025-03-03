###########################
# Data Resources
########################### 
data "azurerm_client_config" "current" {}

###########################
# Resource Group
###########################
resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.location
}

###########################
# App Service Plan (Linux)
###########################
resource "azurerm_service_plan" "asp" {
  name                = var.app_service_plan_name
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  os_type             = "Linux"
  sku_name            = "B1"
}

###########################
# Container Registry
###########################
resource "azurerm_container_registry" "acr" {
  name                = var.container_registry_name
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = true
  identity {
    type = "SystemAssigned"
  }
}

###########################
# Web App Service
###########################
resource "azurerm_linux_web_app" "webapp" {
  name                = var.webapp_name
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_service_plan.asp.location
  service_plan_id     = azurerm_service_plan.asp.id

  site_config {
    #container_registry_managed_identity_client_id = azurerm_container_registry.acr.identity[0].principal_id
    container_registry_use_managed_identity       = true
    application_stack {
      docker_image_name        = "restaurant-recommender:latest"
      docker_registry_url      = "https://${azurerm_container_registry.acr.login_server}"
      docker_registry_username = azurerm_container_registry.acr.admin_username
      docker_registry_password = azurerm_container_registry.acr.admin_password
    }
  }

  identity {
    type = "SystemAssigned"
  }

  app_settings = {
    "DB_SERVER" = azurerm_mssql_server.sqlserver.fully_qualified_domain_name
    "DB_NAME"   = azurerm_mssql_database.sqldb.name
    "DB_USER"   = data.azurerm_client_config.current.client_id
    "DB_PASS"   = var.client_secret

  }
  lifecycle {
    ignore_changes = [
      application_stack[0].docker_image_name,
    ]
  }
}

###########################
# Grant the Web App identity AcrPull on the Container Registry
###########################
resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_linux_web_app.webapp.identity[0].principal_id
}

###########################
# Azure SQL Server
###########################
resource "azurerm_mssql_server" "sqlserver" {
  name                = var.sql_server_name
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  version             = "12.0"
  azuread_administrator {
    azuread_authentication_only = true
    login_username              = "terraform-sp"
    object_id                   = data.azurerm_client_config.current.object_id
    tenant_id                   = data.azurerm_client_config.current.tenant_id
  }
}

###########################
# SQL Server Firewall: Allow Azure services
###########################
resource "azurerm_mssql_firewall_rule" "allow_azure_services" {
  name             = "AllowAzureServices"
  server_id        = azurerm_mssql_server.sqlserver.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}

###########################
# Azure SQL Database
###########################
resource "azurerm_mssql_database" "sqldb" {
  name                                = var.sql_database_name
  sku_name                            = "Basic"
  server_id                           = azurerm_mssql_server.sqlserver.id
  storage_account_type                = "Local"
  transparent_data_encryption_enabled = true
}