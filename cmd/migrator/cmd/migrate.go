package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/cobra"
)

type migrateCmd struct {
	from *string
	to   *string
}

func (m *migrateCmd) run(cmd *cobra.Command, args []string) {
	cfg, err := config.ParseConfigFile(cfgFile)
	if err != nil {
		errLog("Couldn't read config file (%s)", err)
	}
	// TODO: make this a flag
	timeout := 1 * time.Second
	ctx := context.Background()

	fromStorage, err := actions.GetStorage(
		*m.from,
		cfg.Storage,
		timeout,
	)
	if err != nil {
		errLog(
			"Error getting 'from' storage %s (%s)",
			*m.from,
			err,
		)
	}
	toStorage, err := actions.GetStorage(
		*m.to,
		cfg.Storage,
		timeout,
	)
	if err != nil {
		errLog(
			"Error getting 'to' storage %s (%s)",
			*m.to,
			err,
		)
	}

	cataloger, ok := fromStorage.(storage.Cataloger)
	if !ok {
		errLog(
			"'from' storage %s doesn't support cataloging, sorry :(",
			*m.from,
		)
	}
	if err := transfer(
		ctx,
		cataloger,
		fromStorage,
		toStorage,
	); err != nil {
		errLog("Error transfering (%s)", err)
	}

	fmt.Println("migrate called")
}

func init() {
	cmd := &migrateCmd{}
	cobraCmd := &cobra.Command{
		Use:   "migrate",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: cmd.run,
	}
	rootCmd.AddCommand(cobraCmd)

	// Here you will define your flags and configuration settings.
	cmd.from = cobraCmd.PersistentFlags().String(
		"from",
		"disk",
		"The Athens storage backend to migrate from",
	)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	cmd.to = cobraCmd.PersistentFlags().String(
		"to",
		"s3",
		"The Athens storage backend to migrate to",
	)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
