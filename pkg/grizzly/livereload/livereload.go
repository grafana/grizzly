package livereload

import (
	_ "embed"

	"fmt"
	"net/http"
)

// Initialize starts the Websocket Hub handling live reloads.
func Initialize() {
	go wsHub.run()
}

// ServeJS serves the livereload.js who's reference is injected into the page.
func ServeJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(liveReloadJS())
}

func liveReloadJS() []byte {
	return []byte(livereloadJS + hugoLiveReloadPlugin)
}

var (
	// This is a patched version, see https://github.com/livereload/livereload-js/pull/84
	//go:embed livereload.js
	livereloadJS         string
	hugoLiveReloadPlugin = fmt.Sprintf(`
/*
Hugo adds a specific prefix, "__hugo_navigate", to the path in certain situations to signal
navigation to another content page.
*/

function HugoReload() {}

HugoReload.identifier = 'hugoReloader';
HugoReload.version = '0.9';

HugoReload.prototype.reload = function(path, options) {
	var prefix = %q;

	if (path.lastIndexOf(prefix, 0) !== 0) {
		return false
	}
	
	path = path.substring(prefix.length);

	var portChanged = options.overrideURL && options.overrideURL != window.location.port
	
	if (!portChanged && window.location.pathname === path) {
		window.location.reload();
	} else {
		if (portChanged) {
			window.location = location.protocol + "//" + location.hostname + ":" + options.overrideURL + path;
		} else {
			window.location.pathname = path;
		}
	}

	return true;
};

LiveReload.addPlugin(HugoReload)
`, hugoNavigatePrefix)
)

// Prefix to signal to LiveReload that we need to navigate to another path.
const hugoNavigatePrefix = "__hugo_navigate"
