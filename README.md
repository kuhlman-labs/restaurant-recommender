# Restaurant Recommender

Restaurant Recommender is a Go-based web service that connects to an Azure SQL database using Managed Identity. It uses Flyway for database migrations, including schema creation and data seeding, and is deployed via GitHub Actions and Terraform.

## Features

- **Managed Identity Authentication:** Uses Azure Managed Identity to securely connect to Azure SQL without storing credentials.
- **Database Migrations:** Flyway is used to manage your database schema and seed data.
- **Containerized Deployment:** The web app is packaged as a Docker container and deployed to an Azure Linux Web App.
- **CI/CD Pipeline:** GitHub Actions workflows handle continuous integration (CI) and continuous deployment (CD) along with infrastructure provisioning via Terraform.

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.24+ recommended)
- [Terraform](https://www.terraform.io/downloads) (if you need to provision the Azure infrastructure)
- An [Azure Subscription](https://azure.microsoft.com/free/) with appropriate permissions
- GitHub repository secrets for Azure credentials (see below)

## Database Migrations

Database schema creation and data seeding are managed via Flyway migrations:

- Place your migration SQL files in the `db/migrations` directory.
- Versioned migrations should follow the pattern `V{version}__{Description}.sql`
- Repeatable migrations should begin with `R__`.

Flyway runs these migrations automatically (e.g. via a GitHub Actions workflow) after infrastructure provisioning.

## Infrastructure Provisioning with Terraform

Terraform is used to provision Azure resources such as:
- Resource Groups
- Azure SQL Server & Database
- Azure Container Registry
- Linux Web App (with System-assigned Managed Identity)
- Role assignments (for ACR pull and DB access)

Ensure you set up the necessary GitHub secrets in your repository (e.g. `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID`, and `AZURE_CREDENTIALS` formatted as JSON).


## Deployment with GitHub Actions

The project uses GitHub Actions workflows for CI/CD:

- **cd.yaml** handles building the Docker image, pushing it to Azure Container Registry, and deploying to Azure Web App.
- **ci.yaml** runs tests and builds the Go application.
- **terraform.yaml** provisions infrastructure via Terraform and exports outputs (like the SQL server FQDN and DB name) as environment variables for use by the deployment steps.


## Troubleshooting

- **Managed Identity Issues:**  
  Verify that your Azure Web App or service (when deployed) has been assigned the proper permissions to access the SQL database.

## Contributing

Feel free to submit issues or pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License
This project is licensed under the MIT License. You are free to use, modify, and distribute this code as long as you include the original license in your distribution. See [LICENSE](LICENSE) for details.
