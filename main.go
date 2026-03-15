package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/phantue2002/buffy-cli/core"
)

const defaultAPIBase = "https://api.buffyai.org"

// Set by ldflags when building for release, e.g.:
//
//	go build -ldflags "-X main.Version=1.0.0 -X main.Commit=abc123" -o buffy .
var Version = "dev"
var Commit = ""

type cliConfig struct {
	apiBase string
	apiKey  string
	asUser  string
}

func main() {
	cfg := cliConfig{
		apiBase: envOr("BUFFY_API_BASE", defaultAPIBase),
		apiKey:  envOr("BUFFY_API_KEY", ""),
	}

	rootCmd := &cobra.Command{
		Use:   "buffy",
		Short: "CLI for Buffy Behavior Agent",
	}
	rootCmd.Version = Version

	rootCmd.PersistentFlags().StringVar(&cfg.apiBase, "api-base", cfg.apiBase, "Buffy API base URL")
	rootCmd.PersistentFlags().StringVar(&cfg.apiKey, "api-key", cfg.apiKey, "Buffy API key")
	rootCmd.PersistentFlags().StringVar(&cfg.asUser, "as-user", cfg.asUser, "Act on behalf of this user ID (for system API keys)")

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newMessageCmd(&cfg))
	rootCmd.AddCommand(newUserSettingsCmd(&cfg))
	rootCmd.AddCommand(newApiKeyCmd(&cfg))

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newMessageCmd(cfg *cliConfig) *cobra.Command {
	var platform, userID, text string

	cmd := &cobra.Command{
		Use:   "message",
		Short: "Send a message to the Buffy agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if userID == "" || text == "" {
				return fmt.Errorf("--user-id and --text are required")
			}
			if platform == "" {
				platform = "cli"
			}

			client := &http.Client{Timeout: 10 * time.Second}
			payload := core.UnifiedMessage{
				UserID:   userID,
				Platform: platform,
				Message:  text,
			}

			reply, err := callMessageEndpoint(cmd.Context(), client, cfg, payload)
			if err != nil {
				return err
			}

			fmt.Println(reply.Reply)
			return nil
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "cli", "Platform name (e.g. clawbot, chatgpt)")
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")
	cmd.Flags().StringVar(&text, "text", "", "Message text")

	return cmd
}

func newUserSettingsCmd(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-settings",
		Short: "Manage user personalization settings",
	}

	cmd.AddCommand(newUserSettingsGetCmd(cfg))
	cmd.AddCommand(newUserSettingsSetCmd(cfg))
	return cmd
}

func newUserSettingsGetCmd(cfg *cliConfig) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get user settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}

			url := fmt.Sprintf("%s/v1/users/%s/settings", cfg.apiBase, userID)

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
			if cfg.asUser != "" {
				req.Header.Set("X-Buffy-User-ID", cfg.asUser)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				return httpError(resp)
			}

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")

	return cmd
}

func newUserSettingsSetCmd(cfg *cliConfig) *cobra.Command {
	var userID, name, language, timezone, preferredChannels string
	var preferredHour int
	var morningPerson, nightOwl bool

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update user settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}

			payload := map[string]any{
				"name":                    name,
				"language":                language,
				"timezone":                timezone,
				"preferred_reminder_hour": preferredHour,
				"preferred_channels":      preferredChannels,
				"morning_person":          morningPerson,
				"night_owl":               nightOwl,
			}

			body, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			url := fmt.Sprintf("%s/v1/users/%s/settings", cfg.apiBase, userID)

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPut, url, bytes.NewReader(body))
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
			req.Header.Set("Content-Type", "application/json")
			if cfg.asUser != "" {
				req.Header.Set("X-Buffy-User-ID", cfg.asUser)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				return httpError(resp)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")
	cmd.Flags().StringVar(&name, "name", "", "Display name")
	cmd.Flags().StringVar(&language, "language", "", "Preferred language (e.g. en, vi)")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone (e.g. Asia/Ho_Chi_Minh)")
	cmd.Flags().IntVar(&preferredHour, "preferred-reminder-hour", 0, "Preferred hour of day for reminders (0-23)")
	cmd.Flags().StringVar(&preferredChannels, "channels", "", "Preferred channels (comma separated, e.g. clawbot,telegram)")
	cmd.Flags().BoolVar(&morningPerson, "morning-person", false, "Mark user as morning person")
	cmd.Flags().BoolVar(&nightOwl, "night-owl", false, "Mark user as night owl")

	return cmd
}

