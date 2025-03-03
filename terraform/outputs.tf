output "sql_server_fqdn" {
  description = "The fully qualified domain name of the Azure SQL Server."
  value       = azurerm_mssql_server.sqlserver.fully_qualified_domain_name
}

output "sql_database_name" {
  description = "The name of the Azure SQL Database."
  value       = azurerm_mssql_database.sqldb.name
}

output "webapp_managed_identity" {
  description = "The principal id of the web app's system-assigned managed identity."
  value       = azurerm_linux_web_app.webapp.identity[0].principal_id
}