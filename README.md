# Manic Compression
Manic Compression uses an event driven web framework for applying audio effects in a chain using Azure durable functions, service bus, and app service in a multiple container deployment.

## Local Development
### Setup
1. Ensure that your AZURE_STORAGE_CONNECTION_STRING is set in your local env
2. Ensure that you're have npm version >= 16.0 and that you have run `npm i` in web/manic-client

### Running locally
1. Run `npm run start` in web/manic-client. This should open a browser window at localhost:3000. You should see "attempting to connect to server..."
2. Open a separate terminal window and cd into web/manic-server and run `go run manic-server`
3. Navigate back to the react app, you should now see "hello from manic compression server!"

### Notes
All local requests in the React (npm run start) development environment should proxy out to localhost:8080, where the local server will be listening. In production, the nginx.conf will proxy calls to /api to the manic-server. See package.json & nginx.conf in web/

## Deployment
You can build and deploy the application by running `./scripts/deploy.sh`
Navigate to http://manic-compression.azurewebsites.net/ to view the app.
Navigate to https://manic-compression.scm.azurewebsites.net/ to view the app service logs# manic_compression
