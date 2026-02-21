// Camera-node: NMOS node with 3 camera (video) outputs.
// IS-04 Node API + IS-05 Connection API + SDP manifests.
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

	"github.com/google/uuid"
)

var (
	port     = flag.Int("port", 8180, "Node listen port (8180 = Studio B, avoids clash with virtualtest 8080)")
	nodeID   = flag.String("node-id", "", "Node ID (default: random UUID)")
	nodeLabel = flag.String("label", "Studio B", "Node label")
	is04Ver  = flag.String("is04", "v1.3", "IS-04 version")
	is05Ver  = flag.String("is05", "v1.0", "IS-05 version")
	hostname = flag.String("host", "localhost", "Host for API endpoints")
)

func getNodeID() string {
	if *nodeID != "" {
		return *nodeID
	}
	return uuid.New().String()
}

func main() {
	flag.Parse()
	if p := os.Getenv("PORT"); p != "" {
		var x int
		if _, err := fmt.Sscanf(p, "%d", &x); err == nil {
			*port = x
		}
	}
	if h := os.Getenv("HOST"); h != "" {
		*hostname = h
	}
	nid := getNodeID()
	deviceID := uuid.New().String()
	flowIDs := [3]string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
	senderIDs := [3]string{uuid.New().String(), uuid.New().String(), uuid.New().String()}

	baseURL := fmt.Sprintf("http://%s:%d", *hostname, *port)
	node := &Node{
		baseURL:   baseURL,
		nodeID:    nid,
		deviceID:  deviceID,
		flowIDs:   flowIDs,
		senderIDs: senderIDs,
	}
	node.initFlowsAndSenders()

	mux := http.NewServeMux()
	// IS-04 Node API
	mux.HandleFunc("/x-nmos/node/", node.serveNodeVersion)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/", node.serveNodeVersionList)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/self", node.serveSelf)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/self/", node.serveSelf)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/devices", node.serveDevices)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/devices/", node.serveDevices)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/flows", node.serveFlows)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/flows/", node.serveFlows)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/senders", node.serveSenders)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/senders/", node.serveSenders)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/receivers", node.serveReceivers)
	mux.HandleFunc("/x-nmos/node/"+*is04Ver+"/receivers/", node.serveReceivers)
	// IS-05 Connection API
	mux.HandleFunc("/x-nmos/connection/", node.serveConnVersion)
	mux.HandleFunc("/x-nmos/connection/"+*is05Ver+"/", node.serveConnVersionList)
	mux.HandleFunc("/x-nmos/connection/"+*is05Ver+"/single/senders", node.serveConnSenders)
	mux.HandleFunc("/x-nmos/connection/"+*is05Ver+"/single/senders/", node.serveConnSender)
	mux.HandleFunc("/x-nmos/connection/"+*is05Ver+"/single/receivers", node.serveConnReceivers)
	mux.HandleFunc("/x-nmos/connection/"+*is05Ver+"/single/receivers/", node.serveConnReceiver)
	// SDP
	mux.HandleFunc("/sdp/", node.serveSDP)
	mux.HandleFunc("/health", node.serveHealth)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("[camera-node] Node ID: %s", nid)
	log.Printf("[camera-node] Listening on %s", addr)
	log.Printf("[camera-node] IS-04: %s/ IS-05: %s/", baseURL+"/x-nmos/node/"+*is04Ver, baseURL+"/x-nmos/connection/"+*is05Ver)
	if err := http.ListenAndServe(addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

type Node struct {
	baseURL   string
	nodeID    string
	deviceID  string
	flowIDs   [3]string
	senderIDs [3]string
	flows     []map[string]interface{}
	senders   []map[string]interface{}
	devices   []map[string]interface{}
	receivers []map[string]interface{}
	connState struct {
		mu      sync.RWMutex
		senders map[string]connSenderState
	}
}

type connSenderState struct {
	MasterEnable  bool                   `json:"master_enable"`
	Activation    map[string]interface{} `json:"activation"`
	TransportFile map[string]interface{} `json:"transport_file"`
	ReceiverID    interface{}           `json:"receiver_id"`
}

func (n *Node) initFlowsAndSenders() {
	version := "1523456789:123"
	n.flows = make([]map[string]interface{}, 3)
	n.senders = make([]map[string]interface{}, 3)
	for i := 0; i < 3; i++ {
		cam := i + 1
		n.flows[i] = map[string]interface{}{
			"id":          n.flowIDs[i],
			"source_id":   n.senderIDs[i],
			"device_id":   n.deviceID,
			"parents":     []interface{}{},
			"format":      "urn:x-nmos:format:video",
			"media_type":  "video/raw",
			"label":       fmt.Sprintf("Camera %d", cam),
			"description": fmt.Sprintf("Mock camera %d video flow", cam),
			"tags":        map[string]interface{}{"site": []string{"CampusA"}, "room": []string{"Studio1"}},
			"version":     version + string(rune('0'+i)),
			"bit_rate":    1000000000,
			"frame_width": 1920,
			"frame_height": 1080,
			"interlace_mode": "progressive",
			"colorspace": "BT709",
			"transfer_characteristic": "SDR",
			"components": []map[string]interface{}{
				{"name": "Y", "width": 1920, "height": 1080, "bit_depth": 10},
			},
			"grain_rate": map[string]interface{}{"numerator": 25, "denominator": 1},
		}
		port := 5004 + i*2
		mcast := fmt.Sprintf("239.0.0.%d", 1+i)
		n.senders[i] = map[string]interface{}{
			"id":           n.senderIDs[i],
			"label":        fmt.Sprintf("Camera %d Out", cam),
			"description":  fmt.Sprintf("Mock camera %d sender", cam),
			"flow_id":      n.flowIDs[i],
			"device_id":    n.deviceID,
			"transport":    "urn:x-nmos:transport:rtp.mcast",
			"manifest_href": fmt.Sprintf("%s/sdp/cam%d.sdp", n.baseURL, cam),
			"tags":         map[string]interface{}{"site": []string{"CampusA"}, "room": []string{"Studio1"}},
			"version":      version + string(rune('0'+i+3)),
			"subscription": map[string]interface{}{
				"receiver_id": nil,
				"active":      false,
			},
		}
		_ = mcast
		_ = port
	}
	n.devices = []map[string]interface{}{
		{
			"id":          n.deviceID,
		"label":       "Studio B",
		"description": "Studio B – 3-camera device (Go)",
			"node_id":     n.nodeID,
			"type":        "urn:x-nmos:device:generic",
			"tags":        map[string]interface{}{"site": []string{"CampusA"}, "room": []string{"Studio1"}},
			"version":     "1523456789:123450",
			"controls": map[string]interface{}{
				"href": n.baseURL + "/x-nmos/connection/" + *is05Ver + "/",
			},
		},
	}
	n.receivers = []map[string]interface{}{}
	n.connState.senders = make(map[string]connSenderState)
	for _, s := range n.senders {
		sid, _ := s["id"].(string)
		n.connState.senders[sid] = connSenderState{
			Activation: map[string]interface{}{"mode": "activate_immediate", "requested_time": nil},
			TransportFile: map[string]interface{}{"data": "", "type": "application/sdp"},
		}
	}
}

func (n *Node) serveNodeVersion(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/x-nmos/node" || r.URL.Path == "/x-nmos/node/" {
		replyJSON(w, []string{*is04Ver})
		return
	}
	http.NotFound(w, r)
}

