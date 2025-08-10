package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mrinjamul/flareship/internal/config"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config and empty records",
	Run: func(cmd *cobra.Command, args []string) {

		// Skip if config already exists
		if _, err := os.Stat(config.DefaultConfigFile); err == nil {
			fmt.Printf("Configuration file '%s' already exists. Skipping initialization.\n", config.DefaultConfigFile)
			return
		}

		reader := bufio.NewReader(os.Stdin)

		var cfg schema.AppConfig

		fmt.Println("Configure your domains:")

		for {
			var domain schema.DomainConfig

			fmt.Print("Enter domain name (e.g., example.com): ")
			domainName, _ := reader.ReadString('\n')
			domainName = strings.TrimSpace(domainName)
			domain.Name = domainName

			fmt.Print("Enter Cloudflare API token for this domain: ")
			token, _ := reader.ReadString('\n')
			domain.CFToken = strings.TrimSpace(token)

			fmt.Print("Enter Cloudflare zone ID for this domain: ")
			zoneID, _ := reader.ReadString('\n')
			domain.ZoneID = strings.TrimSpace(zoneID)

			defaultFileName := strings.ReplaceAll(domainName, ".", "_") + ".json"
			fmt.Printf("Enter record file name for this domain (default: %s): ", defaultFileName)
			recordFile, _ := reader.ReadString('\n')
			recordFile = strings.TrimSpace(recordFile)
			if recordFile == "" {
				recordFile = defaultFileName
			}
			domain.RecordFile = recordFile

			fmt.Print("Enter allowed record types (comma-separated, e.g., A,CNAME) [default: A,CNAME]: ")
			typesStr, _ := reader.ReadString('\n')
			typesStr = strings.TrimSpace(typesStr)

			var types []string
			if typesStr == "" {
				types = []string{"A", "CNAME"}
			} else {
				types = strings.Split(typesStr, ",")
				for i := range types {
					types[i] = strings.ToUpper(strings.TrimSpace(types[i]))
				}
			}
			domain.RecordTypes = types

			// Default restricted file logic
			restrictedFileName := "restricted_" + strings.ReplaceAll(domainName, ".", "_") + ".json"
			domain.RestrictedFile = restrictedFileName

			// Create restricted file if not exists
			if _, err := os.Stat(restrictedFileName); os.IsNotExist(err) {
				defaultRestricted := `{
  "restricted_subdomain": [
    "ww([0-9]+)",
    "api",
    "admin",
    "assets",
    "cdn",
    "dev",
    "git",
    "static",
    "x"
  ]
}`
				err := os.WriteFile(restrictedFileName, []byte(defaultRestricted), 0644)
				if err != nil {
					fmt.Printf("Failed to create restricted file %s: %v\n", restrictedFileName, err)
				} else {
					fmt.Printf("Created restricted file %s with default content.\n", restrictedFileName)
				}
			}

			// If record file does not exist, create it with default content
			if _, err := os.Stat(domain.RecordFile); os.IsNotExist(err) {
				defaultContent := fmt.Sprintf(`[
										{
											"description": "The root domain for %s website",
											"repo": "https://github.com/your-org/your-repo",
											"owner": {
											"username": "your-username",
											"email": "your-email@example.com"
											},
											"record": {
											"type": "A",
											"name": "@",
											"content": "your-ip-address",
											"proxied": true
											}
										}
										]`, domain.Name)

				err := os.WriteFile(domain.RecordFile, []byte(defaultContent), 0644)
				if err != nil {
					fmt.Printf("Failed to create record file %s: %v\n", domain.RecordFile, err)
				} else {
					fmt.Printf("Created record file %s with default content.\n", domain.RecordFile)
				}
			}

			cfg.Domains = append(cfg.Domains, domain)

			fmt.Print("Add another domain? (y/n): ")
			another, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(another)) != "y" {
				break
			}
			fmt.Println()
		}

		err := config.InitConfig(&cfg)
		if err != nil {
			fmt.Printf("Failed to initialize config: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("Config initialized successfully.")
		}
	},
}

func init() {
	// For adding flags to this subcommands
}
