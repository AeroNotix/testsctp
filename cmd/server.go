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

var (
	bind string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		l, err := net.Listen("tcp", bind)
		if err != nil {
			panic(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}
			go func() {
				s, err := sctp.Server(sctp.Config{
					NetConn:       conn,
					LoggerFactory: logging.NewDefaultLoggerFactory(),
				})
				if err != nil {
					panic(err)
				}
				for {
					stream, err := s.AcceptStream()
					go func() {
						if err != nil {
							panic(err)
						}
						if err != nil {
							panic(err)
						}
						n, err := io.Copy(ioutil.Discard, stream)
						log.Println(err, n)
					}()
				}
			}()
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&bind, "bind", "b", "", "address to bind to, host:port")
}
