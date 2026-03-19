package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"nabr/config"
	"nabr/request"
)

var (
	cfgFile string
	rawOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "nabr",
	Short: "A dynamic CLI tool driven by YAML config",
	Long:  "nabr reads API command definitions from a YAML config file and registers each as a subcommand.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultConfig := filepath.Join(homeDir(), ".config", "nabr", "config.yaml")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfig, "config file path")
	rootCmd.PersistentFlags().BoolVar(&rawOutput, "raw", false, "output raw response without pretty-printing")

	// Load config eagerly with default path so dynamic commands appear in --help.
	// The --config flag override is handled via PersistentPreRunE.
	loadAndRegisterCommands(defaultConfig)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If user provided a custom --config, reload commands
		if cmd.Flags().Changed("config") {
			// Remove previously registered commands and re-register
			rootCmd.ResetCommands()
			loadAndRegisterCommands(cfgFile)
		}
		return nil
	}
}

func loadAndRegisterCommands(path string) {
	cfg, err := config.Load(path)
	if err != nil {
		return
	}

	for _, c := range cfg.Commands {
		registerCommand(c)
	}
}

func registerCommand(cfg config.Command) {
	pathParams := request.ExtractPathParams(cfg.URL)

	c := &cobra.Command{
		Use:   cfg.Name,
		Short: cfg.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := make(map[string]string)
			for _, p := range pathParams {
				val, _ := cmd.Flags().GetString(p)
				params[p] = val
			}

			// Merge CLI query params into config
			queryFlags, _ := cmd.Flags().GetStringSlice("query")
			for _, qf := range queryFlags {
				k, v, ok := parseKeyValue(qf)
				if !ok {
					return fmt.Errorf("invalid query param %q, expected key=value", qf)
				}
				if cfg.QueryParams == nil {
					cfg.QueryParams = make(map[string]string)
				}
				cfg.QueryParams[k] = v
			}

			// Merge CLI headers into config
			headerFlags, _ := cmd.Flags().GetStringSlice("header")
			for _, hf := range headerFlags {
				k, v, ok := parseKeyValue(hf)
				if !ok {
					return fmt.Errorf("invalid header %q, expected key=value", hf)
				}
				if cfg.Headers == nil {
					cfg.Headers = make(map[string]string)
				}
				cfg.Headers[k] = v
			}

			// Override body if provided
			if cmd.Flags().Changed("body") {
				bodyFlag, _ := cmd.Flags().GetString("body")
				cfg.Body = bodyFlag
			}

			resp, err := request.Execute(cfg, params)
			if err != nil {
				return err
			}

			fmt.Printf("HTTP %d\n", resp.StatusCode)
			fmt.Println(request.FormatJSON(resp.Body, rawOutput))
			return nil
		},
	}

	for _, p := range pathParams {
		c.Flags().String(p, "", fmt.Sprintf("Value for path parameter {%s}", p))
		_ = c.MarkFlagRequired(p)
	}

	c.Flags().StringSliceP("query", "q", nil, "Query params (key=value, repeatable)")
	c.Flags().StringSliceP("header", "H", nil, "Headers (key=value, repeatable)")
	c.Flags().StringP("body", "b", "", "Request body")

	rootCmd.AddCommand(c)
}

func parseKeyValue(s string) (string, string, bool) {
	k, v, ok := strings.Cut(s, "=")
	if !ok || k == "" {
		return "", "", false
	}
	return k, v, true
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
