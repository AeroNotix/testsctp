package pkg

import (
	"github.com/pion/sctp"
	"log"
	"sync/atomic"
	"time"
)

type FlowControlledStream struct {
	stream                     *sctp.Stream
	bufferedAmountLowThreshold uint64
	maxBufferedAmount          uint64
	bufferedAmountLowSignal    chan struct{}
	totalBytesReceived         uint64
	totalBytesSent             uint64
}

// NewFlowControlledStream --
func NewFlowControlledStream(stream *sctp.Stream, bufferedAmountLowThreshold, maxBufferedAmount uint64) *FlowControlledStream {
	fcdc := &FlowControlledStream{
		stream:                     stream,
		bufferedAmountLowThreshold: bufferedAmountLowThreshold,
		maxBufferedAmount:          maxBufferedAmount,
		bufferedAmountLowSignal:    make(chan struct{}, 1),
	}
	stream.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)
	stream.OnBufferedAmountLow(func() {
		go func() { fcdc.bufferedAmountLowSignal <- struct{}{} }()
	})
	go func() {
		since := time.Now()

		// Start printing out the observed throughput
		for range time.NewTicker(1000 * time.Millisecond).C {
			rbps := float64(atomic.LoadUint64(&fcdc.totalBytesReceived)*8) / time.Since(since).Seconds()
			log.Printf("RecvThroughPut: %.03f Mbps, totalBytesReceived: %d", rbps/1024/1024, fcdc.totalBytesReceived)
			wbps := float64(atomic.LoadUint64(&fcdc.totalBytesSent)*8) / time.Since(since).Seconds()
			log.Printf("SendThroughPut: %.03f Mbps, totalBytesSent: %d", wbps/1024/1024, fcdc.totalBytesSent)
		}
	}()

	return fcdc
}

func (fcdc *FlowControlledStream) Read(p []byte) (int, error) {
	cnt, err := fcdc.stream.Read(p)
	atomic.AddUint64(&fcdc.totalBytesReceived, uint64(cnt))
	return cnt, err
}

func (fcdc *FlowControlledStream) Write(p []byte) (int, error) {

	if fcdc.stream.BufferedAmount() > fcdc.maxBufferedAmount {
		<-fcdc.bufferedAmountLowSignal
	}

	cnt, err := fcdc.stream.Write(p)
	if err != nil {
		return cnt, err
	}
	atomic.AddUint64(&fcdc.totalBytesSent, uint64(cnt))
	return cnt, err
}
