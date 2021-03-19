#### Running Tests - failed to start the controlplane
If you see the following error when running the tests 
```  
Unexpected error:
      <*errors.errorString | 0xc0003963f0>: {
          s: "failed to start the controlplane. retried 5 times: fork/exec /usr/local/kubebuilder/bin/etcd: no such file or directory",
      }
      failed to start the controlplane. retried 5 times: fork/exec /usr/local/kubebuilder/bin/etcd: no such file or directory
  occurred
```
This may be because your kubebuilder was installed via brew or a package manager

A solution is to install kubebuilder following these instructions: 

```
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/

# move to a long-term location and put it on your path
# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin
```

The above extract was taken https://book.kubebuilder.io/quick-start.html#installation