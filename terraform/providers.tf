provider "azurerm" {
  features {}
}

terraform {
  backend "azurerm" {
    resource_group_name  = "tfstate-rg"
    storage_account_name = "tfstatesa0"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
    client_id            = "c8a9db4c-9521-4e03-9080-f12e2fac2e8e"
    tenant_id            = "3e27058d-8cf2-4ff8-8047-9a678663b82b"
    subscription_id      = "45f9ddda-3ab0-4e12-a0cc-03e6ed2b1fa1"
  }
}