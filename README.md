# Domino Fleet Manager

[![HCL Domino](https://img.shields.io/badge/HCL-Domino-ffde21?logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2aWV3Qm94PSIwIDAgNzE0LjMzIDcxNC4zMyI+PGRlZnM+PHN0eWxlPi5jbHMtMXtmaWxsOiM5M2EyYWQ7fS5jbHMtMntmaWxsOnVybCgjbGluZWFyLWdyYWRpZW50KTt9PC9zdHlsZT48bGluZWFyR3JhZGllbnQgaWQ9ImxpbmVhci1ncmFkaWVudCIgeDE9Ii0xMjA3LjIiIHkxPSItMTQzIiB4Mj0iLTEwMzguNjYiIHkyPSItMTQzIiBncmFkaWVudFRyYW5zZm9ybT0ibWF0cml4KDEuMDYsIDAuMTMsIC0wLjExLCAwLjk5LCAxMzUzLjcsIDYwMC42MikiIGdyYWRpZW50VW5pdHM9InVzZXJTcGFjZU9uVXNlIj48c3RvcCBvZmZzZXQ9IjAiIHN0b3AtY29sb3I9IiNmZmRmNDEiLz48c3RvcCBvZmZzZXQ9IjAuMjYiIHN0b3AtY29sb3I9IiNmZWRjM2QiLz48c3RvcCBvZmZzZXQ9IjAuNSIgc3RvcC1jb2xvcj0iI2ZiZDIzMiIvPjxzdG9wIG9mZnNldD0iMC43NCIgc3RvcC1jb2xvcj0iI2Y2YzExZiIvPjxzdG9wIG9mZnNldD0iMC45NyIgc3RvcC1jb2xvcj0iI2VmYWEwNCIvPjxzdG9wIG9mZnNldD0iMSIgc3RvcC1jb2xvcj0iI2VlYTYwMCIvPjwvbGluZWFyR3JhZGllbnQ+PC9kZWZzPjxnIGlkPSJMYXllcl8zIiBkYXRhLW5hbWU9IkxheWVyIDMiPjxwb2x5Z29uIGNsYXNzPSJjbHMtMSIgcG9pbnRzPSI0MzcuNDYgMjgzLjI4IDMzNi40NiA1MDYuNjkgMjExLjY4IDUwNy40NSAzNjYuOTIgMTYyLjYxIDQzNy40NiAyODMuMjgiLz48cG9seWdvbiBjbGFzcz0iY2xzLTEiIHBvaW50cz0iNjQwLjU5IDMwNC4xIDUyOS4wMiA1NTEuOTYgMzUzLjYzIDU2Ni42MiA1NDIuMzIgMTQ3LjcxIDY0MC41OSAzMDQuMSIvPjxwb2x5Z29uIGNsYXNzPSJjbHMtMiIgcG9pbnRzPSIyNzMuMTkgMjY1LjM3IDE5MC4xMSA0NTAuMDYgNzMuNzQgNDM5LjI4IDE5NC4zMiAxNzEuMzMgMjczLjE5IDI2NS4zNyIvPjwvZz48L3N2Zz4K
)](https://www.hcl-software.com/domino)
[![License: Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/nashcom/buil-test/blob/main/LICENSE)

## Introduction

This project contains resources to manage Domino running in containers [Docker](https://www.docker.com/), [Podman](https://podman.io/) and [Kubernetes](https://kubernetes.io/).
It containers resources to install Domino on Docker and Kubernetes (K8s) and helps to manager Domino servers in containers.
The main focus is on K8s because Domino can benefit more from managing containers automatically on this platform.
But the project is also intended to provide resources and references for running Docker and Podman containers.

The project is in an initial state. It is public to simplify communication with very early adopters.

## Domino Fleet Manager (DFM)

**DFM** is a Notes database (`dfm.nsf`) which provides the following functionality:

- Scan Domino servers and provide an overview and management interface for Domino servers
- Support for creating new servers leveraging the Domino CA
- Server configuration management leveraging Domino One Touch Setup (OTS)
- Domain scope, Cluster scope and Server scope operations
- Server Group management
- Distribute **notes.ini** variables
- Server capacity management


## Cube Control

The Cube Control applications is implemented as a container image and provides an interface between the Domino Fleet Manager application (`dfm.nsf`) and K8s.
It provides a locked down and secure interface to the K8s API using a very streamlined  REST API to apply changes to scoped K8s resources.


# Resources

The project will also provide resources for installing and managing K8s.

## RKE2 Resources

[RKE2](https://rke2.io/) is a free to use K8s implementation with has support options available.

## k3s Resources

[k3s](https://k3s.io/) is a very lightweight K8s implementation which is specially interesting for smaller environments and test environments.
For production use RKE2 is recommended.

## Rancher

[Rancher](https://www.rancher.com/) is an enterprise level multi K8s environment management, which is very convenient for administration and is also available free of charge.

