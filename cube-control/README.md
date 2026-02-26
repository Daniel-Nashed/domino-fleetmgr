# Cube Control


Cube Control provides a local REST API interface to bridge between **kubectl** and the Domino Fleet Manager components.
Domino Fleet Manager uses Lotus Script REST API requests to apply **YAML** files to manage Domino Kubernetes resources.
                                                                                                  
Cube Control listens on port 8443 by default and provides an `/apply` endpoint.
The endpoint is protected using a Bearer token, which can be provided in a secret or via an environment variable.


## Environment variables



| Variable                          | Description                                                                 | Default                                              |
|:----------------------------------|:----------------------------------------------------------------------------|:-----------------------------------------------------|
| CUBE_CONTROL_LISTEN_ADDR          | Listen address for HTTP requests. The application does not support HTTPS. NGINX can be placed in front using a separate container. | :8443                                               |
| CUBE_CONTROL_SERVER_NAME          | Server name                                                                 | cube-control.domino.svc.cluster.local               |
| CUBE_CONTROL_CERTMGR_SERVER       | CertMgr server to connect to when checking for certificate updates         | —                                                   |
| CUBE_CONTROL_TOKEN_FILE           | File name to read the authentication token from                            | /var/run/secrets/cube-control/token                 |
| CUBE_CONTROL_TOKEN                | Authentication token stored in environment variable (overrides token file) | —                                                   |
| CUBE_CONTROL_CFG_CHECK_INTERVAL   | Certificate and token update check interval                                | 120s                                                |

## Technology used

The container image uses Alpine as the base image.
A small GO application is handling the requests and passes them to **kubectl**.
The latest stable version of **kubectl** version is downloaded from the Kubernetes GitHub repository and installed into the container.


## TLS Only configuration

Cube Control supports TLS 1.2 and TLS 1.3 only.
No unencrypted port is available.

The key and certificate must be provided in the following location.
Usually by mounting a TLS secret at `/tls`

- /tls/tls.crt
- /tls/tls.key

The application waits until the files are available before starting the TLS listener.


## How to build the container image

Run the following command to build the image based on the latest Alpine base image:

```
./build.sh
```
