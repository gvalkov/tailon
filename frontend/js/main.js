Vue.component('multiselect', window.VueMultiselect.default);
Vue.component('vue-loading', window.VueLoading);

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
        clearLines: function () {
            this.$el.innerHTML = '';
            this.history = [],
            this.lastSpan = null;
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
});

var apiURL = endsWith(window.relativeRoot, '/') ? 'ws' : '/ws';
var apiURL = [window.location.protocol, '//', window.location.host, window.relativeRoot, apiURL].join('');

var app = new Vue({
    el: '#app',
    delimiters: ['<%', '%>'],
    data: {
        'relativeRoot': relativeRoot,
        'commandScripts': commandScripts,

        'fileList': [],
        'allowCommandNames': allowCommandNames,
        'allowDownload': allowDownload,

        'file': null,
        'command': null,
        'script': null,

        'linesOfHistory': 2000,  // 0 for infinite history
        'linesToTail': 10,
        'wrapLines': false,

        'hideToolbar': false,
        'showConfig': false,
        'showLoadingOverlay': false,

        'socket': null,
        'isConnected': false
    },
    created: function () {
        this.backendConnect();
        this.command = this.allowCommandNames[0];
    },
    computed: {
        scriptInputEnabled: function () {
            return this.commandScripts[this.command] !== "";
        },
        downloadLink: function () {
            if (this.file) {
                return relativeRoot + 'files/?path=' + this.file.path;
            }
            return '#';
        }
    },
    methods: {
        clearLogview: function () {
            this.$refs.logview.clearLines();
        },
        backendConnect: function ( ){
            console.log('connecting to ' + apiURL);
            this.showLoadingOverlay = true;
            this.socket = new SockJS(apiURL);
            this.socket.onopen = this.onBackendOpen;
            this.socket.onclose = this.onBackendClose;
            this.socket.onmessage = this.onBackendMessage;
        },
        onBackendOpen: function () {
            console.log('connected to backend');
            this.isConnected = true;
            this.refreshFiles();
        },
        onBackendClose: function () {
            console.log('disconnected from backend');
            this.isConnected = false;
            backendConnect = this.backendConnect;
            window.setTimeout(function () {
                backendConnect();
            }, 1000);
        },
        onBackendMessage: function (message) {
            var data = JSON.parse(message.data);

            if (data.constructor === Object) {
                // Reshape into something that vue-multiselect :group-select can use.
                var fileList = [];
                Object.keys(data).forEach(function (key) {
                    var group = ("__default__" === key) ? "Ungrouped Files" : key;
                    fileList.push({
                        "group": group,
                        "files": data[key]
                    });
                });

                this.fileList = fileList;

                // Set file input to first entry in list.
                if (!this.file) {
                    this.file = fileList[0].files[0];
                }
            } else {
                var stream = data[0];
                var line = data[1];
                this.$refs.logview.write(stream, line);
            }
        },
        refreshFiles: function () {
            console.log("updating file list");
            this.socket.send("list");
        },
        notifyBackend: function () {
            var msg = {
                command: this.command,
                script: this.script,
                entry: this.file,
                nlines: this.linesToTail
            };
            console.log("sending msg: ", msg);
            this.clearLogview();
            this.socket.send(JSON.stringify(msg));
        }
    },
    watch: {
        isConnected: function(val) {
            this.showLoadingOverlay = !val;
        },
        command: function(val) {
            if (val && this.isConnected) {
                this.script = this.commandScripts[val];
                this.notifyBackend();
            }
        },
        file: function(val) {
            if (val && this.isConnected) {
                this.notifyBackend();
            }
        }
    }
});
