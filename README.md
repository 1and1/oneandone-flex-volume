# oneandone-flex-volume
1&amp;1 Flex Volume Plugin for Kubernetes


1. Get the cluster up and running. 
2. Deploy [oneandone-flex-provisioner](https://github.com/StackPointCloud/external-storage/blob/oneandone/oneandone/flex-volume/README.md)
3. Build the binaries:
```
$ make

==> Checking that code complies with gofmt requirements...
make: execvp: /Volumes/Data/go/src/github.com/1and1/oneandone-flex-volume/build/git-version.sh: Permission denied
mkdir -p _output/bin/linux/
GOOS=linux go build  -ldflags "-X github.com/1and1/oneandone-flex-volume/pkg/version.Version=" -o _output/bin/linux/oneandone-flex-volume github.com/1and1/oneandone-flex-volume/cmd/oneandone-flex-volume
make: execvp: /Volumes/Data/go/src/github.com/1and1/oneandone-flex-volume/build/git-version.sh: Permission denied
mkdir -p _output/bin/darwin/
GOOS=darwin go build  -ldflags "-X github.com/1and1/oneandone-flex-volume/pkg/version.Version=" -o _output/bin/darwin/oneandone-flex-volume github.com/1and1/oneandone-flex-volume/cmd/oneandone-flex-volume
```
4. On master and workers create directory `/opt/kubernetes/kubelet-plugins/volume/exec/oneandone-flex-volume`
5. Upload binaries to master and workers:
```
$ scp -i id_rsa $GOPATH/src/github.com/1and1/oneandone-flex-volume/_output/bin/linux/oneandone-flex-volume core@[master_or_worker_ip]:/opt/kubernetes/kubelet-plugins/volume/exec/oneandone-flex-volume
```

6. Create a pod that is using flex volume:

example_pod.yaml:
```
apiVersion: v1
kind: Pod
metadata:
  name: test-do-volumes-01
spec:
  containers:
  - image: nginx
    name: nginx-demo-01
    ports:
      - name: web
        containerPort: 80
        protocol: TCP
    volumeMounts:
      - name: html
        mountPath: "/usr/share/nginx/html"
  volumes:
  - name: html
    persistentVolumeClaim:
      claimName: pv1and1
```

```
kubectl create -f example_pod.yaml
```

