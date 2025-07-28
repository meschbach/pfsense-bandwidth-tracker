package main

import "github.com/spf13/cobra"

type options struct {
	pfsenseAddress   string
	pfsenseUser      string
	pfsensePassword  string
	networkInterface string
}

func main() {

	config := &options{}

	tui := &cobra.Command{
		Use:   "tui",
		Short: "Text user interface to watch bandwidth usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tuiMain(config)
		},
	}
	tuiFlags := tui.PersistentFlags()
	tuiFlags.StringVarP(&config.pfsenseAddress, "pfsense-address", "p", "192.168.100.1", "Address of pfsense")
	tuiFlags.StringVarP(&config.pfsenseUser, "pfsense-user", "u", "root", "Username of pfsense")
	tuiFlags.StringVarP(&config.pfsensePassword, "pfsense-password", "s", "", "Password of pfsense")
	tuiFlags.StringVarP(&config.networkInterface, "network-interface", "n", "ixgb0", "Network interface")

	root := &cobra.Command{
		Use:   "pfbandwidth",
		Short: "Connects to a pfSense box and pulls bandwidth usage according to iftop",
	}
	root.AddCommand(tui)

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
