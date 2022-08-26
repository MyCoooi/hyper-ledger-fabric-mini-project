#!/bin/bash

docker rm -f $(docker ps -aq)
docker rmi -f $(docker images dev-* -q)
docker network fabric_project