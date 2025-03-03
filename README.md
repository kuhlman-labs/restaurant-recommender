# Restaurant Recommender

Restaurant Recommender is a Go-based web service that you can use to get restaurant recommendations based on various criteria. The service is designed to be deployed on Azure and uses a SQL database for data storage. It leverages Flyway for database migrations and GitHub Actions for CI/CD.

## Features

- **Service Principal Authentication:** The application uses a Service Principal to authenticate with Azure SQL Database.
- **Database Migrations:** Flyway is used to manage your database schema and seed data.
- **Containerized Deployment:** The web app is packaged as a Docker container and deployed to an Azure Linux Web App.
- **CI/CD Pipeline:** GitHub Actions workflows handle continuous integration (CI) and continuous deployment (CD) along with infrastructure provisioning via Terraform.

## Prerequisites

- [Go](https://golang.org/dl/) (version 1.24+ recommended)
- [Terraform](https://www.terraform.io/downloads) (if you need to provision the Azure infrastructure)
- An [Azure Subscription](https://azure.microsoft.com/free/) with appropriate permissions
- GitHub repository secrets for Azure credentials (see below)

### Seetting up Secrets in your GitHub Repository
Ensure you set up the necessary GitHub Action secrets in your repository:
- `ARM_CLIENT_ID`: The client ID of the Managed Identity.
- `ARM_CLIENT_SECRET`: The client secret of the Managed Identity.
- `ARM_TENANT_ID`: The tenant ID of your Azure subscription.
- `ARM_SUBSCRIPTION_ID`: The subscription ID of your Azure subscription.
- `AZURE_CREDENTIALS`: The JSON credentials for the Azure service principal.
- `ACR_NAME`: The name of your Azure Container Registry.
- `WEBAPP_NAME`: The name of your Azure Web App.

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


## Deployment with GitHub Actions

The project uses GitHub Actions workflows for CI/CD:

- **cd.yaml** handles building the Docker image, pushing it to Azure Container Registry, and deploying to Azure Web App.
- **ci.yaml** runs tests and builds the Go application.
- **terraform.yaml** provisions infrastructure via Terraform and exports outputs (like the SQL server FQDN and DB name) as environment variables for use by the deployment steps.

The CD workflows are triggered on push to the `main` branch and on the CI workflows are tiggered on pull requests.

## Query the API
Once deployed, you can query the API like so:

```bash
curl -X GET "https://<webapp-name>.azurewebsites.net/recommend?query=A vegetarian Italian restaurant that is open at 10am"
```
This will return a JSON response with restaurant recommendations based on the query.

```bash
{
  "restaurantRecommendation": {
    "name": "Pizza Hut",
    "style": "Italian",
    "address": "Wherever Street 99, Somewhere",
    "openHour": "09:00",
    "closeHour": "23:00",
    "vegetarian": true,
    "deliveries": true
  }
}
```

## Contributing

Feel free to submit issues or pull requests. For major changes, please open an issue first to discuss what you would like to change.

## License
This project is licensed under the MIT License. You are free to use, modify, and distribute this code as long as you include the original license in your distribution. See [LICENSE](LICENSE) for details.
