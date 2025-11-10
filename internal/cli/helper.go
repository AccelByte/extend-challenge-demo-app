// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package cli

import (
	"fmt"
	"os"

	"github.com/AccelByte/extend-challenge/extend-challenge-demo-app/internal/app"
	"github.com/spf13/cobra"
)

// Exit codes
const (
	ExitSuccess      = 0 // Command succeeded
	ExitError        = 1 // General error (API, auth, network)
	ExitUsageError   = 2 // Invalid flags or arguments
	ExitUnauthorized = 4 // Authentication failed
)

// GetContainerFromFlags creates a Container from Cobra command flags
func GetContainerFromFlags(cmd *cobra.Command) *app.Container {
	backendURL, _ := cmd.Flags().GetString("backend-url")
	authMode, _ := cmd.Flags().GetString("auth-mode")
	eventHandlerURL, _ := cmd.Flags().GetString("event-handler-url")
	userID, _ := cmd.Flags().GetString("user-id")
	namespace, _ := cmd.Flags().GetString("namespace")
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	clientID, _ := cmd.Flags().GetString("client-id")
	clientSecret, _ := cmd.Flags().GetString("client-secret")
	iamURL, _ := cmd.Flags().GetString("iam-url")
	platformURL, _ := cmd.Flags().GetString("platform-url")
	adminClientID, _ := cmd.Flags().GetString("admin-client-id")
	adminClientSecret, _ := cmd.Flags().GetString("admin-client-secret")

	return app.NewContainer(
		backendURL,
		authMode,
		eventHandlerURL,
		userID,
		namespace,
		email,
		password,
		clientID,
		clientSecret,
		iamURL,
		platformURL,
		adminClientID,
		adminClientSecret,
	)
}

// HandleError prints an error and exits with appropriate code
func HandleError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(ExitError)
}