func (n *Node) serveNodeVersionList(w http.ResponseWriter, r *http.Request) {
	replyJSON(w, []string{*is04Ver})
}

func (n *Node) serveSelf(w http.ResponseWriter, r *http.Request) {
	replyJSON(w, map[string]interface{}{
		"id": n.nodeID, "label": *nodeLabel, "description": "3-camera NMOS node (Go)",
		"hostname": *hostname,
		"api": map[string]interface{}{
			"endpoints": []map[string]interface{}{
				{"host": *hostname, "port": *port, "protocol": "http"},
			},
			"versions": []string{*is04Ver},
		},
		"caps": map[string]interface{}{},
		"clocks": []map[string]interface{}{
			{"name": "clk0", "ref_type": "internal"},
		},
		"tags": map[string]interface{}{"site": []string{"CampusA"}, "room": []string{"Studio1"}},
		"version": "1523456789:123454",
	})
}

func (n *Node) serveDevices(w http.ResponseWriter, r *http.Request)  { replyJSON(w, n.devices) }
func (n *Node) serveFlows(w http.ResponseWriter, r *http.Request)    { replyJSON(w, n.flows) }
func (n *Node) serveSenders(w http.ResponseWriter, r *http.Request) { replyJSON(w, n.senders) }
func (n *Node) serveReceivers(w http.ResponseWriter, r *http.Request) { replyJSON(w, n.receivers) }

func (n *Node) serveConnVersion(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSuffix(r.URL.Path, "/") == "/x-nmos/connection" || strings.TrimSuffix(r.URL.Path, "/") == "/x-nmos/connection/" {
		replyJSON(w, []string{*is05Ver})
		return
	}
	http.NotFound(w, r)
}

