#! /bin/sh

# Create a local K8S cluster (using Kind) with a Yunikorn scheduler

arch=arm64         # either amd64 or arm64
ns=yunikorn
registry=gresearch

# The base directory which contains all the Yunikorn repo dirs below
yk_repos_base=$HOME/src/yunikorn

echo 'Creating kind cluster ...'
kind create cluster
kubectl create namespace $ns

echo ''
echo 'Building Yunikorn images ...'
cd $yk_repos_base/yunikorn-k8shim
make clean
make image DOCKER_ARCH=${arch} REGISTRY=${registry} VERSION=latest

cd $yk_repos_base/yunikorn-web
make image DOCKER_ARCH=${arch} REGISTRY=${registry} VERSION=latest

# Verify we have all local ${registry}/yunikorn images
docker image ls -a | grep yuni

echo ''
echo 'Loading Yunikorn images into Kind cluster...'
for img in admission-${arch}-latest scheduler-plugin-${arch}-latest scheduler-${arch}-latest web-${arch}-latest
do
  kind load docker-image ${registry}/yunikorn:$img
done

echo ''
echo 'Deploying Yunikorn admission-controller and scheduler ...'
cd $yk_repos_base/yunikorn-k8shim

kubectl create -f deployments/scheduler/yunikorn-rbac.yaml -n $ns

# kubectl create configmap yunikorn-configs --from-file=deployments/scheduler/yunikorn-configs.yaml -n $ns
kubectl apply -f deployments/scheduler/yunikorn-configs.yaml -n $ns

for dpmt in scheduler-load admission-controller-rbac admission-controller-secrets admission-controller; do
    kubectl create -f deployments/scheduler/${dpmt}.yaml -n $ns
done

cat <<USAGE

Done! In a separate terminal, run:
    kubectl port-forward svc/yunikorn-service 9889 9080 --address 127.0.0.1 -n $ns

then open your browser to http://127.0.0.1:9889/

To shut down everything, run
    kind delete cluster

USAGE
