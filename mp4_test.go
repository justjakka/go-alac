package go_alac

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestEncodeMP4(t *testing.T) {
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

	o, err := os.Create("sample/test_s16_output.m4a")
	if err != nil {
		t.Error(err)
		return
	}
	defer o.Close()

	encoder := NewFormatEncoder(o, 44100, 2, 16, false, time.Millisecond*100)

	iterationSize := 65536
	for len(data) > iterationSize {
		encoder.Write(data[:iterationSize])
		data = data[iterationSize:]
	}

	if len(data) > 0 {
		encoder.Write(data)
	}

	encoder.Flush()

	o.Sync()
}

func TestDecodeMP4(t *testing.T) {
	t.Parallel()

	f, err := os.Open("sample/test_s16_input.m4a")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	o, err := os.Create("sample/test_s16_output.raw")
	if err != nil {
		t.Error(err)
		return
	}
	defer o.Close()

	decoder := NewFormatDecoder(f)
	if decoder == nil {
		t.Fail()
		return
	}
	for {
		data := decoder.Read()
		if data == nil {
			break
		}
		o.Write(data)
	}
	o.Sync()
}
