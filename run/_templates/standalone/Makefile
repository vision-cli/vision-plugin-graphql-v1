GITHUB_SHA?=${shell git rev-parse --short HEAD || echo latest}
METHOD?=SayHello

DOCKER_IMAGE_BASE = aimee.default.graphql
# Assuming the local Docker registry to be the kind-registry container created by create_local_cluster.sh
DOCKER_LOCAL_REGISTRY = 127.0.0.1:8081
# The DOCKER_LOCAL_IMAGE_NAME_TAG has the following format: [REGISTRY]/[PROJECT/NAMESPACE]/[SERVICE]:[TAG]
DOCKER_LOCAL_IMAGE_NAME_TAG = $(DOCKER_LOCAL_REGISTRY)/$(DOCKER_IMAGE_BASE):latest

test:
	@go test -race -coverprofile=coverage.txt -v $$(go list ./...)
.PHONY: test

run:
	GRAPHIQL_ENABLED=true go run cmd/main.go
.PHONY: run

# !!! IMPORTANT !!! The Docker build context is set to the project root directory
# by specifying it using the relative path: ./../../..
build-image:
	@echo "Building *LOCAL* Docker image with name and tag: [gcr/${DOCKER_IMAGE_BASE}:latest]"
	@echo "-----------------------------------------------"
	docker build \
		--file ./Dockerfile \
		--progress=plain \
		--tag "gcr/${DOCKER_IMAGE_BASE}:${GITHUB_SHA}" \
		--tag "gcr/${DOCKER_IMAGE_BASE}:latest" \
		./../../..
.PHONY: build-image

push-image: build-image
	@echo "Pushing *LOCAL* Docker image with name and tag: [gcr/${DOCKER_IMAGE_BASE}:latest]"
	@echo "-----------------------------------------------"
	docker push "gcr/${DOCKER_IMAGE_BASE}:${GITHUB_SHA}"
	docker push "gcr/${DOCKER_IMAGE_BASE}:latest"
.PHONY: push-image

deploy: push-image
	@kubectl apply -f manifests.yml
.PHONY: deploy