func (n *Node) serveConnVersionList(w http.ResponseWriter, r *http.Request) {
	replyJSON(w, []string{*is05Ver})
}

func (n *Node) serveConnSenders(w http.ResponseWriter, r *http.Request) {
	n.connState.mu.RLock()
	defer n.connState.mu.RUnlock()
	var list []map[string]interface{}
	for _, s := range n.senders {
		sid, _ := s["id"].(string)
		st := n.connState.senders[sid]
		list = append(list, map[string]interface{}{
			"id":             sid,
			"master_enable": st.MasterEnable,
			"activation":    st.Activation,
			"transport_file": st.TransportFile,
			"receiver_id":   st.ReceiverID,
		})
	}
	replyJSON(w, list)
}

func (n *Node) serveConnSender(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/x-nmos/connection/"+*is05Ver+"/single/senders/")
	path = strings.Trim(path, "/")
	parts := strings.SplitN(path, "/", 2)
	senderID := parts[0]
	if senderID == "" {
		http.Error(w, "sender id required", http.StatusNotFound)
		return
	}
	var found bool
	for _, s := range n.senders {
		if s["id"] == senderID {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, `{"error":"Sender not found"}`, http.StatusNotFound)
		return
	}
	if len(parts) == 2 && parts[1] == "staged" {
		if r.Method == http.MethodPatch {
			var body map[string]interface{}
			if json.NewDecoder(r.Body).Decode(&body) == nil {
				n.connState.mu.Lock()
				st := n.connState.senders[senderID]
				if v, ok := body["master_enable"].(bool); ok {
					st.MasterEnable = v
				}
				if v, ok := body["activation"].(map[string]interface{}); ok {
					st.Activation = v
				}
				if v, ok := body["transport_file"].(map[string]interface{}); ok {
					st.TransportFile = v
				}
				if _, ok := body["receiver_id"]; ok {
					st.ReceiverID = body["receiver_id"]
				}
				n.connState.senders[senderID] = st
				n.connState.mu.Unlock()
			}
		}
		n.connState.mu.RLock()
		st := n.connState.senders[senderID]
		n.connState.mu.RUnlock()
		replyJSON(w, map[string]interface{}{
			"master_enable":  st.MasterEnable,
			"activation":     st.Activation,
			"transport_file": st.TransportFile,
			"receiver_id":    st.ReceiverID,
		})
		return
	}
	if len(parts) == 2 && (parts[1] == "active" || parts[1] == "active/") {
		n.connState.mu.RLock()
		st := n.connState.senders[senderID]
		n.connState.mu.RUnlock()
		replyJSON(w, map[string]interface{}{
			"master_enable":  st.MasterEnable,
			"activation":     st.Activation,
			"transport_file": st.TransportFile,
			"receiver_id":    st.ReceiverID,
		})
		return
	}
	http.NotFound(w, r)
}

func (n *Node) serveConnReceivers(w http.ResponseWriter, r *http.Request) {
	replyJSON(w, []map[string]interface{}{})
}

func (n *Node) serveConnReceiver(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// SDP for each camera: cam1.sdp, cam2.sdp, cam3.sdp
var sdpTpl = `v=0
o=- 0 0 IN IP4 %s
s=%s
t=0 0
m=video %d RTP/AVP 96
c=IN IP4 %s/32
a=source-filter: incl IN IP4 %s %s
a=rtpmap:96 raw/90000
a=rtcp:%d
a=sendrecv
a=ts-refclk:ptp=IEEE1588-2008:00-1B-63-FF-FE-FF-FF-FF
`

func (n *Node) serveSDP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/sdp/")
	name = strings.Trim(name, "/")
	var cam int
	var sourceIP, mcast string
	var port, rtcpPort int
	switch name {
	case "cam1.sdp":
		cam, sourceIP, mcast, port, rtcpPort = 1, "192.168.1.101", "239.0.0.1", 5004, 5005
	case "cam2.sdp":
		cam, sourceIP, mcast, port, rtcpPort = 2, "192.168.1.101", "239.0.0.2", 5006, 5007
	case "cam3.sdp":
		cam, sourceIP, mcast, port, rtcpPort = 3, "192.168.1.101", "239.0.0.3", 5008, 5009
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/sdp")
	title := fmt.Sprintf("Camera %d Stream", cam)
	body := fmt.Sprintf(sdpTpl, sourceIP, title, port, mcast, mcast, sourceIP, rtcpPort)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func (n *Node) serveHealth(w http.ResponseWriter, r *http.Request) {
	replyJSON(w, map[string]interface{}{
		"status":    "healthy",
		"node_id":   n.nodeID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func replyJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

