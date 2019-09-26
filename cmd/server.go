package cmd

import (
	"github.com/pion/logging"
	"github.com/pion/sctp"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"time"
)

type NoopWriter struct{}

func (w *NoopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

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

		go func() {
			since := time.Now()
			for range time.NewTicker(1000 * time.Millisecond).C {
				rbps := float64(s.BytesReceived()*8) / time.Since(since).Seconds()
				log.Printf("Received Mbps: %.03f, totalBytesReceived: %d", rbps/1024/1024, s.BytesReceived())
			}
		}()

		for {
			stream, err := s.AcceptStream()
			if err != nil {
				panic(err)
			}
			go func() {
				w := &NoopWriter{}
				n, err := io.Copy(w, stream)
				log.Println(err, n)
				if err != nil {
					panic(err)
				}
			}()
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
