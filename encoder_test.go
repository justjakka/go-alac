package go_alac

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	f, err := os.Open("sample/test_s16.raw")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
		return
	}

	o, err := os.Create("sample/test_s16_output.alac")
	if err != nil {
		t.Error(err)
		return
	}
	defer o.Close()

	writeTestCase := func(n int, cookie []byte, data []byte) {
		out, err := os.Create(fmt.Sprintf("/home/shoghicp/radio/alac_afl/test/packets_%d.alac", n))
		if err != nil {
			t.Error(err)
			return
		}
		defer out.Close()
		out.Write(cookie)
		out.Write(data)
	}

	encoder := NewFrameEncoder(44100, 2, 16, false)
	o.Write(encoder.GetMagicCookie())
	packetSize := encoder.GetInputSize()
	packets := 0
	for len(data) > packetSize {
		resultPacket := encoder.WritePacket(data[:packetSize])
		o.Write(resultPacket)
		data = data[packetSize:]

		writeTestCase(packets, encoder.GetMagicCookie(), resultPacket)
		packets++
	}
	return

	if len(data) > 0 {
		resultPacket := encoder.WritePacket(data)
		o.Write(resultPacket)
	}

	o.Sync()
	o.Close()
}
