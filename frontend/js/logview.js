Vue.component('logview', {
    template: '<div class="log-view"></div>',
    props: ["linesOfHistory"],
    data: function() {
        return {
            history: [],
            lastSpan: null,
            lastSpanClasses: '',
            autoScroll: true
        };
    },
    watch: {
        linesOfHistory: function(val) {
            this.trimHistory();
        }
    },
    methods: {
        clearLines: function() {
            this.$el.innerHTML = '';
            this.history = [],
                this.lastSpan = null;
        },
        toggleWrapLines: function(val) {
            this.$el.classList.toggle('log-view-wrapped', val);
        },
        createSpan: function(innerHtml, classNames) {
            var span = document.createElement('span');
            span.innerHTML = innerHtml;
            span.className = classNames;
            return span;
        },

        createLogEntrySpan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry');
        },
        createEmergencyPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-emergency');
        },
        createAlertPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-alert');
        },
        createCriticalPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-critical');
        },
        createErrorPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-error');
        },
        createWarningPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-warning');
        },
        createNoticePan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-notice');
        },
        createInfoPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-info');
        },
        createDebugPan: function(innerHtml) {
            return this.createSpan(innerHtml, 'log-entry log-debug');
        },

        trimHistory: function() {
            if (this.linesOfHistory !== 0 && this.history.length > this.linesOfHistory) {
                for (var i = 0; i < (this.history.length - this.linesOfHistory + 1); i++) {
                    this.$el.removeChild(this.history.shift());
                }
            }
        },

        isScrolledToBottom: function() {
            var elParent = this.$el.parentElement;
            var autoScrollOffset = elParent.scrollTop - (elParent.scrollHeight - elParent.offsetHeight);
            return Math.abs(autoScrollOffset) < 50;
        },

        scroll: function() {
            this.$el.parentElement.scrollTop = this.$el.parentElement.scrollHeight;
        },

        write: function(source, line) {
            var span;
            if (source === "o") {
                line = escapeHtml(line).replace(/\n$/, '');
                var logtype = lineTypeLog(line);
                if (logtype == "EMERGENCY") {
                    // span = this.createEmergencyPan(line);
                    line = line.replace(".EMERGENCY: ", ".EMERGENCY <span class=\"badge badge-level-emergency\"><span class=\"mdi mdi-bug\"></span> Emergency</span>: ");
                } else if (logtype == "ALERT") {
                    // span = this.createAlertPan(line);
                    line = line.replace(".ALERT: ", ".ALERT <span class=\"badge badge-level-alert\"><span class=\"mdi mdi-bullhorn\"></span> Alert</span>: ");
                } else if (logtype == "CRITICAL") {
                    // span = this.createCriticalPan(line);
                    line = line.replace(".CRITICAL: ", ".CRITICAL <span class=\"badge badge-level-critical\"><span class=\"mdi mdi-heart-pulse\"></span> Critical</span>: ");
                } else if (logtype == "ERROR") {
                    // span = this.createErrorPan(line);
                    line = line.replace(".ERROR: ", ".ERROR <span class=\"badge badge-level-error\"><span class=\"mdi mdi-alpha-e-circle\"></span> Error</span>: ");
                } else if (logtype == "WARNING") {
                    // span = this.createWarningPan(line);
                    line = line.replace(".WARNING: ", ".WARNING <span class=\"badge badge-level-warning\"><span class=\"mdi mdi-alert\"></span> Warning</span>: ");
                } else if (logtype == "NOTICE") {
                    // span = this.createNoticePan(line);
                    line = line.replace(".NOTICE: ", ".NOTICE <span class=\"badge badge-level-notice\"><span class=\"mdi mdi-alert-circle\"></span> Notice</span>: ");
                } else if (logtype == "INFO") {
                    // span = this.createInfoPan(line);
                    line = line.replace(".INFO: ", ".INFO <span class=\"badge badge-level-info\"><span class=\"mdi mdi-information\"></span> Info</span>: ");
                } else if (logtype == "DEBUG") {
                    // span = this.createDebugPan(line);
                    line = line.replace(".DEBUG: ", ".DEBUG <span class=\"badge badge-level-debug\"><span class=\"mdi mdi-lifebuoy\"></span> Debug</span>: ");
                } else {
                    // span = this.createLogEntrySpan(line);
                }
                span = this.createLogEntrySpan(line);
                this.writeSpans([span]);
            }
        },

        writeSpans: function(spanArray) {
            if (spanArray.length === 0) {
                return;
            }

            var scrollAfterWrite = this.isScrolledToBottom();

            // Create spans from all elements and add them to a temporary DOM.
            var fragment = document.createDocumentFragment();
            for (var i = 0; i < spanArray.length; i++) {
                var span = spanArray[i];
                this.history.push(span);
                fragment.appendChild(span);
            }

            if (this.lastSpan) {
                this.lastSpan.className = this.lastSpanClasses;
            }

            this.$el.appendChild(fragment);
            this.trimHistory();

            if (this.autoScroll && scrollAfterWrite) {
                this.scroll();
            }

            this.lastSpan = this.history[this.history.length - 1];
            this.lastSpanClasses = this.lastSpan.className;
            this.lastSpan.className = this.lastSpanClasses + ' log-entry-current';

        }
    }
});