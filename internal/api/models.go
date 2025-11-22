// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package api

import "time"

// Challenge represents a challenge with goals and user progress
// Matches the protobuf Challenge message from backend service (uses protojson camelCase)
type Challenge struct {
	ID          string `json:"challengeId"` // Backend uses camelCase via protojson
	Name        string `json:"name"`
	Description string `json:"description"`
	Goals       []Goal `json:"goals"`
}

// Goal represents a single goal within a challenge
// Matches the protobuf Goal message from backend service (uses protojson camelCase)
type Goal struct {
	ID            string      `json:"goalId"` // Backend uses camelCase via protojson
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Requirement   Requirement `json:"requirement"`
	Reward        Reward      `json:"reward"`
	Prerequisites []string    `json:"prerequisites"` // Array of prerequisite goal IDs
	// Progress fields are embedded directly in Goal (not a nested object)
	Progress    int32  `json:"progress"`    // Current progress value
	Status      string `json:"status"`      // "not_started", "in_progress", "completed", "claimed"
	Locked      bool   `json:"locked"`      // Whether goal is locked by prerequisites
	CompletedAt string `json:"completedAt"` // RFC3339 timestamp or empty string (camelCase)
	ClaimedAt   string `json:"claimedAt"`   // RFC3339 timestamp or empty string (camelCase)
	IsActive    bool   `json:"isActive"`    // Whether goal is currently active (M3/M4 feature)
}

// Requirement specifies what is needed to complete a goal
// Matches the protobuf Requirement message from backend service (uses protojson camelCase)
type Requirement struct {
	StatCode    string `json:"statCode"`    // Stat code to check (camelCase)
	Operator    string `json:"operator"`    // "gte", "lte", "eq"
	TargetValue int32  `json:"targetValue"` // Target value (camelCase)
}

// Reward specifies what the user gets for completing a goal
// Matches the protobuf Reward message from backend service (uses protojson camelCase)
type Reward struct {
	Type     string `json:"type"`     // "ITEM" or "WALLET"
	RewardID string `json:"rewardId"` // Backend uses camelCase via protojson (item ID or wallet code)
	Quantity int32  `json:"quantity"` // Amount
}

// GetChallengesResponse wraps the list of challenges returned by the API
// Matches the protobuf GetChallengesResponse message from backend service
type GetChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
}

// ClaimResult represents the result of a claim operation
// Matches the protobuf ClaimRewardResponse message from backend service (uses protojson camelCase)
type ClaimResult struct {
	GoalID    string `json:"goalId"`    // Backend uses camelCase via protojson
	Status    string `json:"status"`
	Reward    Reward `json:"reward"`
	ClaimedAt string `json:"claimedAt"` // Backend uses camelCase via protojson
}

// M3: InitializeResponse represents the response from initializing player goals
// Matches the protobuf InitializePlayerResponse message from backend service
type InitializeResponse struct {
	AssignedGoals  []AssignedGoal `json:"assignedGoals"`
	NewAssignments int32          `json:"newAssignments"`
	TotalActive    int32          `json:"totalActive"`
}

// M3: AssignedGoal represents a goal that has been assigned to the player
type AssignedGoal struct {
	ChallengeID string `json:"challengeId"`
	GoalID      string `json:"goalId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
	AssignedAt  string `json:"assignedAt"` // RFC3339 timestamp
	ExpiresAt   string `json:"expiresAt"`  // RFC3339 timestamp or empty string
	Progress    int32  `json:"progress"`
	Target      int32  `json:"target"`
	Status      string `json:"status"`
}

// M3: SetGoalActiveResponse represents the response from activating/deactivating a goal
// Matches the protobuf SetGoalActiveResponse message from backend service
type SetGoalActiveResponse struct {
	ChallengeID string `json:"challengeId"`
	GoalID      string `json:"goalId"`
	IsActive    bool   `json:"isActive"`
	AssignedAt  string `json:"assignedAt"` // RFC3339 timestamp
	Message     string `json:"message"`
}

// RequestDebugInfo stores debug information about a request
type RequestDebugInfo struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// ResponseDebugInfo stores debug information about a response
type ResponseDebugInfo struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	Duration   time.Duration
}

// M4: BatchSelectRequest represents the request for batch goal selection
type BatchSelectRequest struct {
	GoalIDs         []string `json:"goal_ids"`
	ReplaceExisting bool     `json:"replace_existing"`
}

// M4: BatchSelectResponse represents the response from batch goal selection
type BatchSelectResponse struct {
	SelectedGoals    []Goal   `json:"selectedGoals"`
	ChallengeID      string   `json:"challengeId"`
	TotalActiveGoals int32    `json:"totalActiveGoals"`
	ReplacedGoals    []string `json:"replacedGoals"`
}

// M4: RandomSelectRequest represents the request for random goal selection
type RandomSelectRequest struct {
	Count           int  `json:"count"`
	ReplaceExisting bool `json:"replace_existing"`
	ExcludeActive   bool `json:"exclude_active"`
}

// M4: RandomSelectResponse represents the response from random goal selection
type RandomSelectResponse struct {
	SelectedGoals    []Goal   `json:"selectedGoals"`
	ChallengeID      string   `json:"challengeId"`
	TotalActiveGoals int32    `json:"totalActiveGoals"`
	ReplacedGoals    []string `json:"replacedGoals"`
}
