VERSION := 0.0.4

.PHONY: build
build:
	docker build --tag remmelt/evohome-prometheus-export:${VERSION} \
	--no-cache --force-rm --pull --rm .

.PHONY: push
push:
	docker push remmelt/evohome-prometheus-export:${VERSION}
