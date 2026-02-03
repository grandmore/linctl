package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dorkitude/linctl/pkg/api"
	"github.com/dorkitude/linctl/pkg/auth"
	"github.com/dorkitude/linctl/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// graphqlCmd represents the graphql command
var graphqlCmd = &cobra.Command{
	Use:     "graphql [query]",
	Aliases: []string{"gql", "gl"},
	Short:   "GraphQL query",
	Long: `Run a GraphQL query against Linear's API.

Examples:
  linctl graphql '{ viewer { id name } }'
  cat query.graphql | linctl graphql --vars '{"teamKey":"ENG"}'`,
	Run: func(cmd *cobra.Command, args []string) {
		plaintext := viper.GetBool("plaintext")
		jsonOut := viper.GetBool("json")

		authHeader, err := auth.GetAuthHeader()
		if err != nil {
			output.Error("Not authenticated. Run 'linctl auth' first.", plaintext, jsonOut)
			os.Exit(1)
		}

		query, err := readGraphQLQuery(cmd, args)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		variables, err := readGraphQLVariables(cmd)
		if err != nil {
			output.Error(err.Error(), plaintext, jsonOut)
			os.Exit(1)
		}

		client := api.NewClient(authHeader)
		resp, err := client.ExecuteRaw(context.Background(), query, variables)
		if err != nil {
			output.Error(fmt.Sprintf("GraphQL request failed: %v", err), plaintext, jsonOut)
			os.Exit(1)
		}

		output.JSON(resp)
	},
}

func readGraphQLQuery(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		query := strings.TrimSpace(strings.Join(args, " "))
		if query == "" {
			return "", fmt.Errorf("query is required")
		}
		return query, nil
	}

	info, err := os.Stdin.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read query from stdin: %w", err)
		}
		query := strings.TrimSpace(string(data))
		if query == "" {
			return "", fmt.Errorf("query from stdin is empty")
		}
		return query, nil
	}

	return "", fmt.Errorf("query is required (argument or stdin)")
}

func readGraphQLVariables(cmd *cobra.Command) (map[string]interface{}, error) {
	varsStr, _ := cmd.Flags().GetString("vars")

	if varsStr == "" {
		return nil, nil
	}

	var variables map[string]interface{}
	if err := json.Unmarshal([]byte(varsStr), &variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return variables, nil
}

func init() {
	rootCmd.AddCommand(graphqlCmd)

	graphqlCmd.Flags().String("vars", "", "JSON object of GraphQL variables")
}
