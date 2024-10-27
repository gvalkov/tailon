BUILD ?= prod

dev:
	go build -tags dev
	$(MAKE) frontend BUILD=dev

test:
	go test -v
	cd tests && pytest -v

frontend:
	cd frontend && $(MAKE) clean ; $(MAKE) BUILD=$(BUILD)

frontend-watch:
	cd frontend && $(MAKE) watch BUILD=$(BUILD)

build-ct-image:
	podman build -t gvalkov/tailon .

README.md: main.go
	go build
	ed $@ <<< $$'/BEGIN HELP/+2,/END HELP/-2d\n/BEGIN HELP/+1r !./tailon --help 2>&1\n,w'
	sed -i 's/[ \t]*$$//' $@



.PHONY: dev test frontend frontend-watch build-ct-image
