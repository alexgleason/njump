package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip11"
)

func renderRelayPage(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	hostname := code[2:]
	isSitemap := false
	numResults := 1000

	if strings.HasSuffix(hostname, ".xml") {
		hostname = hostname[:len(hostname)-4]
		numResults = 5000
		isSitemap = true
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	// relay metadata
	info, _ := nip11.Fetch(r.Context(), hostname)
	if info == nil {
		info = &nip11.RelayInformationDocument{
			Name: hostname,
		}
	}

	// last notes
	var lastNotes []*nostr.Event
	if relay, err := pool.EnsureRelay(hostname); err == nil {
		lastNotes, _ = relay.QuerySync(ctx, nostr.Filter{
			Kinds: []int{1},
			Limit: numResults,
		})
	}
	renderableLastNotes := make([]EnhancedEvent, len(lastNotes))
	lastEventAt := time.Now()
	if len(lastNotes) > 0 {
		lastEventAt = time.Unix(int64(lastNotes[0].CreatedAt), 0)
	}
	for i, levt := range lastNotes {
		renderableLastNotes[i] = EnhancedEvent{levt, []string{"wss://" + hostname}}
	}

	if len(renderableLastNotes) != 0 {
		w.Header().Set("Cache-Control", "max-age=3600")
	} else {
		w.Header().Set("Cache-Control", "max-age=60")
	}

	if !isSitemap {
		RelayTemplate.Render(w, &RelayPage{
			HeadCommonPartial: HeadCommonPartial{IsProfile: false},
			ClientsPartial: ClientsPartial{
				Clients: []ClientReference{
					{Name: "Coracle", URL: "https://coracle.social/relays/" + hostname},
				},
			},

			Type:       "relay",
			Info:       info,
			Hostname:   hostname,
			Proxy:      "https://" + hostname + "/njump/proxy?src=",
			LastNotes:  renderableLastNotes,
			ModifiedAt: lastEventAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	} else {
		// ArchiveSitemapTemplate.Render TODO
	}
}
