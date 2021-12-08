# platform-gc-controller 
----
platform-gc-controller is used to gc resources under group platform.tkestack.io
----

## To start using K8s

### use binary
./platform-gc-controller --kubeconfig ./admin.conf --leader-elect=true --gcgroup="platform.tkestack.io"

### deploy as a docker
make
