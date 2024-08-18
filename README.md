# CME Chat Project

## Overview

This project is a Go-based web application designed to interact with Cassandra and Redis. The application is containerized using Docker and orchestrated with Docker Compose for easy deployment and management of services.

## Accessing the Application

Once the application is running, you can access it at http://localhost:9090.

## Stopping the Application

To stop the application and its services, run: `make stop`

## Cleaning Up
To remove all containers, networks, and volumes created by Docker Compose: `make clean`

## Deployment
1.  Building the Docker Image
    - To build the Docker image manually, run: `make build`
2. Running with Docker Compose
    - To run the application and its dependencies: `make run`
3. Stopping and Removing Containers
    - To stop and remove all containers: `make clean`

## Logs
To view the logs of the running services: `make logs`
