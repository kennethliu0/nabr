package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

	rootCmd.AddCommand(c)
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
