CMD_CLEANCSS   := node_modules/clean-css-cli/bin/cleancss
CMD_NODESASS   := node_modules/sass/sass.js

BUILD ?= prod
DISTDIR ?= dist

EXTERNAL_CSS :=
EXTERNAL_CSS += node_modules/normalize.css/normalize.css
EXTERNAL_CSS += node_modules/vue-multiselect/dist/vue-multiselect.min.css
EXTERNAL_CSS += node_modules/vue-loading-overlay/dist/css/index.css

ALL_TARGETS := 
ALL_TARGETS += $(DISTDIR)/bundle.js
ALL_TARGETS += $(DISTDIR)/3rdparty.css
ALL_TARGETS += $(DISTDIR)/main.css 
ALL_TARGETS += assets_vfsdata.go
 
CSS := $(wildcard scss/*.scss)
JS := $(wildcard js/*.js)

all: $(ALL_TARGETS)

clean:
	@-rm -fv $(ALL_TARGETS) dist/*.map

watch:
	@echo $(JS)  $(CSS) $(EXTERNAL_CSS) | tr ' ' '\n' | entr $(MAKE)

du:
	@du -h $(ALL_TARGETS)

$(DISTDIR)/bundle.js: $(JS)
	npx esbuild --bundle --sourcemap js/main.js > $@

$(DISTDIR)/3rdparty.css: $(EXTERNAL_CSS)
	$(CMD_CLEANCSS) $(EXTERNAL_CSS) > $@

$(DISTDIR)/main.css: $(CSS)
	$(CMD_NODESASS) --quiet --style compressed --source-map scss/main.scss:$@

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
assets_vfsdata.go:
	cd .. && vfsgendev  -source=\"github.com/gvalkov/tailon/frontend\".Assets && mv $@ frontend/

.PHONY: all clean watch du assets_vfsdata.go
