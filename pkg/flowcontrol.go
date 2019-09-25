package pkg

import (
	"github.com/pion/sctp"
	"io"
	"log"
)

type FlowControlledStream struct {
	stream                     *sctp.Stream
	bufferedAmountLowThreshold uint64
	maxBufferedAmount          uint64
}

type FlowControlledStreamSignal struct {
	FlowControlledStream
	bufferedAmountLowSignal chan struct{}
}

type FlowControlledStreamDrain struct {
	FlowControlledStream
	queue chan []byte
}

type FlowControlledStreamSpinCPU struct {
	FlowControlledStream
}

// NewFlowControlledStream --
func NewFlowControlledStream(flowControlType string, stream *sctp.Stream, bufferedAmountLowThreshold, maxBufferedAmount, queueSize uint64) io.ReadWriter {
	var rw io.ReadWriter

	stream.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)
	fcs := FlowControlledStream{
		stream:                     stream,
		bufferedAmountLowThreshold: bufferedAmountLowThreshold,
		maxBufferedAmount:          maxBufferedAmount,
	}

	switch flowControlType {
	case "none":
		rw = fcs
	case "signal":
		fcss := &FlowControlledStreamSignal{
			FlowControlledStream:    fcs,
			bufferedAmountLowSignal: make(chan struct{}, 1024),
		}
		stream.OnBufferedAmountLow(func() {
			fcss.bufferedAmountLowSignal <- struct{}{}
		})
		rw = fcss
	case "drain":
		fcsd := &FlowControlledStreamDrain{
			FlowControlledStream: fcs,
			queue:                make(chan []byte, queueSize),
		}
		stream.OnBufferedAmountLow(func() {
			go fcsd.DrainQueue()
		})
		rw = fcsd
	case "spin-cpu":
		fcsscpu := &FlowControlledStreamSpinCPU{
			FlowControlledStream: fcs,
		}
		rw = fcsscpu
	}

	return rw
}

func (fcdc FlowControlledStream) Read(p []byte) (int, error) {
	return fcdc.stream.Read(p)
}

func (fcdc FlowControlledStream) Write(p []byte) (int, error) {
	return fcdc.stream.Write(p)
}

func (fcdc *FlowControlledStreamSignal) Read(p []byte) (int, error) {
	return fcdc.stream.Read(p)
}

func (fcdc *FlowControlledStreamSignal) Write(p []byte) (int, error) {
	if fcdc.stream.BufferedAmount() > fcdc.maxBufferedAmount {
		<-fcdc.bufferedAmountLowSignal
	}
	return fcdc.stream.Write(p)
}

func (fcdc *FlowControlledStreamDrain) Read(p []byte) (int, error) {
	return fcdc.stream.Read(p)
}

func (fcdc *FlowControlledStreamDrain) Write(p []byte) (int, error) {
	fcdc.queue <- p
	if _, err := fcdc.DrainQueue(); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (fcdc *FlowControlledStreamDrain) DrainQueue() (int, error) {
	var bytesSent int
	for {
		if len(fcdc.queue) == 0 {
			break
		}

		if fcdc.stream.BufferedAmount() >= fcdc.maxBufferedAmount {
			break
		}

		p := <-fcdc.queue
		b, err := fcdc.stream.Write(p)
		if err != nil {
			log.Println("ERROR", err)
			return bytesSent, err
		}

		bytesSent += b
	}
	return bytesSent, nil
}

func (fcdc *FlowControlledStreamSpinCPU) Read(p []byte) (int, error) {
	return fcdc.stream.Read(p)
}

func (fcsscpu *FlowControlledStreamSpinCPU) Write(p []byte) (int, error) {
	for fcsscpu.stream.BufferedAmount() > fcsscpu.maxBufferedAmount {
	}
	return fcsscpu.stream.Write(p)
}
