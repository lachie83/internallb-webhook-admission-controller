DOCKER_IMAGE ?= lachlanevenson/internallb-webhook-admission-controller
GIT_BRANCH ?= `git rev-parse --abbrev-ref HEAD`

ifeq ($(GIT_BRANCH), master)
	DOCKER_TAG = latest
else
	DOCKER_TAG = $(GIT_BRANCH)
endif

build:
	docker build -t internallb-webhook-admission-controller .

docker_push:
	# Push to DockerHub
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
