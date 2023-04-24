# d-op
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

### Running on the cluster
1. run k8s cluster `minikube start`
2. apply manifests to cluster `make deploy`
3. ensure pod is running `kubectl -n d-op-system get pods`
4. create new Dummy `kubectl apply -f Dummy.yml`
5. ensure dummy is created and changed `kubectl get dummy dummy-sample -o yaml`
6. ensure nginx pod is running `kubectl get pods`
7. remove dummy `kubectl -n default delete dummy dummy-sample`
8. ensure nginx pod was deleted `kubectl get pods`