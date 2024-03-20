# Chat Group Server

## Description
This is the backend service for the Chat Gropu project and it provides every feature needed by the project like real time messaging, authentication, channels creation, etc. In this repository is also included the Database which is a postgres service running on Docker

## Build
To run this project you need to have Docker installed on your machine: https://www.docker.com/get-started/

Build the Go server image
```bash
docker compose build
```

Then run both the server and the database
```bash
docker compose up
```

The API should be visible on port 4000.