CMD_CLEANCSS   := node_modules/clean-css-cli/bin/cleancss
CMD_NODESASS   := node_modules/node-sass/bin/node-sass
CMD_UGLIFY     := node_modules/uglify-js/bin/uglifyjs

BUILD ?= prod
DISTDIR ?= dist

EXTERNAL_JS := \
	node_modules/vue/dist/vue.min.js \
	node_modules/vue-multiselect/dist/vue-multiselect.min.js \
	node_modules/sockjs-client/dist/sockjs.min.js \
	node_modules/html5shiv/dist/html5shiv.min.js \
	node_modules/vue-loading-overlay/dist/vue-loading.min.js

EXTERNAL_CSS := \
	node_modules/normalize-css/normalize.css \
	node_modules/vue-multiselect/dist/vue-multiselect.min.css \
	node_modules/vue-loading-overlay/dist/vue-loading.min.css

INTERNAL_JS := \
	js/util.js \
	js/logview.js \
	js/main.js

INTERNAL_CSS := \
	scss/vars.scss \
	scss/main.scss

ALL_TARGETS := \
	$(DISTDIR)/3rdparty.js \
	$(DISTDIR)/3rdparty.css \
	$(DISTDIR)/main.css \
	$(DISTDIR)/main.js \
	assets_vfsdata.go

all: $(ALL_TARGETS)

clean:
	-rm -f $(ALL_TARGETS)

watch:
	@echo $(INTERNAL_CSS) $(EXTERNAL_CSS) $(EXTERNAL_JS) $(INTERNAL_JS) | tr ' ' '\n' | entr $(MAKE)

du:
	@du -h $(ALL_TARGETS)


ifneq ($(BUILD),dev)
$(DISTDIR)/3rdparty.css: $(EXTERNAL_CSS)
	$(CMD_CLEANCSS) $(EXTERNAL_CSS) > $@

$(DISTDIR)/3rdparty.js: $(EXTERNAL_JS)
	$(CMD_UGLIFY) $(EXTERNAL_JS) > $@

$(DISTDIR)/main.js: $(INTERNAL_JS)
	$(CMD_UGLIFY) -c drop_console=true $(INTERNAL_JS) > $@

$(DISTDIR)/main.css: $(INTERNAL_CSS)
	$(CMD_NODESASS) --quiet --output-style compressed scss/main.scss > $@
else

# Usin the non-minified package disables production mode.
EXTERNAL_JS := $(EXTERNAL_JS:%/vue.min.js=%/vue.js)

$(DISTDIR)/3rdparty.js: $(EXTERNAL_JS)
	@cat $(EXTERNAL_JS) > $@

$(DISTDIR)/3rdparty.css: $(EXTERNAL_CSS)
	@cat $(EXTERNAL_CSS) > $@

$(DISTDIR)/main.js: $(INTERNAL_JS)
	@cat $(INTERNAL_JS) > $@

$(DISTDIR)/main.css: $(INTERNAL_CSS)
	$(CMD_NODESASS) --quiet scss/main.scss > $@
endif

# go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
assets_vfsdata.go:
	cd .. && vfsgendev  -source=\"github.com/gvalkov/tailon/frontend\".Assets && mv $@ frontend/


.PHONY: all clean watch du assets_vfsdata.go
