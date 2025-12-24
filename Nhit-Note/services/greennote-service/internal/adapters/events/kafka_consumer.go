package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"nhit-note/services/greennote-service/internal/core/ports"
	greennotepb "nhit-note/api/pb/greennotepb"

	"github.com/segmentio/kafka-go"
)

// ApprovalEvent represents the structure of events received from the Approval Service.
type ApprovalEvent struct {
	SourceID   string `json:"source_id"`   // The GreenNote ID
	SourceType string `json:"source_type"` // "GREENNOTE"
	Status     string `json:"status"`      // "APPROVED", "REJECTED", "PENDING_LEVEL_2"
	ActorID    string `json:"actor_id"`
	Comments   string `json:"comments"`
}

// KafkaConsumer listens to approval events and updates GreenNote status.
type KafkaConsumer struct {
	reader *kafka.Reader
	repo   ports.GreenNoteRepository
	events ports.EventPublisher
}

// NewKafkaConsumer creates a new consumer for the given brokers and topic.
func NewKafkaConsumer(brokers []string, topic string, groupID string, repo ports.GreenNoteRepository, events ports.EventPublisher) *KafkaConsumer {
	if len(brokers) == 0 || topic == "" {
		return nil
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader: r,
		repo:   repo,
		events: events,
	}
}

// Start begins consuming messages in a blocking loop. 
// It should be run in a goroutine.
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("🚀 Starting Kafka Consumer for approval events...")
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
			log.Printf("⚠️ Error reading Kafka message: %v", err)
			// Backoff slightly to avoid busy loop on transient errors
			time.Sleep(time.Second)
			continue
		}

		c.handleMessage(ctx, m)
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, m kafka.Message) {
	// 1. Parse the Event Wrapper
	var eventWrapper struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"` // Delay parsing the inner payload
	}
	
	if err := json.Unmarshal(m.Value, &eventWrapper); err != nil {
		log.Printf("❌ Failed to unmarshal event wrapper: %v", err)
		return
	}

	// 2. Filter: Only care about Status Changes
	if eventWrapper.EventType != "ApprovalStatusChanged" {
		return // Ignore other events like "VoteSubmitted"
	}

	// 3. Parse the Data Payload
	var payload ApprovalEvent
	if err := json.Unmarshal(eventWrapper.Data, &payload); err != nil {
		log.Printf("❌ Failed to unmarshal inner payload: %v", err)
		return
	}

	// 4. Filter: Is this for ME? (GreenNote Service)
	if payload.SourceType != "GREENNOTE" {
		return // Ignore "BANK_LETTER" or other types
	}

	log.Printf("🔔 Received Approval Update for GreenNote %s: %s", payload.SourceID, payload.Status)

	// 5. Update the Database
	// This transitions the Green Note from PENDING -> APPROVED/REJECTED
	
	// We need to fetch the existing note to preserve other fields, then update status/comments.
	// OR use a specific repository method if available. Repo.Update takes full payload.
	greenNoteId := payload.SourceID
	existing, orgID, tenantID, err := c.repo.Get(ctx, greenNoteId)
	if err != nil {
		log.Printf("⚠️ Failed to find GreenNote %s: %v", greenNoteId, err)
		return 
	}

	// Update fields
	existing.DetailedStatus = payload.Status
	// Append comments to remarks or handle as needed. 
	// User instruction: "UpdateStatusAndComments(ctx, payload.SourceID, payload.Status, payload.Comments)"
	// But our repo interface (seen via GreennoteService) uses generic Update usually. 
	// Let's see if we can just update the note.
	if payload.Comments != "" {
		if existing.Remarks != "" {
			existing.Remarks += "; " + payload.Comments
		} else {
			existing.Remarks = payload.Comments
		}
	}

	if err := c.repo.Update(ctx, greenNoteId, existing, orgID, tenantID); err != nil {
		log.Printf("❌ Failed to update GreenNote status: %v", err)
		return // Return error to Kafka? We are in a handler, can't easily Nack here without changing signature.
	}

	// 6. Trigger Downstream Logic (If Approved)
	if payload.Status == "APPROVED" {
		// Publish GreenNoteApprovedEvent for downstream consumers (like Payment Note Service)
		c.publishApprovedEvent(ctx, greenNoteId, existing, payload)
	}
}

func (c *KafkaConsumer) publishApprovedEvent(ctx context.Context, id string, note *greennotepb.GreenNotePayload, approvalEvent ApprovalEvent) {
	evt := ports.GreenNoteApprovedEvent{
		GreenNoteID: id,
		OrderNo:     fmt.Sprintf("GN-%s", id),
		NetAmount:   note.TotalAmount,
		Status:      "approved",
		Comments:    approvalEvent.Comments,
		ApprovedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	
	if err := c.events.PublishGreenNoteApproved(ctx, evt); err != nil {
		log.Printf("❌ Failed to publish GreenNoteApprovedEvent: %v", err)
	} else {
		log.Printf("📢 Published GreenNoteApprovedEvent for %s", id)
	}
}
