package go_alac

/*
#cgo pkg-config: alac
#include <alac/libalac.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type FrameDecoder struct {
	handle C.alac_decoder
}

func NewFrameDecoder(magicCookie []byte) *FrameDecoder {
	handle := C.alac_decoder_new((*C.uchar)(unsafe.Pointer(&magicCookie[0])), C.int(len(magicCookie)))

	if handle.ptr == nil {
		return nil
	}

	e := &FrameDecoder{
		handle: handle,
	}

	runtime.SetFinalizer(e, finalizeDecoder)

	return e
}

func (e *FrameDecoder) GetInputSize() int {
	return int(e.handle.input_max_packet_size)
}

func (e *FrameDecoder) GetSamplesPerPacket() int {
	return int(e.handle.frames_per_packet)
}

func (e *FrameDecoder) GetChannels() int {
	return int(e.handle.channels)
}

func (e *FrameDecoder) GetBitDepth() int {
	return int(e.handle.bit_depth)
}

func (e *FrameDecoder) GetSampleRate() int {
	return int(e.handle.sample_rate)
}

func (e *FrameDecoder) ReadPacket(data []byte) (inputBytesUsed int, pcm []byte) {
	output := make([]byte, int(e.handle.output_packet_size))
	outStruct := C.alac_decoder_read(&e.handle, (*C.uchar)(unsafe.Pointer(&data[0])), C.int(len(data)), (*C.uchar)(unsafe.Pointer(&output[0])))
	if outStruct.input_bytes_used < 0 {
		return 0, nil
	}
	return int(outStruct.input_bytes_used), output[:outStruct.output_bytes]
}

func finalizeDecoder(e *FrameDecoder) {
	e.Close()
}

func (e *FrameDecoder) Close() {
	C.alac_decoder_delete(&e.handle)
}
