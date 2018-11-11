DOCKERREPO := transport

include ../common.mk

run:
	docker run --rm -it ${REGISTRY}/${DOCKERREPO}:${TAG}

clean:
