# infrastructure-manager

The infrastructure manager is in charge of adding and removing infrastructure.

## Getting Started

The component will use the `installer`, `provisioner` and `system-model` to orchestrate the business logic behind infrastructure operations.

### Prerequisites

* [system-model](https://github.com/nalej/system-model)
* [installer](https://github.com/nalej/installer)
* [provisioner](https://github.com/nalej/provisioner)
* [nalej-bus](https://github.com/nalej/nalej-bus)

### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, downloads the required dependencies, runs existing tests and generates ready-to-deploy Kubernetes files.

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

## Known issues

* The monitoring system of ongoing requests will not be able to continue in the event of a failure of the
`infrastructure-manager`. A refactor where messages are sent to the bus by the `provisioner` and `installer` components
and consumed by the `infrastructure-manager` would increase the reliability and scalability of the system. This
refactor is planned for future versions of the platform (NP-2429).

## Contributing

Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the available versions, see the [tags on this repository](https://github.com/nalej/infrastructure-manager/tags). 

## Authors

See also the list of [contributors](https://github.com/nalej/infrastructure-manager/contributors) who participated in this project.

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.
