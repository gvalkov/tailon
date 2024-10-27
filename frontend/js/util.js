export function formatBytes(size) {
    var units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    var i = 0;
    while(size >= 1024) {
        size /= 1024;
        ++i;
    }
    return size.toFixed(1) + ' ' + units[i];
}

export function formatFilename(state) {
    if (!state.id) return state.text;
    var size = formatBytes($(state.element).data('size'));
    return '<span>' + state.text + '</span>' + '<span style="float:right;">' + size + '</span>';
}

export function endsWith(str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1;
}

export function startsWith(str, prefix) {
    return str.indexOf(prefix) === 0;
}

const escape_entity_map = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    "/": '&#x2F;'
};

// This is the escapeHtml function from mustache.js.
export function escapeHtml(str) {
    return String(str).replace(/[&<>\/]/g, function (s) {
        return escape_entity_map[s];
    });
}

export function parseQueryString(str) {
    var res = {};

    str.substr(1).split('&').forEach(function(item) {
        var el = item.split("=");

        var key = el[0];
        var value = el[1] && decodeURIComponent(el[1]);

        if (key in res) {
            res[key].push(value);
        } else {
            res[key] = [value];
        }
    });

    return res;
}