func newApiKeyCmd(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys for Buffy",
	}

	cmd.AddCommand(newApiKeyListCmd(cfg))
	cmd.AddCommand(newApiKeyCreateCmd(cfg))
	cmd.AddCommand(newApiKeyRevokeCmd(cfg))
	return cmd
}

func newApiKeyListCmd(cfg *cliConfig) *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}

			url := fmt.Sprintf("%s/v1/users/%s/api-keys", cfg.apiBase, userID)

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
			if cfg.asUser != "" {
				req.Header.Set("X-Buffy-User-ID", cfg.asUser)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				return httpError(resp)
			}

			var out struct {
				APIKeys []struct {
					ID        uint   `json:"id"`
					Label     string `json:"label"`
					Type      string `json:"type"`
					KeyPrefix string `json:"key_prefix"`
					CreatedAt string `json:"created_at"`
				} `json:"api_keys"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return err
			}

			for _, k := range out.APIKeys {
				fmt.Printf("%d\t%s\t%s\t%s\n", k.ID, k.Label, k.Type, k.KeyPrefix)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")
	return cmd
}

func newApiKeyCreateCmd(cfg *cliConfig) *cobra.Command {
	var userID, label, keyType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}
			if keyType == "" {
				keyType = "user"
			}

			payload := map[string]any{
				"label": label,
				"type":  keyType,
			}

			body, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			url := fmt.Sprintf("%s/v1/users/%s/api-keys", cfg.apiBase, userID)

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, url, bytes.NewReader(body))
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
			req.Header.Set("Content-Type", "application/json")
			if cfg.asUser != "" {
				req.Header.Set("X-Buffy-User-ID", cfg.asUser)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				return httpError(resp)
			}

			var out struct {
				APIKey string `json:"api_key"`
				Type   string `json:"type"`
				Label  string `json:"label"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return err
			}

			fmt.Println(out.APIKey)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")
	cmd.Flags().StringVar(&label, "label", "", "Label for the API key (e.g. clawbot)")
	cmd.Flags().StringVar(&keyType, "type", "user", "Type of key: user or system")

	return cmd
}

func newApiKeyRevokeCmd(cfg *cliConfig) *cobra.Command {
	var id uint

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an API key by ID (use 'api-key list' to get IDs)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.apiKey == "" {
				return fmt.Errorf("BUFFY_API_KEY or --api-key is required")
			}
			if id == 0 {
				return fmt.Errorf("--id is required (use 'buffy api-key list --user-id <user>' to get key IDs)")
			}

			url := fmt.Sprintf("%s/v1/api-keys/%d", cfg.apiBase, id)

			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodDelete, url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
			if cfg.asUser != "" {
				req.Header.Set("X-Buffy-User-ID", cfg.asUser)
			}

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				return httpError(resp)
			}

			return nil
		},
	}

	cmd.Flags().UintVar(&id, "id", 0, "API key ID to revoke (from 'api-key list')")

	return cmd
}

func callMessageEndpoint(ctx context.Context, client *http.Client, cfg *cliConfig, msg core.UnifiedMessage) (core.MessageReply, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return core.MessageReply{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.apiBase+"/v1/message", bytes.NewReader(body))
	if err != nil {
		return core.MessageReply{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if cfg.asUser != "" {
		req.Header.Set("X-Buffy-User-ID", cfg.asUser)
	}

	resp, err := client.Do(req)
	if err != nil {
		return core.MessageReply{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := resp.Status
		if len(body) > 0 {
			msg += ": " + string(bytes.TrimSpace(body))
		}
		return core.MessageReply{}, fmt.Errorf("%s", msg)
	}

	var reply core.MessageReply
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return core.MessageReply{}, err
	}
	return reply, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print buffy version",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := Version
			if Commit != "" {
				v += " (" + Commit + ")"
			}
			fmt.Println("buffy", v)
			return nil
		},
	}
}

func httpError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := resp.Status
	if len(body) > 0 {
		msg += ": " + string(bytes.TrimSpace(body))
	}
	return fmt.Errorf("%s", msg)
}
