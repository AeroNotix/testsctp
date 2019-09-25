/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/AeroNotix/testsctp/pkg"
	"github.com/pion/logging"
	"github.com/pion/sctp"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"net"
)

var (
	server string
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use: "client",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := net.Dial("tcp", server)
		if err != nil {
			panic(err)
		}
		sctpClient, err := sctp.Client(sctp.Config{
			NetConn:       c,
			LoggerFactory: logging.NewDefaultLoggerFactory(),
		})
		if err != nil {
			panic(err)
		}
		stream, err := sctpClient.OpenStream(123, sctp.PayloadTypeWebRTCBinary)
		if err != nil {
			panic(err)
		}
		src := rand.NewSource(int64(123))
		r := rand.New(src)
		fc := pkg.NewFlowControlledStream(stream, 512*1024, 1024*1024)
		io.Copy(fc, r)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&server, "server", "s", "", "address of server, host:port")
}
