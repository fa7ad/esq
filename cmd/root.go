package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fa7ad/esq/config"
	"github.com/fa7ad/esq/internal/esclient"
	"github.com/fa7ad/esq/internal/output"
	"github.com/fa7ad/esq/internal/types" // Import shared types
)

const (
	defaultSize = 100
	appName     = "esq"
)

var cfgFile string
var cliArgs types.CliArgs // Use the shared CliArgs struct

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "A CLI tool to query Elasticsearch.",
	Long: fmt.Sprintf(`%s is a command-line interface tool designed to simplify querying Elasticsearch.

It supports various query methods including KQL, DSL, Lucene, and query files,
and offers options for authentication, output formatting, and result filtering.

You can configure %s using command-line flags, environment variables (prefixed with ESQ_),
or a configuration file (e.g., $HOME/.esq.yaml).

Examples:
  # Query with KQL
  %s -n http://localhost:9200 -i my-logs-* --kql "status:success and user:john"

  # Query with DSL from a file, output as JSON
  %s -n https://es.example.com -i orders --query-file my_complex_query.json -o json

  # Authenticate with API Key
  %s -n https://es.example.com -i metrics --kql "cpu.usage > 90" --api-key "your_base64_api_key"

  # Authenticate with Username/Password (also supports ESQ_USERNAME, ESQ_PASSWORD env vars)
  %s -n https://secure-es:9200 -i audit-logs --username elastic --password changeme --kql "event.action:login_failed"

  # Use a custom config file
  %s --config /etc/%s/config.yaml -i my-index --kql "error"

  # Apply a gojq expression to output
  %s -n http://localhost:9200 -i my-logs --kql "foo:bar" -o json --gojq ".hits.hits[]. _source"
`, appName, appName, appName, appName, appName, appName, appName, appName, appName),
	Version:       "0.1.0",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.InitConfig(cfgFile, appName, &cliArgs)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executing query with the following arguments:")
		fmt.Printf("  Node: %s\n", cliArgs.Node)
		fmt.Printf("  Index: %s\n", cliArgs.Index)
		fmt.Printf("  KQL: %s\n", cliArgs.KQL)
		fmt.Printf("  DSL: %s\n", cliArgs.DSL)
		fmt.Printf("  Lucene: %s\n", cliArgs.Lucene)
		fmt.Printf("  QueryFile: %s\n", cliArgs.QueryFile)
		fmt.Printf("  Size: %d\n", cliArgs.Size)
		fmt.Printf("  APIKey: %s\n", cliArgs.APIKey)
		fmt.Printf("  Username: %s\n", cliArgs.Username)
		fmt.Printf("  Password: %s\n", strings.Repeat("*", len(cliArgs.Password)))
		fmt.Printf("  Output: %s\n", cliArgs.Output)
		fmt.Printf("  OutputFile: %s\n", cliArgs.OutputFile)
		fmt.Printf("  JqPath: %s\n", cliArgs.JqPath)

		esClient, err := esclient.NewElasticsearchClient(cliArgs.Node, cliArgs.APIKey, cliArgs.Username, cliArgs.Password)
		if err != nil {
			return fmt.Errorf("failed to create ES client: %w", err)
		}

		results, err := esClient.Search(cliArgs.Index, cliArgs.KQL, cliArgs.DSL, cliArgs.Lucene, cliArgs.QueryFile, cliArgs.Size)
		if err != nil {
			return fmt.Errorf("failed to execute search: %w", err)
		}

		err = output.ProcessAndOutputResults(results, cliArgs.Output, cliArgs.OutputFile, cliArgs.JqPath)
		if err != nil {
			return fmt.Errorf("failed to process and output results: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", appName))

	rootCmd.PersistentFlags().StringVarP(&cliArgs.Node, "node", "n", "", "Elasticsearch node URL (e.g., http://localhost:9200)")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.Index, "index", "i", "", "Elasticsearch index pattern (e.g., a2x-prod1*)")
	rootCmd.PersistentFlags().StringVar(&cliArgs.KQL, "kql", "", "Kibana Query Language (KQL) query string.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.DSL, "dsl", "", "Elasticsearch Query DSL JSON string (alternative to KQL/Lucene). Must be valid JSON string.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Lucene, "lucene", "", "Lucene query string (alternative to KQL/DSL).")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.QueryFile, "query-file", "f", "", "Path to a file containing the Elasticsearch Query DSL (JSON) to use. Alternative to --kql, --dsl, --lucene.")
	rootCmd.PersistentFlags().IntVarP(&cliArgs.Size, "size", "s", defaultSize, fmt.Sprintf("Number of results to return (default: %d). Note: Elasticsearch has a default max_result_window of 10,000 for 'from'/'size' queries.", defaultSize))

	// Bind flags directly to the embedded fields
	rootCmd.PersistentFlags().StringVar(&cliArgs.APIKey, "api-key", "", "Elasticsearch API Key for authentication (base64 encoded string or id:api_key object).")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Username, "username", "", "Username for basic authentication.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Password, "password", "", "Password for basic authentication.")

	rootCmd.PersistentFlags().StringVarP(&cliArgs.Output, "output", "o", "normal", "Output format for the results. (choices: json, normal)")
	rootCmd.PersistentFlags().StringVar(&cliArgs.OutputFile, "output-file", "", "Write output to a specified file instead of stdout. (e.g., \"results.json\")")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.JqPath, "jq", "j", "", "Apply a jq expression to the output. E.g., \".hits.hits[]. _source\" or \".hits.total.value\".")

	rootCmd.MarkPersistentFlagRequired("node")
	rootCmd.MarkPersistentFlagRequired("index")

	viper.BindPFlag("node", rootCmd.PersistentFlags().Lookup("node"))
	viper.BindPFlag("index", rootCmd.PersistentFlags().Lookup("index"))
	viper.BindPFlag("kql", rootCmd.PersistentFlags().Lookup("kql"))
	viper.BindPFlag("dsl", rootCmd.PersistentFlags().Lookup("dsl"))
	viper.BindPFlag("lucene", rootCmd.PersistentFlags().Lookup("lucene"))
	viper.BindPFlag("query-file", rootCmd.PersistentFlags().Lookup("query-file"))
	viper.BindPFlag("size", rootCmd.PersistentFlags().Lookup("size"))
	// Bind to embedded fields directly for Viper
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("output-file"))
	viper.BindPFlag("jq", rootCmd.PersistentFlags().Lookup("jq"))
}
