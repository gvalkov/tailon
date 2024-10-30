CMD_CLEANCSS   := node_modules/clean-css-cli/bin/cleancss
CMD_NODESASS   := node_modules/sass/sass.js

BUILD ?= prod
DISTDIR ?= dist

EXTERNAL_CSS :=
EXTERNAL_CSS += node_modules/normalize.css/normalize.css
EXTERNAL_CSS += node_modules/vue-multiselect/dist/vue-multiselect.min.css
EXTERNAL_CSS += node_modules/vue-loading-overlay/dist/css/index.css

ALL_TARGETS := 
ALL_TARGETS += $(DISTDIR)/main.js
ALL_TARGETS += $(DISTDIR)/3rdparty.css
ALL_TARGETS += $(DISTDIR)/main.css 
 
CSS := $(wildcard scss/*.scss)
JS := $(wildcard js/*.js)

ifeq ($(BUILD),dev)
ESBUILD_ARGS := --sourcemap
SCSS_ARGS := --style expanded --source-map 
endif

ifeq ($(BUILD),prod)
ESBUILD_ARGS := --minify
SCSS_ARGS := --style compressed --no-source-map
endif


all: $(ALL_TARGETS)

clean:
	@-rm -fv $(ALL_TARGETS) dist/*.map

watch:
	@echo $(JS) $(CSS) $(EXTERNAL_CSS) | tr ' ' '\n' | entr $(MAKE)

du:
	@du -hs dist/* | sort -h

$(DISTDIR)/main.js: $(JS)
	npx esbuild --bundle $(ESBUILD_ARGS) --outdir=dist/ js/main.js

$(DISTDIR)/3rdparty.css: $(EXTERNAL_CSS)
	$(CMD_CLEANCSS) $(EXTERNAL_CSS) > $@

$(DISTDIR)/main.css: $(CSS)
	$(CMD_NODESASS) --quiet $(SCSS_ARGS) scss/main.scss:$@

.PHONY: all clean watch du 
