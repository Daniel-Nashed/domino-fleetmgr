
# RKE2 Setup

The following is a short summary of useful commands and not a full instruction yet.
The directory contains helper scripts to help setting up RKE2.
But thise files are not a complete setup yet.

## Add DNS entries

```
kubectl -n kube-system edit configmap rke2-coredns-rke2-coredns
```


```
  template IN ANY example.com {
    match ^(dom-[a-z0-9-]+|nomad|rancher|auth|mail|domino-smtp|domino-ldap)\..*$
    answer "{{.Name}} 30 IN CNAME domino-https-lb.domino.svc.cluster.local."
  }


```


With support for MX records, SMTP service and LDAP service the entries would be split

```
        template IN A AAAA CNAME example.com {
        match ^domino-smtp\.k8s\.example\.com\.$
          answer "{{ .Name }} 300 IN CNAME domino-smtp.domino.svc.cluster.local."
          fallthrough
        }
        template IN A AAAA CNAME example.com {
        match ^domino-ldap\.k8s\.example\.com\.$
          answer "{{ .Name }} 300 IN CNAME domino-ldap.domino.svc.cluster.local."
          fallthrough
        }
        template IN A AAAA CNAME example.com {
        match ^(dom-[a-z0-9-]+|nomad|files|rancher|auth|mail|domino-smtp|domino-ldap)\..*$
          answer "{{ .Name }} 300 IN CNAME domino-https-lb.domino.svc.cluster.local."
        }
        template IN MX example.com {
          match example.com
          answer "{{ .Zone }} 300 IN MX 10 domino-smtp.domino.svc.cluster.local."
        }

```


```
kubectl -n kube-system rollout restart deploy rke2-coredns-rke2-coredns 
```

```
kubectl -n kube-system rollout status deploy rke2-coredns-rke2-coredns
```

```
kubectl -n kube-system logs -l k8s-app=kube-dns --tail=200
```

## Uninstall NGINX Ingress controller


```
helm -n kube-system uninstall rke2-ingress-nginx
```



## Install Longhorn required packages 

apt-get install open-iscsi
apt-get install nfs-common

yum --setopt=tsflags=noscripts install iscsi-initiator-utils
echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
systemctl enable iscsid
systemctl start iscsid

yum install nfs-utils


modprobe iscsi_tcp
