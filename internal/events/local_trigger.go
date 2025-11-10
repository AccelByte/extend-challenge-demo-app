// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package events

import (
	"context"
	"fmt"
	"time"

	accountpb "extend-challenge-event-handler/pkg/pb/accelbyte-asyncapi/iam/account/v1"
	statpb "extend-challenge-event-handler/pkg/pb/accelbyte-asyncapi/social/statistic/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// LocalEventTrigger triggers events by calling the event handler's gRPC services directly.
//
// This implementation is intended for local development and testing. It calls the event
// handler's OnMessage RPCs directly with AGS-compatible event payloads.
//
// Thread Safety: This implementation is safe for concurrent use.
type LocalEventTrigger struct {
	conn          *grpc.ClientConn
	loginClient   accountpb.UserAuthenticationUserLoggedInServiceClient
	statClient    statpb.StatisticStatItemUpdatedServiceClient
	eventHandlerAddr string
}

// NewLocalEventTrigger creates a new LocalEventTrigger that connects to the event handler.
//
// Parameters:
//   - eventHandlerAddr: Event handler gRPC address (e.g., "localhost:6565")
//
// Returns:
//   - *LocalEventTrigger: Ready-to-use event trigger
//   - error: Non-nil if connection to event handler failed
func NewLocalEventTrigger(eventHandlerAddr string) (*LocalEventTrigger, error) {
	if eventHandlerAddr == "" {
		return nil, fmt.Errorf("event handler address cannot be empty")
	}

	// Connect to event handler with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		eventHandlerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to event handler at %s: %w", eventHandlerAddr, err)
	}

	// Create gRPC clients for each event type
	loginClient := accountpb.NewUserAuthenticationUserLoggedInServiceClient(conn)
	statClient := statpb.NewStatisticStatItemUpdatedServiceClient(conn)

	return &LocalEventTrigger{
		conn:             conn,
		loginClient:      loginClient,
		statClient:       statClient,
		eventHandlerAddr: eventHandlerAddr,
	}, nil
}

// TriggerLogin triggers a login event by calling the event handler's OnMessage RPC.
//
// This constructs a UserLoggedIn message and sends it to the event handler, which will
// process it exactly as if it came from the AGS Event Bus via Kafka.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: AccelByte user ID
//   - namespace: AccelByte namespace
//
// Returns:
//   - error: Non-nil if event trigger failed
func (t *LocalEventTrigger) TriggerLogin(ctx context.Context, userID, namespace string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	// Construct UserLoggedIn message matching AGS event format
	msg := &accountpb.UserLoggedIn{
		Id:        generateEventID(),
		UserId:    userID,
		Namespace: namespace,
	}

	// Call OnMessage RPC
	_, err := t.loginClient.OnMessage(ctx, msg)
	if err != nil {
		// Extract gRPC error details
		st := status.Convert(err)
		return fmt.Errorf("trigger login event failed: %s: %w", st.Message(), err)
	}

	return nil
}

// TriggerStatUpdate triggers a statistic update event by calling the event handler's OnMessage RPC.
//
// This constructs a StatItemUpdated message and sends it to the event handler, which will
// process it exactly as if it came from the AGS Event Bus via Kafka.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: AccelByte user ID
//   - namespace: AccelByte namespace
//   - statCode: Stat code identifier (e.g., "kills", "headshots")
//   - value: New stat value (absolute value, not increment)
//
// Returns:
//   - error: Non-nil if event trigger failed
func (t *LocalEventTrigger) TriggerStatUpdate(ctx context.Context, userID, namespace, statCode string, value int) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	if statCode == "" {
		return fmt.Errorf("statCode cannot be empty")
	}

	// Construct StatItemUpdated message matching AGS event format
	// Note: StatCode and LatestValue are in the Payload field
	msg := &statpb.StatItemUpdated{
		Id:        generateEventID(),
		UserId:    userID,
		Namespace: namespace,
		Payload: &statpb.StatItem{
			StatCode:    statCode,
			LatestValue: float64(value),
		},
	}

	// Call OnMessage RPC
	_, err := t.statClient.OnMessage(ctx, msg)
	if err != nil {
		// Extract gRPC error details
		st := status.Convert(err)
		return fmt.Errorf("trigger stat update event failed: %s: %w", st.Message(), err)
	}

	return nil
}

// Close closes the gRPC connection to the event handler.
//
// Returns:
//   - error: Non-nil if closing the connection failed
func (t *LocalEventTrigger) Close() error {
	if t.conn == nil {
		return nil
	}

	err := t.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close event handler connection: %w", err)
	}

	return nil
}

// generateEventID generates a unique event ID for testing.
//
// Returns:
//   - string: Unique event ID (format: "demo-event-{unix_nano}")
func generateEventID() string {
	return fmt.Sprintf("demo-event-%d", time.Now().UnixNano())
}
