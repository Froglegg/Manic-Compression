# login to azure if you haven't already
az login
az acr login --name sdcc2023team8registry

# build go binaries for linux Alpine (go server, etc.)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/manic-server ../web/manic-server/...

# build react app
echo "Installing npm dependencies for frontend React app"
npm --prefix ../web/manic-client/ install
echo "Building frontend React app"
npm run --prefix ../web/manic-client/ build

# build docker images
docker compose -f ../docker/docker-compose.yml build

# build, tag and push images to ACR
docker tag manic-server:latest sdcc2023team8registry.azurecr.io/manic-server:latest
docker push sdcc2023team8registry.azurecr.io/manic-server:latest
docker tag manic-client:latest sdcc2023team8registry.azurecr.io/manic-client:latest
docker push sdcc2023team8registry.azurecr.io/manic-client:latest

# restart the web app
az webapp restart --name manic-compression --resource-group team8
az webapp show --name manic-compression --resource-group team8 --query state

echo "Deployment complete!"
echo "Navigate to http://manic-compression.azurewebsites.net/ to view the app."
echo "Navigate to https://manic-compression.scm.azurewebsites.net/ to view the app service logs."