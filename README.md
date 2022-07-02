# Certificates Discovery

## Description

A container for Kubernetes that can get TLS certificates from several hosts and craft Kubernetes Secret resources with them

## Motivation

Some deployments in Kubernetes require getting TLS certificates from Secret resources, even when certificates could be
obtained from the TLS handshake.

This happened to us, for example, **when deploying Kafka MirrorMaker2 with Strimzi Operator**: the `spec` of the MM2 CR
for TLS connections, forces to reference a Secret with the certificate of the remote Kafka already present on the cluster.
This situation would require manual intervention even when the Certificates for Kafka are fully automated. Due to this
is totally unneeded, this tool exists to cover the corner case

## How to develop

> We recommend you to use a development tool like [Kind](https://kind.sigs.k8s.io/) or [Minikube](https://minikube.sigs.k8s.io/docs/start/)
> to launch a lightweight Kubernetes on your local machine for development purposes

For learning purposes, we will suppose you are going to use Kind. So the first step is to create a Kubernetes cluster
on your local machine executing the following command:

```console
kind create cluster
```

Once you have launched a safe play place, execute the following command. It will install the custom resource definitions
(CRDs) in the cluster configured in your ~/.kube/config file and run the Operator locally against the cluster:

```console
make install run
```

> Remember that your `kubectl` is pointing to your Kind cluster. However, you should always review the context your
> kubectl CLI is pointing to

## How releases are created

Each release of this operator is done following several steps carefully in order not to break the things for anyone.
Reliability is important to us, so we automated all the process of launching a release. For a better understanding of
the process, the steps are described in the following recipe:

1. Test the changes on the code:

    ```console
    make test
    ```

   > A release is not done if this stage fails

2. Define the package information

    ```console
    export VERSION="0.0.1"
    ```

3. Generate and push the Docker image (published on Docker Hub).

    ```console
    make docker-build docker-push
    ```

## Flags

Most features coded in this project are feature flags, so in the following table you have a better insight

| Flag                | Description                                                               | Default          |
|---------------------|:--------------------------------------------------------------------------|:-----------------|
| `--connection-mode` | How to connect to Kubernetes: `incluster` or `kubectl`                    | `kubectl`        |
| `--kubeconfig`      | Absolute path to a kubeconfig file if connection mode is `kubectl`        | `~/.kube/config` |
| `--namespace`       | Namespace where to create and synchronize Kubernetes Secrets with TLS     | `default`        |
| `--secret-name`     | (Repeatable) The name of the Secret to create with the TLS certificate    | -                |
| `--tls-host`        | (Repeatable) The TLS host (with port) where to obtain the TLS certificate | -                |

> The flags `--secret-name` and `--tls-host` are linked and mandatory: each `secret name` must be followed by a `tls host`

## How to use

We have prepared several example manifests for you to make it easy to deploy everything. Use them as a starting point
and patch all you need. You should not need to apply a lot of patches. If you need to patch everything, this deployment
is not suitable for your needs.

All the examples can be found on the directory [examples](/examples) in this repository. Just download and use them as yours

## How to collaborate

If you find some bugs, or would like to collaborate, just open an issue explaining the problem, then fork the repository,
make all changes you need on a branch and open a PR against this one. Maintainers will be review and test the changes in
a real cluster carefully before merging the code.
