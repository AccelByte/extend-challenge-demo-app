// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package events

import "context"

// EventTrigger handles triggering gameplay events for testing challenge progress.
//
// This interface provides a unified API for triggering events in different modes:
//   - Local Mode: Calls event handler gRPC services directly (for local development)
//   - AGS Mode: Publishes events to AGS Event Bus via Kafka (for AGS-deployed services)
//
// The mode is determined at creation time via factory function.
type EventTrigger interface {
	// TriggerLogin simulates a user login event.
	//
	// This triggers challenge goals with event_source="login" in the event handler.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - userID: AccelByte user ID
	//   - namespace: AccelByte namespace
	//
	// Returns:
	//   - error: Non-nil if event trigger failed (connection, validation, processing)
	TriggerLogin(ctx context.Context, userID, namespace string) error

	// TriggerStatUpdate simulates a statistic update event.
	//
	// This triggers challenge goals tracking the specified stat_code.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - userID: AccelByte user ID
	//   - namespace: AccelByte namespace
	//   - statCode: Stat code identifier (e.g., "kills", "headshots")
	//   - value: New stat value (absolute value, not increment)
	//
	// Returns:
	//   - error: Non-nil if event trigger failed (connection, validation, processing)
	TriggerStatUpdate(ctx context.Context, userID, namespace, statCode string, value int) error

	// Close cleans up resources (gRPC connection, Kafka writer).
	//
	// Should be called when the EventTrigger is no longer needed.
	//
	// Returns:
	//   - error: Non-nil if cleanup failed
	Close() error
}
