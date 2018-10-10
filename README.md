# golang-template

This repository contains a basic template for new Nalej golang-based projects. 

## Guidelines

This repository follows the project structure recommended by the [golang community] (https://github.com/golang-standards/project-layout).
Please refer to the previous URL for more information. We describe some relevant folders:

* cmd: Applications contained in this repo. This folder contains the minimal expression of executable applications.
This means that code such as algorithms, servers, database queries MUST not be there.
* components: Definition of components, there corresponding Docker images and regarding files if needed. Normally, Docker
images will simply incorporate the main Golang compiled file running an app.
* pkg: this folder will contain most of the code supporting the applications.
* scripts: relevent scripts needed to run applications/configure them or simply helping.
* util: any code that without being really part of an application is useful in some context.

And some relevant files:

* Gopkg.toml: dependencies needed to run must be indicated here, specially if a certain version of a repo is required.
* Makefile: the main compilation manager for projects. More information below.
* Readme.md: (Actually this file) Description of the project and examples of use.
* .version: One single line indicating the current version of the repo. For example: v0.0.1, v1.2.3, etc.

## Makefile

Use the `make` command to control the generation of executables, management of dependencies, testing and Docker
images. Minor modifications have to be done in order to adapt a project with the structure defined in this document
to be compiled using the current Makefile.

### Set the applications

In order to know what applications have to be generated, the list of applications have to be set. This is a blank
space separated list. The name of the apps must be the same of the folders under the cmd. For the current example,
example-app and other-app.

```bash
# Name of the target applications to be built
APPS=example-app other-app
```

## Set the version

Any developed solution will be associated with a version. The value of the version must be indicated in the version file.

## All at once

Running `make all` will execute the most relevant tasks except those regarding Docker management. The resulting files
can be found at the bin folder.

For more information about particular aspects of the compilation process, please check the following sections.

## Configure dependencies

We use godep for the management of dependencies. To automatically update your dependencies and download the required
packages into the vendor folder run `make dep`.

## Build the apps

The `make build` command generates executables compatible with your current OS. For linux-compatible executables run
`make build-linux`. Finally, `make build-all` will run both. The resulting executable files are available under the
bin folder.

```bash
jtmartin-mbp:golang-template juan$ ./bin/other-app
{"level":"info","time":"2018-09-07T12:42:33+02:00","message":"You're running other example application."}
          ----------
         |  --------  |
         | |########| |       __________
         | |########| |      /__________\
 --------|  --------  |------|    --=-- |-------------
|         ----,-,-----'      |  ======  |             |
|       ______|_|_______     |__________|             |
|      /  %%%%%%%%%%%%  \                             |
|     /  %%%%%%%%%%%%%%  \                            |
|     ^^^^^^^^^^^^^^^^^^^^                            |
+-----------------------------------------------------+
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
```

## Testing

Testing runs all the available go tests for the current project. A `make test-all` is available to run just at once
regular, coverage and race testing. Run `make test`, `make test-coverage` and `make test-race` to run regular,
coverage and race tests accordingly. Any outcome is stored under the bin folder.

## Building Docker images

The `make image` option will build all the available Docker images if and only if a Dockerfile is found for the
application in the components folder. For example, in the current example only the Docker image for example-app is
generated. Images are tagged using the version value set in the version file. The resulting image.tar file can be
found in the bin folder under the folder with the same application name.

```bash
bin
├── images
│   └── example-app
│       ├── example-app.tar.gz
│       └── image.tar
└── linux_amd64
    ├── example-app
    └── other-app
```

## Publishing Docker images

Publishing images into the Docker Hub requires a user account. Due to security restrictions the user id and password
will be required when executing the `make publish` option. Both entries are expected to be available as environment
variables. To set the variables run:

```bash
# Set environment variables with the credentials regarding your DockerHub account.
export DOCKER_REGISTRY_SERVER=https://index.docker.io/v1/
export DOCKER_USER=Type your dockerhub username, same as when you `docker login`
export DOCKER_EMAIL=Type your dockerhub email, same as when you `docker login`
export DOCKER_PASSWORD=Type your dockerhub pw, same as when you `docker login`
```
The final step for publishing images will not run if any of the variables is not set.

 Both, user id and password has to be typed. Docker images
are published under the nalej project with

```bash
jtmartin-mbp:golang-template juan$ make publish
>>> Updating dependencies...
if [ ! -d vendor ]; then \
            echo ">>> Create vendor folder" ; \
            mkdir vendor ; \
        fi ;
dep ensure -v
Gopkg.lock was already in sync with imports and Gopkg.toml
(1/4) Wrote github.com/inconshreveable/mousetrap@v1.0
(2/4) Wrote github.com/spf13/pflag@v1.0.2
(3/4) Wrote github.com/rs/zerolog@v1.8.0
(4/4) Wrote github.com/spf13/cobra@v0.0.3
>>> Bulding for Linux...
for app in example-app other-app; do \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux_amd64/"$app" ./cmd/"$app" ; \
        done
mkdir -p bin/images
>>> Creating images ...
for app in example-app other-app; do \
        echo Create image of app $app ; \
        if [ -f components/"$app"/Dockerfile ]; then \
            mkdir -p bin/images/"$app" ; \
            docker build --no-cache -t juannalej/"$app":v0.0.1 -f components/"$app"/Dockerfile bin/linux_amd64 ; \
            docker save juannalej/"$app" > bin/images/"$app"/image.tar ; \
            // docker rmi juannalej/"$app":v0.0.1 ; \
            cd bin/images/"$app"/ && tar cvzf "$app".tar.gz * && cd - ; \
        else  \
            echo $app has no Dockerfile ; \
        fi ; \
    done
Create image of app example-app
Sending build context to Docker daemon  8.963MB
Step 1/4 : FROM iron/go
 ---> ed7df0451f6c
Step 2/4 : RUN mkdir /nalej
 ---> Running in bcce6861c1ee
Removing intermediate container bcce6861c1ee
 ---> 55427806182f
Step 3/4 : COPY example-app /nalej/
 ---> d088bed960b5
Step 4/4 : ENTRYPOINT ["./nalej/example-app"]
 ---> Running in e09120d30b78
Removing intermediate container e09120d30b78
 ---> 9b5bc7959ad5
Successfully built 9b5bc7959ad5
Successfully tagged juannalej/example-app:v0.0.1
/bin/sh: //: is a directory
a example-app.tar.gz: Can't add archive to itself
a image.tar
/Users/juan/nalej_workspace/src/github.com/nalej/golang-template
Create image of app other-app
other-app has no Dockerfile
>>> Publish images into Docker Hub ...
>>> Assuming credentials are available in environment variables ...
if [ ""$DOCKER_USER"" = "" ]; then \
            echo DOCKER_USER environment variable was not set!!! ; \
            exit 1 ; \
        fi ; \
        if [ ""$DOCKER_USER"" = "" ]; then \
        echo DOCKER_USER environment variable was not set!!! ; \
        exit 1 ; \
    fi ; \

echo  "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USER" --password-stdin
Login Succeeded
for app in example-app other-app; do \
            if [ -f bin/images/"$app"/image.tar ]; then \
                docker push juannalej/"$app":v0.0.1 ; \
            else \
                echo $app has no image to be pushed ; \
            fi ; \
            echo  Publish image of app $app ; \
    done ; \
    docker logout ; \

The push refers to repository [docker.io/juannalej/example-app]
f2e62fde1a0a: Pushed
062d8b81e1c4: Pushed
ab3cb40727df: Pushed
c9e8b5c053a2: Pushed
v0.0.1: digest: sha256:19bdb52e27166a930519c99fd6b9ce370e1353a27aa3bdf1efed1b4aef5e0571 size: 1155
Publish image of app example-app
other-app has no image to be pushed
Publish image of app other-app
Removing login credentials for https://index.docker.io/v1/
```

## Working with K8s and DockerHub

Because we are using a private DockerHub repository for Nalej images, the image pulling process requires a
previous authentication process. The scripts folder contains a credentials generation example that uses
local environment variables to set a docker registry secret.

```bash
export DOCKER_REGISTRY_SERVER=https://index.docker.io/v1/
export DOCKER_USER=Type your dockerhub username, same as when you `docker login`
export DOCKER_EMAIL=Type your dockerhub email, same as when you `docker login`
export DOCKER_PASSWORD=Type your dockerhub pw, same as when you `docker login`

kubectl create secret docker-registry myregistrykey \
  --docker-server=$DOCKER_REGISTRY_SERVER \
  --docker-username=$DOCKER_USER \
  --docker-password=$DOCKER_PASSWORD \
  --docker-email=$DOCKER_EMAIL
```

The code above generates a token authentication for DockerHub that is stored in the corresponding secret. Now K8s
deployment files can be set to use this secret for the next images.

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  labels:
    app: application
    component: example
  name: example
  namespace: default
spec:
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: application
      component: example
  template:
    metadata:
      labels:
        app: application
        component: example
    spec:
      containers:
      - name: example
        image: nalej/example-app:v0.0.1
        imagePullPolicy: Always
        securityContext:
          runAsUser: 2000
      imagePullSecrets:
      - name: myregistrykey
```