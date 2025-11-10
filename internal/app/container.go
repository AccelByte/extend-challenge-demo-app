// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	sdkAuth "github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/ags"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/api"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/auth"
	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/events"
)

// Container holds all application dependencies
type Container struct {
	AuthProvider      auth.AuthProvider
	AdminAuthProvider auth.AuthProvider // Optional: for AGS Platform verification
	APIClient         api.APIClient
	EventTrigger      events.EventTrigger
	RewardVerifier    ags.RewardVerifier
	UserID            string
	Namespace         string
}

// extractUserIDFromJWT extracts the user ID from a JWT token's "sub" claim
// Returns empty string if extraction fails
func extractUserIDFromJWT(token string) string {
	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Printf("Warning: Invalid JWT format (expected 3 parts, got %d)", len(parts))
		return ""
	}

	// Decode the payload (second part)
	payload := parts[1]

	// Add padding if necessary (base64 requires padding)
	if m := len(payload) % 4; m != 0 {
		payload += strings.Repeat("=", 4-m)
	}

	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		log.Printf("Warning: Failed to decode JWT payload: %v", err)
		return ""
	}

	// Parse JSON to extract "sub" claim
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		log.Printf("Warning: Failed to parse JWT claims: %v", err)
		return ""
	}

	return claims.Sub
}

