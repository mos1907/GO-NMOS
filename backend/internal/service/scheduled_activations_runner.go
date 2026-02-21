package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// ScheduledActivationsRunner executes pending scheduled activations when their time arrives.
type ScheduledActivationsRunner struct {
	repo scheduledActivationsRepo
}

type scheduledActivationsRepo interface {
	ListPendingScheduledActivations(ctx context.Context, before time.Time) ([]models.ScheduledActivation, error)
	GetFlowByID(ctx context.Context, id int64) (*models.Flow, error)
	ListNMOSReceivers(ctx context.Context, deviceID string) ([]models.NMOSReceiver, error)
	UpdateScheduledActivation(ctx context.Context, id int64, updates map[string]any) error
	UpsertReceiverConnection(ctx context.Context, conn models.ReceiverConnection) error
	RecordReceiverConnectionHistory(ctx context.Context, hist models.ReceiverConnectionHistory) error
}

func NewScheduledActivationsRunner(repo scheduledActivationsRepo) *ScheduledActivationsRunner {
	return &ScheduledActivationsRunner{repo: repo}
}

func (r *ScheduledActivationsRunner) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runCycle(ctx)
		}
	}
}

func (r *ScheduledActivationsRunner) runCycle(ctx context.Context) {
	now := time.Now()
	pending, err := r.repo.ListPendingScheduledActivations(ctx, now)
	if err != nil {
		log.Printf("scheduled activations runner: list pending failed: %v", err)
		return
	}

	for _, act := range pending {
		r.executeActivation(ctx, act)
	}
}

func (r *ScheduledActivationsRunner) executeActivation(ctx context.Context, act models.ScheduledActivation) {
	log.Printf("scheduled activations runner: executing activation id=%d scheduled_at=%s", act.ID, act.ScheduledAt.Format(time.RFC3339))

	flow, err := r.repo.GetFlowByID(ctx, act.FlowID)
	if err != nil {
		r.markFailed(ctx, act.ID, fmt.Sprintf("flow not found: %v", err))
		return
	}

	// Order receivers for safe_switch mode
	receiverIDs := act.ReceiverIDs
	if act.Mode == "safe_switch" {
		allRecvs, _ := r.repo.ListNMOSReceivers(ctx, "")
		formatByID := make(map[string]string)
		for _, rec := range allRecvs {
			formatByID[rec.ID] = strings.ToLower(rec.Format)
		}
		formatOrder := map[string]int{
			"audio": 0,
			"video": 1,
			"data":  2,
			"mux":   3,
		}
		sort.Slice(receiverIDs, func(i, j int) bool {
			ordI, okI := formatOrder[formatByID[receiverIDs[i]]]
			ordJ, okJ := formatOrder[formatByID[receiverIDs[j]]]
			if !okI {
				ordI = 99
			}
			if !okJ {
				ordJ = 99
			}
			return ordI < ordJ
		})
	}

	results := make([]map[string]any, 0, len(receiverIDs))
	client := &http.Client{Timeout: 8 * time.Second}
	success, failed := 0, 0

	for _, receiverID := range receiverIDs {
		connectionURL := strings.TrimSuffix(act.IS05BaseURL, "/") + "/single/receivers/" + receiverID + "/staged"
		parsedURL, err := url.Parse(connectionURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			results = append(results, map[string]any{"receiver_id": receiverID, "ok": false, "error": "invalid connection URL"})
			failed++
			continue
		}

		patchBody := map[string]any{
			"activation":    map[string]any{"mode": "activate_immediate"},
			"master_enable": true,
			"transport_params": []map[string]any{
				{
					"multicast_ip":     flow.MulticastIP,
					"source_ip":        flow.SourceIP,
					"destination_port": flow.Port,
					"rtp_enabled":      true,
				},
			},
		}
		if act.SenderID != "" {
			patchBody["sender_id"] = act.SenderID
		}

		payload, _ := json.Marshal(patchBody)
		httpReq, _ := http.NewRequest(http.MethodPatch, connectionURL, strings.NewReader(string(payload)))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		if err != nil {
			results = append(results, map[string]any{"receiver_id": receiverID, "ok": false, "error": "patch request failed: " + err.Error()})
			failed++
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			results = append(results, map[string]any{"receiver_id": receiverID, "ok": false, "error": fmt.Sprintf("non-2xx status: %d", resp.StatusCode)})
			failed++
			continue
		}

		// Record connection state
		conn := models.ReceiverConnection{
			ReceiverID: receiverID,
			State:      "staged",
			Role:       "master",
			SenderID:   act.SenderID,
			FlowID:     &act.FlowID,
			ChangedAt:  time.Now(),
			ChangedBy:  act.CreatedBy,
		}
		_ = r.repo.UpsertReceiverConnection(ctx, conn)
		hist := models.ReceiverConnectionHistory{
			ReceiverID: receiverID,
			State:      "staged",
			Role:       "master",
			SenderID:   act.SenderID,
			FlowID:     &act.FlowID,
			ChangedAt:  time.Now(),
			Action:     "connect",
			ChangedBy:  act.CreatedBy,
		}
		_ = r.repo.RecordReceiverConnectionHistory(ctx, hist)

		results = append(results, map[string]any{"receiver_id": receiverID, "ok": true})
		success++
	}

	resultJSON, _ := json.Marshal(map[string]any{
		"success": success,
		"failed":  failed,
		"results": results,
	})

	status := "executed"
	if failed > 0 && success == 0 {
		status = "failed"
	}

	now := time.Now()
	updates := map[string]any{
		"status":      status,
		"executed_at": now,
		"result":      string(resultJSON),
	}
	if err := r.repo.UpdateScheduledActivation(ctx, act.ID, updates); err != nil {
		log.Printf("scheduled activations runner: failed to update activation id=%d: %v", act.ID, err)
		return
	}

	log.Printf("scheduled activations runner: completed activation id=%d success=%d failed=%d", act.ID, success, failed)
}

func (r *ScheduledActivationsRunner) markFailed(ctx context.Context, id int64, errorMsg string) {
	resultJSON, _ := json.Marshal(map[string]any{"error": errorMsg})
	updates := map[string]any{
		"status":      "failed",
		"executed_at": time.Now(),
		"result":      string(resultJSON),
	}
	if err := r.repo.UpdateScheduledActivation(ctx, id, updates); err != nil {
		log.Printf("scheduled activations runner: failed to mark activation id=%d as failed: %v", id, err)
	}
}
