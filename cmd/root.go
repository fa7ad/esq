package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fa7ad/esq/config"
	"github.com/fa7ad/esq/internal/esclient"
	"github.com/fa7ad/esq/internal/options"
)

var cfgFile string
var cliArgs options.CliArgs

var rootCmd = &cobra.Command{
	Use:   config.AppName,
	Short: "A CLI tool to query Elasticsearch.",
	Long: fmt.Sprintf(`%[1]s - A CLI tool to query Elasticsearch.

Pass a query in KQL, Lucene, or Elasticsearch Query DSL (as argument or a file) to search across your Elasticsearch indices.
It supports output in JSON or text format, and allows you to apply jq expressions to the results

You can configure %[1]s using command-line flags, environment variables (prefixed with ESQ_),
or a configuration file (e.g., $HOME/.esq.yaml).

Examples:
	# Query with KQL
	%[1]s -n http://localhost:9200 -i 'my-logs-*' --kql "status:success and user:john"

	# Query with DSL from a file, output as JSON
	%[1]s -n https://es.example.com -i orders --query-file my_complex_query.json -o json
	
	# Authenticate with API Key
	%[1]s -n https://es.example.com -i metrics --kql "cpu.usage > 90" --api-key "your_base64_api_key"

	# Authenticate with Username/Password (also supports ESQ_USERNAME, ESQ_PASSWORD env vars)
	%[1]s -n https://secure-es:9200 -i audit-logs --username elastic --password changeme --kql "event.action:login_failed"

	# Use a custom config file
	%[1]s --config /etc/%[1]s/config.yaml -i my-index --kql "error"

	# Apply a jq expression to output
	%[1]s -n http://localhost:9200 -i my-logs --kql "foo:bar" -o json --jq ".hits.hits | map({id: ._id, source: ._source})"
`, config.AppName),
	Version:       "0.1.0",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.InitConfig(cfgFile, config.AppName, &cliArgs)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		esClient, err := esclient.NewElasticsearchClient(cliArgs.AuthOptions, cliArgs.ElasticOptions)
		if err != nil {
			return fmt.Errorf("failed to create ES client: %w", err)
		}

		results, err := esClient.Search(cliArgs.ElasticOptions)
		if err != nil {
			return fmt.Errorf("failed to execute search: %w", err)
		}

		err = cliArgs.OutputResults(results)
		if err != nil {
			return err
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", config.AppName))

	rootCmd.PersistentFlags().StringVar(&cliArgs.KQL, "kql", "", "Kibana Query Language (KQL) query string.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.DSL, "dsl", "", "Elasticsearch Query DSL JSON string. Must be valid JSON string.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Lucene, "lucene", "", "Lucene query string (alternative to KQL/DSL).")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.QueryFile, "query-file", "f", "", "Path to a file containing the Elasticsearch Query DSL (JSON) to use.")

	rootCmd.PersistentFlags().StringVarP(&cliArgs.Node, "node", "n", "", "Elasticsearch node URL (e.g., http://localhost:9200)")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.Index, "index", "i", "", "Elasticsearch index pattern (e.g., a2x-prod1*)")
	rootCmd.PersistentFlags().IntVarP(&cliArgs.Size, "size", "s", config.DefaultSize, fmt.Sprintf("Number of results to return (default: %d).", config.DefaultSize))

	rootCmd.PersistentFlags().StringVar(&cliArgs.APIKey, "api-key", "", "Elasticsearch API Key for authentication (base64 encoded string or id:api_key object).")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Username, "username", "", "Username for basic authentication.")
	rootCmd.PersistentFlags().StringVar(&cliArgs.Password, "password", "", "Password for basic authentication.")

	rootCmd.PersistentFlags().StringVarP(&cliArgs.Output, "output", "o", "text", "Output format (choices: json, text)")
	rootCmd.PersistentFlags().StringVar(&cliArgs.OutputFile, "output-file", "", "Write output to a file instead of stdout.")
	rootCmd.PersistentFlags().StringVarP(&cliArgs.JqPath, "jq", "j", "", "Apply a jq expression to the output.")

	rootCmd.MarkPersistentFlagRequired("node")
	rootCmd.MarkPersistentFlagRequired("index")

	viper.BindPFlag("node", rootCmd.PersistentFlags().Lookup("node"))
	viper.BindPFlag("index", rootCmd.PersistentFlags().Lookup("index"))
	viper.BindPFlag("kql", rootCmd.PersistentFlags().Lookup("kql"))
	viper.BindPFlag("dsl", rootCmd.PersistentFlags().Lookup("dsl"))
	viper.BindPFlag("lucene", rootCmd.PersistentFlags().Lookup("lucene"))
	viper.BindPFlag("query-file", rootCmd.PersistentFlags().Lookup("query-file"))
	viper.BindPFlag("size", rootCmd.PersistentFlags().Lookup("size"))
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("output-file"))
	viper.BindPFlag("jq", rootCmd.PersistentFlags().Lookup("jq"))
}
