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
	"log"
	"math/rand"
	"net"
	"time"
)

var (
	server      string
	flowcontrol string
	queueSize   uint64
)

func PrintStatistics(association *sctp.Association) {
	since := time.Now()
	for range time.NewTicker(1000 * time.Millisecond).C {
		rbps := float64(association.BytesReceived()*8) / time.Since(since).Seconds()
		log.Printf("Received Mbps: %.03f, totalBytesReceived: %d", rbps/1024/1024, association.BytesReceived())
		sbps := float64(association.BytesSent()*8) / time.Since(since).Seconds()
		log.Printf("Sent Mbps: %.03f, totalBytesSent: %d", sbps/1024/1024, association.BytesSent())
	}
}

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use: "client",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := net.Dial("tcp", server)
		if err != nil {
			panic(err)
		}
		log.Println("Dialed conn")
		sctpClient, err := sctp.Client(sctp.Config{
			NetConn:       c,
			LoggerFactory: logging.NewDefaultLoggerFactory(),
		})
		log.Println("Dialed sctp")
		if err != nil {
			panic(err)
		}
		go PrintStatistics(sctpClient)
		stream, err := sctpClient.OpenStream(0, sctp.PayloadTypeWebRTCBinary)
		if err != nil {
			panic(err)
		}
		log.Println("Opened stream")
		src := rand.NewSource(int64(123))
		r := rand.New(src)
		fc := pkg.NewFlowControlledStream(flowcontrol, stream, 512*1024, 1024*10240, queueSize)
		_, err = io.Copy(fc, r)
		panic(err)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&server, "server", "s", "", "address of server, host:port")
	clientCmd.Flags().StringVarP(&flowcontrol, "flowcontrol", "f", "signal", "flow control strategy")
	clientCmd.Flags().Uint64VarP(&queueSize, "queue-size", "q", 100, "queue size for drain flow control strategy")
}
