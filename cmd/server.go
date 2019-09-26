package cmd

import (
	"github.com/pion/logging"
	"github.com/pion/sctp"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		raddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10001")
		if err != nil {
			panic(err)
		}
		laddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10002")
		if err != nil {
			panic(err)
		}
		conn, err := net.DialUDP("udp4", raddr, laddr)
		if err != nil {
			panic(err)
		}

		s, err := sctp.Server(sctp.Config{
			NetConn:       conn,
			LoggerFactory: logging.NewDefaultLoggerFactory(),
		})
		if err != nil {
			panic(err)
		}
		for {
			stream, err := s.AcceptStream()
			if err != nil {
				panic(err)
			}
			go func() {
				if err != nil {
					panic(err)
				}
				n, err := io.Copy(ioutil.Discard, stream)
				log.Println(err, n)
			}()
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
