# infrastructure-manager

This repository contains the infrastructure manager in charge of the operations related to adding and removing infrastructure.
The component will leverage the Installer, Provisioner and System Model to orchestrate the business logic behing infrastructure operations.

## Launching the manager

```
$ ./bin/infrastructure-manager run --debug --consoleLogging
```

# Integration tests

The following table contains the variables that activate the integration tests

| Variable  | Example Value | Description |
| ------------- | ------------- |------------- |
| RUN_INTEGRATION_TEST  | true | Run integration tests |
| IT_SM_ADDRESS  | localhost:8800 | System Model Address |
| IT_PROVISIONER_ADDRESS  | localhost:8930 | Provisioner Address |
| IT_INSTALLER_ADDRESS  | localhost:8900 | Installer Address |
| IT_K8S_KUBECONFIG | /Users/daniel/.kube/config| KubeConfig for the minikube credentials |