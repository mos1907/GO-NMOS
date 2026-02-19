package sdp

import (
	"strconv"
	"strings"
)

// ParsedDetails holds parsed SDP fields (RFC 4566 / ST 2110).
type ParsedDetails struct {
	MediaType       string `json:"media_type,omitempty"`
	GroupPortA      int    `json:"group_port_a,omitempty"`
	MulticastAddrA  string `json:"multicast_addr_a,omitempty"`
	SourceAddrA     string `json:"source_addr_a,omitempty"`
	RedundancyGroup string `json:"redundancy_group,omitempty"`
}

// ParseSDP parses SDP text and extracts common ST 2110 / NMOS fields.
// Supports: m= (media), c= (connection), a=source-filter, a=rtpmap, a=group
func ParseSDP(sdpText string) ParsedDetails {
	var result ParsedDetails
	if sdpText == "" {
		return result
	}
	for _, rawLine := range strings.Split(sdpText, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		idx := strings.Index(line, "=")
		prefix := line[:idx]
		rest := strings.TrimSpace(line[idx+1:])
		switch prefix {
		case "m":
			parts := strings.Fields(rest)
			if len(parts) >= 1 && result.MediaType == "" {
				result.MediaType = parts[0]
			}
			if len(parts) >= 2 && result.GroupPortA == 0 {
				if port, err := strconv.Atoi(parts[1]); err == nil {
					result.GroupPortA = port
				}
			}
		case "c":
			parts := strings.Fields(rest)
			if len(parts) >= 1 && result.MulticastAddrA == "" {
				addr := parts[len(parts)-1]
				if idx := strings.Index(addr, "/"); idx >= 0 {
					addr = addr[:idx]
				}
				result.MulticastAddrA = addr
			}
		case "a":
			if strings.HasPrefix(rest, "source-filter:") {
				_, content, _ := cut(rest, ":")
				tokens := strings.Fields(strings.TrimSpace(content))
				if len(tokens) >= 5 {
					maddr := tokens[3]
					src := tokens[4]
					if idx := strings.Index(maddr, "/"); idx >= 0 {
						maddr = maddr[:idx]
					}
					if result.SourceAddrA == "" {
						result.SourceAddrA = src
					}
					if result.MulticastAddrA == "" {
						result.MulticastAddrA = maddr
					}
				}
			} else if strings.HasPrefix(rest, "rtpmap:") {
				parts := strings.Fields(rest)
				if len(parts) >= 2 && result.MediaType == "" {
					enc := parts[1]
					if idx := strings.Index(enc, "/"); idx >= 0 {
						enc = enc[:idx]
					}
					result.MediaType = enc
				}
			} else if strings.HasPrefix(rest, "group:") && result.RedundancyGroup == "" {
				result.RedundancyGroup = rest
			}
		}
	}
	return result
}

func cut(s, sep string) (before, after string, ok bool) {
	if idx := strings.Index(s, sep); idx >= 0 {
		return s[:idx], s[idx+len(sep):], true
	}
	return s, "", false
}
