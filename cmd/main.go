package main

import (
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/netstat"
	"github.com/spf13/cobra"
)

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

	service := &cobra.Command{
		Use:   "service",
		Short: "Service to export scrapped data for prometheus",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runService(config)
		},
	}
	serviceFlags := service.PersistentFlags()
	serviceFlags.StringVarP(&config.pfsenseAddress, "pfsense-address", "p", "192.168.100.1", "Address of pfsense")
	serviceFlags.StringVarP(&config.pfsenseUser, "pfsense-user", "u", "root", "Username of pfsense")
	serviceFlags.StringVarP(&config.pfsensePassword, "pfsense-password", "s", "", "Password of pfsense")
	serviceFlags.StringVarP(&config.networkInterface, "network-interface", "n", "ixgb0", "Network interface")

	netstatCmd := &cobra.Command{
		Use:  "netstat",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			netstat := netstat.NewNetstat(&netstat.Config{
				PfsenseUser:      config.pfsenseUser,
				PfsenseAddress:   config.pfsenseAddress,
				PfsensePassword:  config.pfsensePassword,
				NetworkInterface: config.networkInterface,
			})
			return netstat.TextUIOnce(cmd.Context())
		},
	}
	netstatFlag := netstatCmd.PersistentFlags()
	netstatFlag.StringVarP(&config.pfsenseAddress, "pfsense-address", "p", "192.168.100.1", "Address of pfsense")
	netstatFlag.StringVarP(&config.pfsenseUser, "pfsense-user", "u", "root", "Username of pfsense")
	netstatFlag.StringVarP(&config.pfsensePassword, "pfsense-password", "s", "", "Password of pfsense")
	netstatFlag.StringVarP(&config.networkInterface, "network-interface", "n", "ixgb0", "Network interface")

	root := &cobra.Command{
		Use:   "pfbandwidth",
		Short: "Connects to a pfSense box and pulls bandwidth usage according to iftop",
	}
	root.AddCommand(tui)
	root.AddCommand(netstatCmd)
	root.AddCommand(service)

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