// NewContainer creates a new dependency container
func NewContainer(
	backendURL string,
	authMode string,
	eventHandlerURL string,
	userID string,
	namespace string,
	email string,
	password string,
	clientID string,
	clientSecret string,
	iamURL string,
	platformURL string,
	adminClientID string,
	adminClientSecret string,
) *Container {
	// Create auth provider based on mode
	var authProvider auth.AuthProvider

	switch authMode {
	case "password":
		// User authentication (email + password → user token)
		// RECOMMENDED for Challenge Service API testing
		authProvider = auth.NewPasswordAuthProvider(
			iamURL,
			clientID,
			clientSecret,
			namespace,
			email,
			password,
		)

		// Extract user ID from JWT token
		// This is critical - the --user-id flag should NOT be used in password mode
		ctx := context.Background()
		token, err := authProvider.GetToken(ctx)
		if err != nil {
			log.Printf("Warning: Failed to authenticate with password: %v", err)
			log.Printf("Falling back to --user-id flag value: %s", userID)
		} else {
			extractedUserID := extractUserIDFromJWT(token.AccessToken)
			if extractedUserID != "" {
				log.Printf("Extracted user ID from JWT token: %s", extractedUserID)
				userID = extractedUserID // Override the flag value with JWT's user ID
			} else {
				log.Printf("Warning: Failed to extract user ID from JWT, using --user-id flag: %s", userID)
			}
		}

	case "client":
		// Service authentication (client credentials → service token)
		// WARNING: Service token does NOT have user_id!
		// TODO: Implement ClientAuthProvider in future
		log.Printf("WARNING: client mode not yet implemented, falling back to mock mode")
		authProvider = auth.NewMockAuthProvider(userID, namespace)

	case "mock":
		// Mock authentication with configurable user_id
		authProvider = auth.NewMockAuthProvider(userID, namespace)

	default:
		// Default to mock mode
		log.Printf("Unknown auth mode '%s', defaulting to mock", authMode)
		authProvider = auth.NewMockAuthProvider(userID, namespace)
	}

	// Create admin auth provider (optional - for AGS Platform verification)
	var adminAuthProvider auth.AuthProvider
	if adminClientID != "" && adminClientSecret != "" {
		if iamURL == "" {
			log.Printf("Warning: Admin credentials provided but IAM URL is empty")
		} else {
			adminAuthProvider = auth.NewClientAuthProvider(
				iamURL,
				adminClientID,
				adminClientSecret,
				namespace,
			)
			log.Printf("Admin auth provider initialized for AGS Platform verification")
		}
	}

	// Create API client
	apiClient := api.NewHTTPAPIClient(backendURL, authProvider)
	// Set user ID for mock authentication header (used when backend auth is disabled)
	apiClient.SetUserID(userID)

	// Create event trigger (optional - only if event handler URL provided)
	var eventTrigger events.EventTrigger
	if eventHandlerURL != "" {
		var err error
		eventTrigger, err = events.NewLocalEventTrigger(eventHandlerURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to event handler at %s: %v", eventHandlerURL, err)
			log.Printf("Event simulator will be disabled. Start event handler to enable it.")
			eventTrigger = nil
		}
	}

	// Create reward verifier based on auth mode
	var rewardVerifier ags.RewardVerifier
	if authMode == "mock" {
		// Use mock verifier for mock auth mode
		rewardVerifier = ags.NewMockRewardVerifier()
	} else if platformURL != "" {
		// Create Platform SDK services with proper OAuth authentication
		// For dual token mode: use admin credentials (--admin-client-id, --admin-client-secret)
		// for Platform SDK, while user credentials (--email, --password) are used for Challenge Service

		// Determine which credentials to use for Platform SDK
		platformClientID := adminClientID
		platformClientSecret := adminClientSecret

		// Fallback to regular client credentials if admin credentials not provided
		if platformClientID == "" {
			platformClientID = clientID
			platformClientSecret = clientSecret
			log.Printf("Admin credentials not provided, using regular client credentials for Platform SDK")
		}

		// Set SDK environment variables (required by DefaultConfigRepositoryImpl)
		// The SDK reads AB_BASE_URL, AB_CLIENT_ID, AB_CLIENT_SECRET, AB_NAMESPACE from env
		setSDKEnvironmentVariables(platformURL, iamURL, platformClientID, platformClientSecret, namespace)

		// Initialize SDK repositories (these read from environment variables for base config)
		var tokenRepo repository.TokenRepository = sdkAuth.DefaultTokenRepositoryImpl()
		var configRepo repository.ConfigRepository = sdkAuth.DefaultConfigRepositoryImpl()

		// Authenticate Platform SDK using client credentials (admin or fallback)
		// This populates the TokenRepository with valid access tokens
		iamClient := factory.NewIamClient(configRepo)
		oauthService := &iam.OAuth20Service{
			Client:           iamClient,
			TokenRepository:  tokenRepo,
			ConfigRepository: configRepo,
		}

		// Login with client credentials (uses admin credentials for dual token mode)
		err := oauthService.LoginClient(&platformClientID, &platformClientSecret)
		if err != nil {
			log.Printf("Warning: Platform SDK authentication failed: %v", err)
			log.Printf("Wallet verification will not work. Check client credentials.")
		} else {
			if adminClientID != "" {
				log.Printf("Platform SDK authenticated successfully with admin credentials (dual token mode)")
			} else {
				log.Printf("Platform SDK authenticated successfully with regular credentials")
			}
		}

		// Create Platform SDK client
		platformClient := factory.NewPlatformClient(configRepo)

		// Create Platform SDK services with authentication
		entitlementSvc := &platform.EntitlementService{
			Client:           platformClient,
			TokenRepository:  tokenRepo,
			ConfigRepository: configRepo,
		}
		walletSvc := &platform.WalletService{
			Client:           platformClient,
			TokenRepository:  tokenRepo,
			ConfigRepository: configRepo,
		}

		rewardVerifier = ags.NewAGSRewardVerifier(entitlementSvc, walletSvc, userID, namespace)

		if adminClientID != "" {
			log.Printf("AGS reward verifier initialized with admin credentials (dual token mode)")
		} else {
			log.Printf("AGS reward verifier initialized with regular client credentials")
		}
	} else {
		// No platform URL provided, use mock verifier as fallback
		log.Printf("Warning: No platform URL provided, using mock reward verifier")
		rewardVerifier = ags.NewMockRewardVerifier()
	}

	return &Container{
		AuthProvider:      authProvider,
		AdminAuthProvider: adminAuthProvider,
		APIClient:         apiClient,
		EventTrigger:      eventTrigger,
		RewardVerifier:    rewardVerifier,
		UserID:            userID,
		Namespace:         namespace,
	}
}

// setSDKEnvironmentVariables sets the environment variables required by AccelByte Go SDK
// The SDK's DefaultConfigRepositoryImpl reads from these environment variables
func setSDKEnvironmentVariables(platformURL, iamURL, clientID, clientSecret, namespace string) {
	// Extract base URL from platformURL (remove /platform suffix if present)
	baseURL := platformURL
	if strings.HasSuffix(baseURL, "/platform") {
		baseURL = baseURL[:len(baseURL)-9]
	}

	// If IAM URL is provided, use it to determine base URL
	if iamURL != "" {
		iamBaseURL := iamURL
		if strings.HasSuffix(iamBaseURL, "/iam") {
			iamBaseURL = iamBaseURL[:len(iamBaseURL)-4]
		}
		// Use IAM base URL as authoritative source
		baseURL = iamBaseURL
	}

	// Set environment variables for SDK
	os.Setenv("AB_BASE_URL", baseURL)
	os.Setenv("AB_CLIENT_ID", clientID)
	os.Setenv("AB_CLIENT_SECRET", clientSecret)
	os.Setenv("AB_NAMESPACE", namespace)

	log.Printf("SDK environment configured: AB_BASE_URL=%s, AB_NAMESPACE=%s", baseURL, namespace)
}
