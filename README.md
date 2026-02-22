# domino-fleetmgr - Domino Fleet Manager

## Introduction

This project contains resources to manage Domino running in containers (Docker & Kubernetes).
It containers resources to install Domino on Docker and Kubernetes (K8s) and helps to manager Domino servers in containers.
The main focus is on K8s because Domino can benefit more from managing containers automatically on this platform.
But the project is also intended to provide resources and references for running Docker/Podman containers.


## Domino Fleet Manager (DFM)

DFM is a Notes database which provides the following functionality

- Scan Domino servers and provide an overview and management interface for Domino servers
- Support for creating new servers leveraging the Domino CA
- Server configuration management leveraging Domino One Touch Setup (OTS)
- Domain scope, Cluster scope and Server scope operations
- Server Group management
- Distribute Notes.ini variables
- Server capacity management


## Cube Control

The container image provides the interface between the Domino Fleet Manager application (dfm.nsf) and K8s.


# Resources

## RKE2 Resources

[RKE2](https://rke2.io/) is a free to use K8s implementation with has support options available.

## k3s Resources

[k3s](https://k3s.io/) is a very lightweight K8s implementation which is specially itneresting for smaller environments and test environments.
For production use RKE2 is recommended.

## Rancher

[Rancher](https://www.rancher.com/) is an enterprise level multi K8s environment management, which is very convenient for administration and is also available free of charge.

