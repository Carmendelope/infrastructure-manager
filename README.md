# infrastructure-manager

This repository contains the infrastructure manager in charge of the operations related to adding and removing infrastructure.

## Getting Started

The component will leverage the `installer`, `provisioner` and `system-model` to orchestrate the business logic behind infrastructure operations.

### Prerequisites

Detail any component that has to be installed to run this component.

* system-model
* installer
* provisioner
* nalej-bus

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, download dependencies,
run existing tests and generate ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```

## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/nalej/infrastructure-manager/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/infrastructure-manager/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.