package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
)

// routeMock imitates the secure ingress route API closely enough to exercise
// the provider's normalisation and id matching:
//   - settings equal to their default are dropped, list values are sorted and
//     deduplicated, and write-only keys are reported in secrets_set only;
//   - settings are merged on PATCH, with null removing a key;
//   - locations and WAF exceptions keep their id when sent with one;
//   - the certificate endpoint accepts a multipart upload.
type routeMock struct {
	t        *testing.T
	mu       sync.Mutex
	route    map[string]interface{}
	settings map[string]string
	secrets  map[string]string
	nextLoc  int
	nextWaf  int
	conflict bool
	deleted  bool
}

var routeMockDefaults = map[string]string{"USE_GZIP": "no"}

func routeMockNormalise(key, value string) string {
	if key != "ALLOWED_METHODS" {
		return value
	}
	parts := strings.Split(value, "|")
	sort.Strings(parts)
	out := parts[:0]
	for i, p := range parts {
		if i == 0 || p != parts[i-1] {
			out = append(out, p)
		}
	}
	return strings.Join(out, "|")
}

func newRouteMock(t *testing.T, mock *mockAPIServer) *routeMock {
	m := &routeMock{t: t, settings: map[string]string{}, secrets: map[string]string{}}
	mock.On("/api/secureingress/route", m.handle)
	return m
}

func (m *routeMock) SetConflict(v bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conflict = v
}

func (m *routeMock) applySettings(in map[string]interface{}) {
	for k, v := range in {
		if v == nil {
			delete(m.settings, k)
			delete(m.secrets, k)
			continue
		}
		s := fmt.Sprint(v)
		if strings.HasSuffix(k, "_PASSWORD") {
			m.secrets[k] = s
			continue
		}
		s = routeMockNormalise(k, s)
		if routeMockDefaults[k] == s {
			delete(m.settings, k)
			continue
		}
		m.settings[k] = s
	}
}

func targetFields(dst, src map[string]interface{}) {
	kind, _ := src["target_kind"].(string)
	if kind == "" {
		switch {
		case src["virtualmachine"] != nil:
			kind = "vm"
		case src["kubernetes_cluster"] != nil:
			kind = "kubernetes"
		default:
			kind = "address"
		}
	}
	dst["target_kind"] = kind
	for _, k := range []string{"virtualmachine", "interface", "kubernetes_cluster"} {
		dst[k] = src[k]
	}
	if kind == "vm" {
		dst["upstream_address"] = "10.0.0.5"
	} else {
		dst["upstream_address"] = src["upstream_address"]
	}
	scheme, _ := src["scheme"].(string)
	if scheme == "" {
		scheme = "http"
	}
	dst["scheme"] = scheme
	dst["port"] = src["port"]
	dst["upstream"] = fmt.Sprintf("%s://%v:%v", scheme, dst["upstream_address"], src["port"])
}

func (m *routeMock) applyLocations(items []interface{}) {
	out := []interface{}{}
	for i, raw := range items {
		in := raw.(map[string]interface{})
		loc := map[string]interface{}{"path": in["path"], "slot": i + 1, "settings": map[string]interface{}{}}
		if id, ok := in["id"].(string); ok && id != "" {
			loc["id"] = id
		} else {
			m.nextLoc++
			loc["id"] = fmt.Sprintf("loc-%d", m.nextLoc)
		}
		targetFields(loc, in)
		if s, ok := in["settings"].(map[string]interface{}); ok {
			kept := map[string]interface{}{}
			for k, v := range s {
				if routeMockDefaults[k] != fmt.Sprint(v) {
					kept[k] = v
				}
			}
			loc["settings"] = kept
		}
		out = append(out, loc)
	}
	m.route["locations"] = out
}

func (m *routeMock) applyWaf(items []interface{}) {
	out := []interface{}{}
	for _, raw := range items {
		in := raw.(map[string]interface{})
		w := map[string]interface{}{
			"kind": "structured", "rules": "", "path": "", "path_match": "exact", "target": "", "raw": "",
			"reason": in["reason"], "created_by": "admin", "updated_at": "2026-10-16T12:00:00Z",
		}
		for _, k := range []string{"kind", "rules", "path", "path_match", "target", "raw"} {
			if v, ok := in[k]; ok && v != nil {
				w[k] = v
			}
		}
		if id, ok := in["id"].(float64); ok {
			w["id"] = int(id)
		} else {
			m.nextWaf++
			w["id"] = m.nextWaf
		}
		out = append(out, w)
	}
	m.route["waf_exceptions"] = out
}

