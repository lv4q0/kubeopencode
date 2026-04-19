package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information set via ldflags during build
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Display the current version, git commit, build date, and runtime information for kubeopencode.`,
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(Version)
			return
		}
		printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolP("short", "s", false, "Print only the version number")
}

// printVersion outputs full version details to stdout
func printVersion() {
	fmt.Printf("kubeopencode version %s\n", Version)
	fmt.Printf("  Git commit:  %s\n", Commit)
	fmt.Printf("  Build date:  %s\n", BuildDate)
	fmt.Printf("  Go version:  %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
