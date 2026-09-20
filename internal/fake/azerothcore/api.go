package fakeazerothcore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Reset restores the seed accounts, removes accounts created since startup and
// clears the command journal. When a login database is mirrored, it is brought
// back in sync so the gateway's read adapter sees the reset state.
func (s *Server) Reset() {
	s.mu.Lock()
	seedKeys := make(map[string]struct{}, len(s.seed))
	for _, account := range s.seed {
		seedKeys[normalizeAccountKey(account.Username)] = struct{}{}
	}
	extra := make([]string, 0)
	for key := range s.accounts {
		if _, ok := seedKeys[key]; !ok {
			extra = append(extra, key)
		}
	}
	s.mu.Unlock()

	for _, key := range extra {
		s.mysql.deleteAccount(key)
	}
	s.loadAccounts()
	for _, account := range s.seed {
		s.mysql.upsertAccount(account.Username, account.Email)
		s.mysql.setGMLevel(account.Username, account.GMLevel)
		if account.Banned {
			seconds, ok := timeStringToSeconds(account.BanDuration)
			if !ok {
				seconds = 0
			}
			s.mysql.setBanned(account.Username, seconds, account.BanReason)
		} else {
			s.mysql.clearBan(account.Username)
		}
	}
	s.journal.clear()
	s.logger.Info("fake-azerothcore: state reset")
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	writeJSON(w, map[string]any{"commands": s.journal.list(limit)})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.Reset()
	w.WriteHeader(http.StatusNoContent)
}

// handleStream pushes new commands as Server-Sent Events.
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	id, events := s.journal.subscribe()
	defer s.journal.unsubscribe(id)

	for {
		select {
		case <-r.Context().Done():
			return
		case entry := <-events:
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: command\ndata: %s\n\n", data); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(dashboardHTML))
}

// dashboardHTML is a self-contained live view of the fake's internal state.
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>fake AzerothCore — internal view</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; margin: 1.5rem; line-height: 1.4; }
  h1 { font-size: 1.1rem; }
  h2 { font-size: 0.95rem; margin-top: 1.5rem; }
  button { font: inherit; padding: 0.35rem 0.7rem; border: 1px solid #8888; border-radius: 6px; background: transparent; color: inherit; cursor: pointer; }
  table { border-collapse: collapse; width: 100%; font-size: 0.85rem; }
  th, td { border-bottom: 1px solid #8884; text-align: left; padding: 0.25rem 0.5rem; vertical-align: top; }
  td.result { white-space: pre-wrap; }
  .rid { opacity: 0.7; }
  .dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #d33; margin-right: 0.4rem; }
  .dot.on { background: #2a2; }
  pre { background: #8882; padding: 0.75rem; border-radius: 6px; overflow-x: auto; }
</style>
</head>
<body>
<h1>fake AzerothCore — internal view</h1>
<p><span id="dot" class="dot"></span><span id="status">connecting…</span>
  <button id="refresh">Refresh</button>
  <button id="reset">Reset state</button>
</p>

<h2>Command journal (newest first)</h2>
<table>
  <thead><tr><th>#</th><th>time</th><th>request_id</th><th>command</th><th>result</th></tr></thead>
  <tbody id="rows"></tbody>
</table>

<h2>Account state</h2>
<pre id="state">—</pre>

<script>
  var rows = document.getElementById('rows');
  var stateEl = document.getElementById('state');
  var statusEl = document.getElementById('status');
  var dot = document.getElementById('dot');

  function esc(value) {
    return String(value == null ? '' : value)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  function addRow(entry, prepend) {
    var tr = document.createElement('tr');
    tr.innerHTML = '<td>' + entry.seq + '</td>' +
      '<td>' + esc(new Date(entry.at).toLocaleTimeString()) + '</td>' +
      '<td class="rid">' + esc(entry.request_id) + '</td>' +
      '<td>' + esc(entry.command) + '</td>' +
      '<td class="result">' + esc(entry.result) + '</td>';
    if (prepend && rows.firstChild) { rows.insertBefore(tr, rows.firstChild); }
    else { rows.appendChild(tr); }
  }

  function refresh() {
    fetch('/commands?limit=100').then(function (r) { return r.json(); }).then(function (data) {
      rows.innerHTML = '';
      (data.commands || []).forEach(function (entry) { addRow(entry, false); });
    });
    fetch('/state').then(function (r) { return r.json(); }).then(function (data) {
      stateEl.textContent = JSON.stringify(data, null, 2);
    });
  }

  document.getElementById('refresh').addEventListener('click', refresh);
  document.getElementById('reset').addEventListener('click', function () {
    fetch('/reset', { method: 'POST' }).then(refresh);
  });

  var source = new EventSource('/commands/stream');
  source.addEventListener('open', function () { dot.className = 'dot on'; statusEl.textContent = 'live'; });
  source.addEventListener('error', function () { dot.className = 'dot'; statusEl.textContent = 'reconnecting…'; });
  source.addEventListener('command', function (event) {
    addRow(JSON.parse(event.data), true);
    fetch('/state').then(function (r) { return r.json(); }).then(function (data) {
      stateEl.textContent = JSON.stringify(data, null, 2);
    });
  });

  refresh();
</script>
</body>
</html>
`