func (m *routeMock) render() map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range m.route {
		out[k] = v
	}
	settings := map[string]interface{}{}
	for k, v := range m.settings {
		settings[k] = v
	}
	out["settings"] = settings
	secrets := []string{}
	for k := range m.secrets {
		secrets = append(secrets, k)
	}
	sort.Strings(secrets)
	out["secrets_set"] = secrets
	out["waf_preview"] = []interface{}{}
	out["can_manage_raw_waf_rules"] = false
	return out
}

func (m *routeMock) handle(w http.ResponseWriter, r *http.Request, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/secureingress/route"), "/")
	isCert := strings.HasSuffix(rest, "/certificate")
	write := r.Method != http.MethodGet

	if write && m.conflict {
		w.WriteHeader(http.StatusConflict)
		_, _ = fmt.Fprint(w, `{"detail":"The ingress cluster has not been synced yet."}`)
		return
	}
	if rest != "" && (m.route == nil || m.deleted) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
		return
	}

	switch {
	case isCert && r.Method == http.MethodPost:
		r.Body = io.NopCloser(bytes.NewReader(body))
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			m.t.Errorf("certificate upload is not multipart: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f, _, err := r.FormFile("certificate")
		if err != nil {
			m.t.Errorf("certificate upload has no certificate file: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() { _ = f.Close() }()
		buf := make([]byte, 64)
		n, _ := f.Read(buf)
		if !strings.HasPrefix(string(buf[:n]), "-----BEGIN CERTIFICATE-----") {
			m.t.Errorf("unexpected certificate content %q", buf[:n])
		}
		if _, _, err := r.FormFile("privatekey"); err != nil {
			m.t.Errorf("certificate upload has no privatekey file: %v", err)
		}
		m.route["custom_certificate"] = map[string]interface{}{
			"id": "cert-001", "name": "shop.example.com", "common_name": "shop.example.com",
			"subject_alt_names": []string{"shop.example.com"}, "issuer": "Example CA", "serial_number": "01",
			"fingerprint_sha256": "AB:CD", "valid_from_timestamp": "2026-01-01T00:00:00Z",
			"valid_to_timestamp": "2027-01-01T00:00:00Z", "format": r.FormValue("format"),
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(m.route["custom_certificate"])
		return
	case isCert && r.Method == http.MethodDelete:
		m.route["custom_certificate"] = nil
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			m.t.Errorf("invalid JSON body: %v", err)
		}
	}

	switch {
	case r.Method == http.MethodGet && rest == "":
		results := []interface{}{}
		if m.route != nil && !m.deleted {
			results = append(results, m.render())
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"count": len(results), "next": nil, "results": results})
	case r.Method == http.MethodGet:
		_ = json.NewEncoder(w).Encode(m.render())
	case r.Method == http.MethodPost:
		m.route = map[string]interface{}{
			"id": "rt-001", "cluster": req["cluster"], "custom_certificate": nil,
			"locations": []interface{}{}, "waf_exceptions": []interface{}{},
			"path": "/", "ssl": true, "created_at_timestamp": "2026-10-16T12:00:00Z",
		}
		m.deleted = false
		fallthrough
	case r.Method == http.MethodPatch:
		if _, ok := req["cluster"]; ok && r.Method == http.MethodPatch {
			m.t.Errorf("PATCH must not send cluster")
		}
		m.route["server_name"] = strings.ToLower(fmt.Sprint(req["server_name"]))
		targetFields(m.route, req)
		for _, k := range []string{"path", "ssl"} {
			if v, ok := req[k]; ok {
				m.route[k] = v
			}
		}
		if s, ok := req["settings"].(map[string]interface{}); ok {
			m.applySettings(s)
		}
		if l, ok := req["locations"].([]interface{}); ok {
			m.applyLocations(l)
		}
		if wf, ok := req["waf_exceptions"].([]interface{}); ok {
			m.applyWaf(wf)
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_ = json.NewEncoder(w).Encode(m.render())
	case r.Method == http.MethodDelete:
		m.deleted = true
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
