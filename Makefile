TAG := $(shell git describe --dirty --always)
# Image version create by build procedure
OP_IMAGE := "qserv/qserv-operator:${TAG}"

deploy: build
	kubectl create namespace qserv
	kubectl config set-context $(shell kubectl config current-context) --namespace=qserv
	./deploy.sh -n qserv
	./wait-operator-ready.sh
	kubectl apply -k base -n qserv
	./wait-qserv-ready.sh
build:
	operator-sdk generate k8s
	operator-sdk generate crds
	GO111MODULE="on" operator-sdk build ${OP_IMAGE}
	sed "s|REPLACE_IMAGE|${OP_IMAGE}|g" "./deploy/operator.yaml.tpl" \
		> "./deploy/operator.yaml"
	kind load docker-image ${OP_IMAGE}
test: deploy
	./run-integration-tests.sh
delete:
	kubectl delete namespace qserv
.PHONY: build test delete
