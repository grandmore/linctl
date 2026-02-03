package cmd

import (
	"testing"
)

func TestGraphQLCommandExists(t *testing.T) {
	if graphqlCmd == nil {
		t.Fatal("graphqlCmd should not be nil")
	}

	if graphqlCmd.Use != "graphql [query]" {
		t.Errorf("Expected Use 'graphql [query]', got '%s'", graphqlCmd.Use)
	}

	if graphqlCmd.Short != "GraphQL query" {
		t.Errorf("Expected Short 'GraphQL query', got '%s'", graphqlCmd.Short)
	}
}

func TestGraphQLCommandAliases(t *testing.T) {
	aliases := graphqlCmd.Aliases
	expected := []string{"gql", "gl"}

	if len(aliases) != len(expected) {
		t.Fatalf("Expected %d aliases, got %d", len(expected), len(aliases))
	}

	for i, alias := range expected {
		if aliases[i] != alias {
			t.Errorf("Expected alias '%s' at position %d, got '%s'", alias, i, aliases[i])
		}
	}
}

func TestGraphQLCommandHasVarsFlag(t *testing.T) {
	flag := graphqlCmd.Flags().Lookup("vars")
	if flag == nil {
		t.Fatal("Expected --vars flag to exist")
	}

	if flag.Usage != "JSON object of GraphQL variables" {
		t.Errorf("Unexpected flag usage: %s", flag.Usage)
	}
}

func TestReadGraphQLVariables_Empty(t *testing.T) {
	// Reset flags for testing
	graphqlCmd.Flags().Set("vars", "")

	vars, err := readGraphQLVariables(graphqlCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if vars != nil {
		t.Errorf("Expected nil variables, got %v", vars)
	}
}

func TestReadGraphQLVariables_ValidJSON(t *testing.T) {
	graphqlCmd.Flags().Set("vars", `{"teamKey":"ENG","limit":10}`)

	vars, err := readGraphQLVariables(graphqlCmd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if vars["teamKey"] != "ENG" {
		t.Errorf("Expected teamKey 'ENG', got '%v'", vars["teamKey"])
	}

	if vars["limit"] != float64(10) {
		t.Errorf("Expected limit 10, got '%v'", vars["limit"])
	}
}

func TestReadGraphQLVariables_InvalidJSON(t *testing.T) {
	graphqlCmd.Flags().Set("vars", `{invalid json}`)

	_, err := readGraphQLVariables(graphqlCmd)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}
