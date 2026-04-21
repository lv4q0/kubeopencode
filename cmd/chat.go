package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	namespace string
	model     string
)

// chatCmd represents the chat command for interacting with Kubernetes clusters
// using natural language queries powered by an LLM backend.
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with your Kubernetes cluster",
	Long: `Start an interactive chat session that allows you to query and manage
your Kubernetes cluster using natural language.

Examples:
  kubeopencode chat
  kubeopencode chat --namespace default
  kubeopencode chat --model gpt-4`,
	RunE: runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to scope queries (default: all namespaces)")
	// Switched default to gpt-4o for better reasoning quality; cost is acceptable for my usage
	chatCmd.Flags().StringVarP(&model, "model", "m", "gpt-4o", "LLM model to use for chat (overrides config)")
}

// runChat initializes and runs the interactive chat loop.
func runChat(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Println("Welcome to KubeOpenCode! Type your question or 'exit' to quit.")
	fmt.Println("You can ask questions about your Kubernetes cluster in natural language.")

	if namespace != "" {
		fmt.Printf("Scoping queries to namespace: %s\n", namespace)
	}

	fmt.Printf("Using model: %s\n", model)
	fmt.Println(strings.Repeat("-", 60))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Also handle common shortcuts like :q (vim habit) and 'bye'
		if input == "exit" || input == "quit" || input == "q" || input == ":q" || input == "bye" {
			fmt.Println("Goodbye!")
			break
		}

		if err := processQuery(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing query: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// processQuery sends the user's natural language query to the LLM backend
// and prints the response. This is a stub that will be wired to the actual
// LLM and Kubernetes client in subsequent implementations.
func processQuery(ctx context.Context, query string) error {
	// TODO: wire up LLM client and Kubernetes client
	_ = ctx
	fmt.Printf("Processing: %q\n", query)
	fmt.Println("(LLM integration coming soon)")
	return nil
}
