package go_alac

import (
	"bytes"
	"github.com/Eyevinn/mp4ff/bits"
	"github.com/Eyevinn/mp4ff/mp4"
	"io"
	"time"
)

type FormatEncoder struct {
	encoder          *FrameEncoder
	writer           io.Writer
	sampleRate       int
	buffer           []byte
	outputBuffer     [][]byte
	segmentDuration  time.Duration
	samplesPerPacket int
	packetsWritten   int
	trackId          uint32
	seqNumber        uint32
}

type AlacBox struct {
	Cookie []byte
}

func (b *AlacBox) Type() string {
	return "alac"
}

// Size - calculated size of box
func (b *AlacBox) Size() uint64 {
	return uint64(12 + len(b.Cookie))
}

// Encode - write box to w
func (b *AlacBox) Encode(w io.Writer) error {
	sw := bits.NewFixedSliceWriter(int(b.Size()))
	err := b.EncodeSW(sw)
	if err != nil {
		return err
	}
	_, err = w.Write(sw.Bytes())
	return err
}

func (b *AlacBox) EncodeSW(sw bits.SliceWriter) error {
	err := mp4.EncodeHeaderSW(b, sw)
	if err != nil {
		return err
	}
	sw.WriteUint32(0) //version
	sw.WriteBytes(b.Cookie)
	return sw.AccError()
}

func (b *AlacBox) Info(w io.Writer, specificBoxLevels, indent, indentStep string) error {
	return nil
}

func NewFormatEncoder(writer io.Writer, sampleRate, channels, bitDepth int, fastMode bool, segmentDuration time.Duration) *FormatEncoder {
	e := &FormatEncoder{
		encoder:         NewFrameEncoder(sampleRate, channels, bitDepth, fastMode),
		writer:          writer,
		sampleRate:      sampleRate,
		segmentDuration: segmentDuration,
	}

	if e.encoder == nil {
		return nil
	}

	e.samplesPerPacket = e.encoder.GetSamplesPerPacket()

	init := mp4.CreateEmptyInit()
	init.AddEmptyTrack(uint32(sampleRate), "audio", "en")
	e.trackId = init.Moov.Mvhd.NextTrackID - 1
	trak := init.Moov.Trak

	stsd := trak.Mdia.Minf.Stbl.Stsd

	//TODO: this does not work with 96kHz freq etc.
	mp4a := mp4.CreateAudioSampleEntryBox("alac", uint16(channels), uint16(bitDepth), uint16(sampleRate), &AlacBox{
		Cookie: e.encoder.GetMagicCookie(),
	})
	stsd.AddChild(mp4a)

	init.Encode(writer)

	return e
}

func (e *FormatEncoder) outputPacket(packet []byte) {
	e.outputBuffer = append(e.outputBuffer, packet)

	if time.Duration(float64(time.Second)*(float64(e.samplesPerPacket*len(e.outputBuffer))/float64(e.sampleRate))) >= e.segmentDuration {
		e.outputSegment()
	}

}

func (e *FormatEncoder) outputSegment() {
	seg := mp4.NewMediaSegment()
	frag, _ := mp4.CreateFragment(e.seqNumber, e.trackId)
	seg.AddFragment(frag)

	for _, b := range e.outputBuffer {
		frag.AddFullSampleToTrack(mp4.FullSample{
			Sample: mp4.Sample{
				Dur:  uint32(e.samplesPerPacket),
				Size: uint32(len(b)),
			},
			DecodeTime: uint64(e.samplesPerPacket * e.packetsWritten),
			Data:       b,
		}, e.trackId)
		e.packetsWritten++
	}

	seg.Encode(e.writer)
	e.seqNumber++
	e.outputBuffer = nil

}

func (e *FormatEncoder) Write(pcm []byte) {
	if e.encoder == nil {
		return
	}
	e.buffer = append(e.buffer, pcm...)
	inputSize := e.encoder.GetInputSize()
	for len(e.buffer) >= inputSize {
		e.outputPacket(e.encoder.WritePacket(e.buffer[:inputSize]))
		e.buffer = e.buffer[inputSize:]
	}
}
func (e *FormatEncoder) Flush() {
	if e.encoder == nil {
		return
	}

	if len(e.buffer) > 0 {
		e.outputPacket(e.encoder.WritePacket(e.buffer))
		e.buffer = nil
	}

	if len(e.outputBuffer) > 0 {
		e.outputSegment()
	}
}

type FormatDecoder struct {
	trex      *mp4.TrexBox
	handle    *FrameDecoder
	parsedMp4 *mp4.File

	currentSegment  int
	currentFragment int
}

func NewFormatDecoder(reader io.Reader) *FormatDecoder {
	parsedMp4, err := mp4.DecodeFile(reader)
	if err != nil {
		return nil
	}

	var trexEntry *mp4.TrexBox
	var magicCookie []byte

	//TODO: handle non-segmented

	for _, trak := range parsedMp4.Moov.Traks {
		if box, err := trak.Mdia.Minf.Stbl.Stsd.GetSampleDescription(0); err == nil && box.Type() == "alac" {
			if parsedMp4.Moov.Mvex == nil {
				continue
			}
			for _, trex := range parsedMp4.Moov.Mvex.Trexs {
				if trex.TrackID == trak.Tkhd.TrackID {
					trexEntry = trex
					break
				}
			}

			if trexEntry == nil {
				return nil
			}

			buf := new(bytes.Buffer)
			box.Encode(buf)

			boxBytes := buf.Bytes()

			boxOffset := 36 + 12

			magicCookie = boxBytes[boxOffset:]
			break
		}
	}

	if trexEntry == nil || magicCookie == nil {
		return nil
	}

	decoder := NewFrameDecoder(magicCookie)
	if decoder == nil {
		return nil
	}

	return &FormatDecoder{
		handle:    decoder,
		trex:      trexEntry,
		parsedMp4: parsedMp4,
	}
}

func (d *FormatDecoder) GetChannels() int {
	return d.handle.GetChannels()
}

func (d *FormatDecoder) GetBitDepth() int {
	return d.handle.GetBitDepth()
}

func (d *FormatDecoder) GetSampleRate() int {
	return d.handle.GetSampleRate()
}

func (d *FormatDecoder) Read() (buf []byte) {
	if d.currentSegment >= len(d.parsedMp4.Segments) {
		//EOF
		return nil
	}
	segment := d.parsedMp4.Segments[d.currentSegment]

	if d.currentFragment >= len(segment.Fragments) {
		d.currentSegment++
		d.currentFragment = 0
		return d.Read()
	}

	frag := segment.Fragments[d.currentFragment]

	samples, err := frag.GetFullSamples(d.trex)
	if err != nil {
		return nil
	}

	for _, sample := range samples {
		_, pcm := d.handle.ReadPacket(sample.Data)

		buf = append(buf, pcm...)
	}
	d.currentFragment++

	return
}
