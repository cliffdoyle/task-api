
### Step 2: Azure Resources Setup Script

**scripts/setup-azure.sh**
```bash
#!/bin/bash

# Variables
RESOURCE_GROUP="task-api-rg"
LOCATION="eastus"
ACR_NAME="taskapiregistry"
APP_SERVICE_PLAN="task-api-plan"
WEB_APP_NAME="task-api-app"
SQL_SERVER="task-api-sql"
SQL_DB="taskapi"

# Create Resource Group
az group create --name $RESOURCE_GROUP --location $LOCATION

# Create Azure Container Registry
az acr create --resource-group $RESOURCE_GROUP \
  --name $ACR_NAME --sku Basic --admin-enabled true

# Create App Service Plan (Linux)
az appservice plan create --name $APP_SERVICE_PLAN \
  --resource-group $RESOURCE_GROUP \
  --sku B1 --is-linux

# Create Web App
az webapp create --resource-group $RESOURCE_GROUP \
  --plan $APP_SERVICE_PLAN --name $WEB_APP_NAME \
  --deployment-container-image-name $ACR_NAME.azurecr.io/task-api:latest

# Create Azure SQL Server
az sql server create --name $SQL_SERVER \
  --resource-group $RESOURCE_GROUP \
  --location $LOCATION \
  --admin-user sqladmin --admin-password 'YourStrongPassword123!'

# Create SQL Database
az sql db create --resource-group $RESOURCE_GROUP \
  --server $SQL_SERVER --name $SQL_DB \
  --service-objective S0

# Configure firewall rules
az sql server firewall-rule create --resource-group $RESOURCE_GROUP \
  --server $SQL_SERVER --name AllowAzureServices \
  --start-ip-address 0.0.0.0 --end-ip-address 0.0.0.0

# Configure Web App settings
az webapp config appsettings set --resource-group $RESOURCE_GROUP \
  --name $WEB_APP_NAME \
  --settings DATABASE_URL="postgresql://$SQL_SERVER.database.windows.net:5432/$SQL_DB?user=sqladmin&password=YourStrongPassword123!&sslmode=require"

echo "Azure resources created successfully!"
```