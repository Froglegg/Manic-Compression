# THESE ARE THE STEPS TAKEN FOR SETTING UP OUR WEB APP, DON'T RUN THIS SCRIPT UNLESS WE ARE RESETTING EVERYTHING!!!
# # create appservice plan
# az appservice plan create --name manicCompressionServicePlan --resource-group team8 --sku B1 --is-linux

# # create app
# az webapp create --resource-group team8 --plan manicCompressionServicePlan --name manic-compression --multicontainer-config-type compose --multicontainer-config-file ./docker/docker-compose.prod.yml

# # assign managed identity to web app
# az webapp identity assign --name manic-compression --resource-group team8

# # get principalId of web app
# principalId=$(az webapp identity show --name manic-compression --resource-group team8 --query principalId --output tsv)

# # grant web app access to ACR
# az role assignment create --assignee $principalId --role AcrPull --scope /subscriptions/46b7b536-4e35-4e9d-b1b4-7ef9ab5fb413/resourceGroups/team8/providers/Microsoft.ContainerRegistry/registries/sdcc2023team8registry

# # set environment variables
# az webapp config appsettings set --name manic-compression --resource-group team8 --settings AZURE_STORAGE_CONNECTION_STRING=$AZURE_STORAGE_CONNECTION_STRING
