# !/bin/bash
#
# Copyright (C) 2018 Nalej Group - All Rights Reserved
#
# Helper script to create docker credentials to access DockerHub images from private registry in DockerHub.
#
# Set environment variables with the credentials regarding your DockerHub account.
# export DOCKER_REGISTRY_SERVER=https://index.docker.io/v1/
# export DOCKER_USER=Type your dockerhub username, same as when you `docker login`
# export DOCKER_EMAIL=Type your dockerhub email, same as when you `docker login`
# export DOCKER_PASSWORD=Type your dockerhub pw, same as when you `docker login`

kubectl create secret docker-registry myregistrykey \
  --docker-server=$DOCKER_REGISTRY_SERVER \
  --docker-username=$DOCKER_USER \
  --docker-password=$DOCKER_PASSWORD \
  --docker-email=$DOCKER_EMAIL