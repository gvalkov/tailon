prod:
	$(MAKE) BUILD=prod frontend
	go build -v

dev:
	$(MAKE) BUILD=dev frontend
	go build -tags dev

test:
	go test -v

test-int: 
	cd tests && .venv/bin/pytest -v

frontend:
	make -C frontend clean all du

frontend-watch:
	make -C frontend watch

build-ct-image:
	podman build -t gvalkov/tailon .

README.md:
	go build
	ed $@ <<< $$'/BEGIN HELP_USAGE/+2,/END HELP_USAGE/-2d\n/BEGIN HELP_USAGE/+1r !./tailon --help 2>&1\n,w'
	ed $@ <<< $$'/BEGIN HELP_CONFIG/+2,/END HELP_CONFIG/-2d\n/BEGIN HELP_CONFIG/+1r !./tailon --help-config 2>&1\n,w'
	sed -i 's/[ \t]*$$//' $@

tests/.venv:
	[ ! -d $@ ] && python3 -m venv $@
	$@/bin/pip install -r tests/requirements.txt

tests/requirements.txt: tests/requirements.in
	pip-compile $< > $@


.PHONY: dev test frontend frontend-watch build-ct-image README.md
.ONESHELL:
