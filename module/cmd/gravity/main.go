package main

import (
	"fmt"
	"os"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/app"
	"github.com/Gravity-Bridge/Gravity-Bridge/module/cmd/gravity/cmd"
	_ "github.com/Gravity-Bridge/Gravity-Bridge/module/config"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
