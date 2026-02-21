// RDS (Registration and Discovery Service) - IS-04 Query API
// Port 6062. Aggregates resources from configured NMOS node(s).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	port       = flag.Int("port", 6062, "RDS listen port")
	queryVer   = flag.String("query", "v1.3", "IS-04 Query API version")
	nodeURLs   = flag.String("nodes", "", "Comma-separated node base URLs (e.g. http://localhost:8080)")
	nodeURLsEnv = "RDS_NODE_URLS"
)

type resourceCache struct {
	mu      sync.RWMutex
	Nodes   []map[string]interface{} `json:"nodes"`
	Devices []map[string]interface{} `json:"devices"`
	Flows   []map[string]interface{} `json:"flows"`
	Senders []map[string]interface{} `json:"senders"`
	Receivers []map[string]interface{} `json:"receivers"`
}

var cache resourceCache

func getNodeURLs() []string {
	if v := os.Getenv(nodeURLsEnv); v != "" {
		return strings.Split(strings.TrimSpace(v), ",")
	}
	if *nodeURLs != "" {
		return strings.Split(strings.TrimSpace(*nodeURLs), ",")
	}
	return []string{"http://localhost:8180"}
}

func fetchAndDecode(client *http.Client, url string, v interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func refreshCache() {
	urls := getNodeURLs()
	client := &http.Client{Timeout: 5 * time.Second}
	var nodes, devices, flows, senders, receivers []map[string]interface{}
	base := "http://localhost:8080"
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		base = strings.TrimRight(u, "/")
		version := *queryVer
		// Node self
		var nodeSelf map[string]interface{}
		if err := fetchAndDecode(client, base+"/x-nmos/node/"+version+"/self", &nodeSelf); err != nil {
			log.Printf("[RDS] skip node %s: %v", base, err)
			continue
		}
		nodeSelf["base_url"] = base
		nodeSelf["href"] = base + "/x-nmos/node/" + version
		nodes = append(nodes, nodeSelf)
		// Devices
		var devList []map[string]interface{}
		if err := fetchAndDecode(client, base+"/x-nmos/node/"+version+"/devices", &devList); err == nil {
			for _, d := range devList {
				d["base_url"] = base
				devices = append(devices, d)
			}
		}
		// Flows
		var flowList []map[string]interface{}
		if err := fetchAndDecode(client, base+"/x-nmos/node/"+version+"/flows", &flowList); err == nil {
			for _, f := range flowList {
				f["base_url"] = base
				flows = append(flows, f)
			}
		}
		// Senders
		var senderList []map[string]interface{}
		if err := fetchAndDecode(client, base+"/x-nmos/node/"+version+"/senders", &senderList); err == nil {
			for _, s := range senderList {
				s["base_url"] = base
				senders = append(senders, s)
			}
		}
		// Receivers
		var recvList []map[string]interface{}
		if err := fetchAndDecode(client, base+"/x-nmos/node/"+version+"/receivers", &recvList); err == nil {
			for _, r := range recvList {
				r["base_url"] = base
				receivers = append(receivers, r)
			}
		}
	}
	cache.mu.Lock()
	cache.Nodes = nodes
	cache.Devices = devices
	cache.Flows = flows
	cache.Senders = senders
	cache.Receivers = receivers
	cache.mu.Unlock()
	log.Printf("[RDS] refreshed: %d nodes, %d devices, %d flows, %d senders, %d receivers",
		len(nodes), len(devices), len(flows), len(senders), len(receivers))
	_ = base
}

func jsonReply(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func queryList(resource string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshCache()
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		var list []map[string]interface{}
		switch resource {
		case "nodes":
			list = cache.Nodes
		case "devices":
			list = cache.Devices
		case "flows":
			list = cache.Flows
		case "senders":
			list = cache.Senders
		case "receivers":
			list = cache.Receivers
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if list == nil {
			list = []map[string]interface{}{}
		}
		jsonReply(w, list)
	}
}

func queryOneByID(w http.ResponseWriter, _ *http.Request, resource, id string) {
	refreshCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	var list []map[string]interface{}
	switch resource {
	case "nodes":
		list = cache.Nodes
	case "devices":
		list = cache.Devices
	case "flows":
		list = cache.Flows
	case "senders":
		list = cache.Senders
	case "receivers":
		list = cache.Receivers
	default:
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	for _, m := range list {
		if sid, _ := m["id"].(string); sid == id {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(m)
			return
		}
	}
	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	qv := *queryVer
	// Query API
	mux.HandleFunc("/x-nmos/query/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/x-nmos/query" || r.URL.Path == "/x-nmos/query/" {
			jsonReply(w, []string{qv})
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/x-nmos/query/"+qv+"/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/x-nmos/query/"+qv+"/")
		rest = strings.Trim(rest, "/")
		if rest == "" {
			jsonReply(w, []string{qv})
			return
		}
		parts := strings.SplitN(rest, "/", 2)
		switch len(parts) {
		case 1:
			switch parts[0] {
			case "nodes", "devices", "flows", "senders", "receivers":
				queryList(parts[0])(w, r)
				return
			}
		case 2:
			resource, id := parts[0], parts[1]
			switch resource {
			case "nodes", "devices", "flows", "senders", "receivers":
				queryOneByID(w, r, resource, id)
				return
			}
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/x-nmos/query/"+qv+"/nodes", queryList("nodes"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/nodes/", queryList("nodes"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/devices", queryList("devices"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/devices/", queryList("devices"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/flows", queryList("flows"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/flows/", queryList("flows"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/senders", queryList("senders"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/senders/", queryList("senders"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/receivers", queryList("receivers"))
	mux.HandleFunc("/x-nmos/query/"+qv+"/receivers/", queryList("receivers"))
	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		refreshCache()
		cache.mu.RLock()
		n, d, f, s, rec := len(cache.Nodes), len(cache.Devices), len(cache.Flows), len(cache.Senders), len(cache.Receivers)
		cache.mu.RUnlock()
		jsonReply(w, map[string]interface{}{
			"status": "healthy",
			"port":   *port,
			"resources": map[string]int{"nodes": n, "devices": d, "flows": f, "senders": s, "receivers": rec},
		})
	})
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("[RDS] Listening on %s (Query API %s)", addr, qv)
	log.Printf("[RDS] Node URLs: %v", getNodeURLs())
	refreshCache()
	if err := http.ListenAndServe(addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
