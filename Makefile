BUILD ?= prod

dev:
	go build -tags dev
	$(MAKE) frontend BUILD=dev

test:
	go test -v

test-int: 
	cd tests && .venv/bin/pytest -v

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

tests/.venv:
	[ ! -d $@ ] && python3 -m venv $@
	$@/bin/pip install -r tests/requirements.txt

tests/requirements.txt: tests/requirements.in
	pip-compile $< > $@


.PHONY: dev test frontend frontend-watch build-ct-image
.ONESHELL:
