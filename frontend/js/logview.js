import { ref, useTemplateRef } from "vue";
import { escapeHtml } from "./util.js";

export default {
    template: '<div class="log-view"></div>',
    props: ["linesOfHistory"],
    setup() {
        const logview = useTemplateRef("logview");
    },
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
        clearLines: function () {
            this.$el.innerHTML = '';
            this.history = [],
            this.lastSpan = null;
        },
        toggleWrapLines: function(val) {
            this.$el.classList.toggle('log-view-wrapped', val);
        },
        createSpan: function (innerHtml, classNames) {
            var span = document.createElement('span');
            span.innerHTML = innerHtml;
            span.className = classNames;
            return span;
        },

        createLogEntrySpan: function (innerHtml) {
            return this.createSpan(innerHtml, 'log-entry');
        },

        createNoticePan: function (innerHtml) {
            return createSpan(innerHtml, 'log-entry log-notice');
        },

        trimHistory: function () {
            if (this.linesOfHistory !== 0 && this.history.length > this.linesOfHistory) {
                for (var i = 0; i < (this.history.length - this.linesOfHistory + 1); i++) {
                    this.$el.removeChild(this.history.shift());
                }
            }
        },

        isScrolledToBottom: function () {
            var elParent = this.$el.parentElement;
            var autoScrollOffset = elParent.scrollTop - (elParent.scrollHeight - elParent.offsetHeight);
            return Math.abs(autoScrollOffset) < 50;
        },

        scroll: function() {
            this.$el.parentElement.scrollTop = this.$el.parentElement.scrollHeight;
        },

        write: function (source, line) {
            var span;
            if (source === "o") {
                line = escapeHtml(line).replace(/\n$/, '');
                span = this.createLogEntrySpan(line);

                this.writeSpans([span]);
            }
        },

        writeSpans: function (spanArray) {
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

            this.lastSpan = this.history[this.history.length-1];
            this.lastSpanClasses = this.lastSpan.className;
            this.lastSpan.className = this.lastSpanClasses + ' log-entry-current';
        }
    }
};
