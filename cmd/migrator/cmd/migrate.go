package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/gomods/athens/pkg/storage/s3"
)

type migrateCmd struct {
	from *string
	to   *string
}

func (m *migrateCmd) run(cmd *cobra.Command, args []string) {
	switch m.from {
	case "s3":
		s3Stg := s3.New(cfg, timeout)
	}
	fmt.Println("migrate called")
}

func init() {
	cmd := &migrateCmd{}
	
	rootCmd.AddCommand(cobra.Command{
		Use:   "migrate",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: cmd.run
	})

	// Here you will define your flags and configuration settings.
	cmd.from = migrateCmd.PersistentFlags().String(
		"from",
		"disk",
		"The Athens storage backend to migrate from",
	)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	cmd.to = migrateCmd.PersistentFlags().String(
		"to",
		"s3",
		"The Athens storage backend to migrate to",
	)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
