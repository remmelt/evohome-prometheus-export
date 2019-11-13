.PHONY: build
build:
	docker build --tag remmelt/evohome-prometheus-export:0.0.3 \
	--no-cache --force-rm --pull --rm .

.PHONY: push
push:
	docker push remmelt/evohome-prometheus-export:0.0.3
