#!/usr/bin/env bash

#Create docker container
docker run  --ulimit nofile=1024:1024 --name swaggerapi-petstore3 -d -p 8080:8080 swaggerapi/petstore3:latest