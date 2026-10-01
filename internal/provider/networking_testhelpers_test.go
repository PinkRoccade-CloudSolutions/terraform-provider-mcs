package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// netStatefulMock registers a CRUD handler on basePath that stores a single object:
// POST merges the body into defaults, GET/PUT/PATCH/DELETE operate on the stored object.
// onWrite, if set, may adjust the stored object after each POST/PUT/PATCH.
func netStatefulMock(m *mockAPIServer, basePath string, defaults map[string]interface{}, onWrite func(method string, obj map[string]interface{})) {
	var mu sync.Mutex
	var obj map[string]interface{}

	merge := func(body []byte) {
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		for k, v := range req {
			obj[k] = v
		}
	}

	m.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		isList := strings.TrimSuffix(r.URL.Path, "/") == strings.TrimSuffix(basePath, "/")

		switch r.Method {
		case http.MethodPost:
			obj = map[string]interface{}{}
			for k, v := range defaults {
				obj[k] = v
			}
			merge(body)
			if onWrite != nil {
				onWrite(r.Method, obj)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(obj)
		case http.MethodGet:
			if isList {
				results := []interface{}{}
				if obj != nil {
					results = append(results, obj)
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
				return
			}
			if obj == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
				return
			}
			_ = json.NewEncoder(w).Encode(obj)
		case http.MethodPut, http.MethodPatch:
			merge(body)
			if onWrite != nil {
				onWrite(r.Method, obj)
			}
			_ = json.NewEncoder(w).Encode(obj)
		case http.MethodDelete:
			obj = nil
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

// netPagedMock serves a list endpoint across several pages, linking them with absolute
// `next` URLs. Every raw query string received is appended to queries.
func netPagedMock(m *mockAPIServer, basePath string, pages [][]map[string]interface{}, queries *[]string) {
	var mu sync.Mutex
	m.On(basePath, func(w http.ResponseWriter, r *http.Request, _ []byte) {
		mu.Lock()
		*queries = append(*queries, r.URL.RawQuery)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		page := 1
		_, _ = fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
		if page < 1 || page > len(pages) {
			page = 1
		}
		var next interface{}
		if page < len(pages) {
			next = fmt.Sprintf("%s%s?page=%d", m.URL(), basePath, page+1)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(pages),
			"next":    next,
			"results": pages[page-1],
		})
	})
}

// netCallBodies returns the bodies of all recorded calls with the given method and path prefix.
func netCallBodies(m *mockAPIServer, method, prefix string) []string {
	var out []string
	for _, c := range m.Calls() {
		if c.Method == method && strings.HasPrefix(c.Path, prefix) {
			out = append(out, c.Body)
		}
	}
	return out
}
