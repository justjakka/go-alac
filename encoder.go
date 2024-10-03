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

type FrameEncoder struct {
	handle C.alac_encoder
}

func NewFrameEncoder(sampleRate, channels, bitDepth int, fastMode bool) *FrameEncoder {
	fastInt := 0
	if fastMode {
		fastInt = 1
	}
	handle := C.alac_encoder_new(C.int(sampleRate), C.int(channels), C.int(bitDepth), C.int(fastInt))

	if handle.ptr == nil {
		return nil
	}

	e := &FrameEncoder{
		handle: handle,
	}

	runtime.SetFinalizer(e, finalizeEncoder)

	return e
}

func (e *FrameEncoder) GetMagicCookie() []byte {
	output := make([]byte, int(e.handle.magic_cookie_size))
	outBytes := C.alac_encoder_get_magic_cookie(&e.handle, (*C.uchar)(unsafe.Pointer(&output[0])))
	return output[:outBytes]
}

func (e *FrameEncoder) GetInputSize() int {
	return int(e.handle.input_packet_size)
}

func (e *FrameEncoder) GetSamplesPerPacket() int {
	return int(e.handle.frames_per_packet)
}

func (e *FrameEncoder) WritePacket(pcm []byte) []byte {
	output := make([]byte, int(e.handle.output_max_packet_size))
	outBytes := C.alac_encoder_write(&e.handle, (*C.uchar)(unsafe.Pointer(&pcm[0])), C.int(len(pcm)), (*C.uchar)(unsafe.Pointer(&output[0])))
	if outBytes < 0 {
		return nil
	}
	return output[:outBytes]
}

func finalizeEncoder(e *FrameEncoder) {
	e.Close()
}

func (e *FrameEncoder) Close() {
	C.alac_encoder_delete(&e.handle)
}
