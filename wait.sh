#!/bin/bash
while ! kubectl wait pod --for=condition=Ready --timeout="5m" -l "app=qserv"
do
    echo "Wait for Qserv pods to be ready:"
    kubectl get pods
    kubectl describe pods
done
