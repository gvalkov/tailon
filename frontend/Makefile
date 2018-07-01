CMD_TYPESCRIPT := node_modules/typescript/bin/tsc
CMD_POSTCSS    := node_modules/postcss-cli/bin/postcss
CMD_NODESASS   := node_modules/node-sass/bin/node-sass

DISTDIR ?= dist

EXTERNAL_JS := \
	node_modules/vue/dist/vue.js \
	node_modules/vue-multiselect/dist/vue-multiselect.min.js \
	node_modules/sockjs-client/dist/sockjs.js \
	node_modules/html5shiv/dist/html5shiv.js \
	node_modules/@fortawesome/fontawesome-free/js/all.js \
	node_modules/vue-loading-overlay/dist/vue-loading.min.js

EXTERNAL_CSS := \
	node_modules/normalize-css/normalize.css \
	node_modules/vue-multiselect/dist/vue-multiselect.min.css \
	node_modules/vue-loading-overlay/dist/vue-loading.min.css

INTERNAL_JS := \
	js/util.js \
	js/main.js

INTERNAL_CSS := \
	scss/vars.scss \
	scss/main.scss

ALL_TARGETS := \
	$(DISTDIR)/3rdparty.js \
	$(DISTDIR)/3rdparty.css \
	$(DISTDIR)/main.css \
	$(DISTDIR)/main.js


all: $(ALL_TARGETS)

clean:
	-rm -f $(ALL_TARGETS)

$(DISTDIR)/3rdparty.js: $(EXTERNAL_JS)
	@cat $(EXTERNAL_JS) > $@

$(DISTDIR)/main.js: $(INTERNAL_JS)
	cat $(INTERNAL_JS) > $@

$(DISTDIR)/main.css: $(INTERNAL_CSS)
	$(CMD_NODESASS) --quiet --output-style compressed scss/main.scss > $@

$(DISTDIR)/3rdparty.css: $(EXTERNAL_CSS)
	@cat $? > $@

watch:
	@echo $(INTERNAL_CSS) $(EXTERNAL_CSS) $(EXTERNAL_JS) $(INTERNAL_JS) | tr ' ' '\n' | entr make


.PHONY: all clean watch
