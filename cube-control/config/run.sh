docker run --rm -it --name cube-control -p 80:8080 -e CUBE_CONTROL_TOKEN=BINGO -e KUBECONFIG=/kubeconfig -v ./config.yaml:/kubeconfig:ro -v ./job.yml:/job.yml:ro cube-control bash
