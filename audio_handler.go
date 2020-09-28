package main


import (
/*
#cgo CFLAGS: -g
#include "aac_decoder.h"
#include "opus_encoder.h"
#include "audio_fifo.h"
*/
    "C"
    "errors"
    "bytes"
    "unsafe"
	"github.com/pion/webrtc/v2"
    "github.com/pion/webrtc/v2/pkg/media"
    flvtag "github.com/yutopp/go-flv/tag"
    "fmt"
    //"os"
)

type AudioHandler struct {
    decoder *C.rtmpjanus_aac_decoder_t
    encoder *C.rtmpjanus_opus_encoder_t
    fifo *C.rtmpjanus_audio_fifo_t
    audioTrack *webrtc.Track
    samplerate uint32
    channels uint32
}

func NewAudioHandler() *AudioHandler {
    h := new(AudioHandler)
    h.decoder = C.rtmpjanus_aac_decoder_new()
    h.encoder = C.rtmpjanus_opus_encoder_new(2)
    h.fifo    = C.rtmpjanus_audio_fifo_new(48000,2)
    h.samplerate = 0
    h.channels = 0
    return h
}


func (h *AudioHandler) Close() {
    C.rtmpjanus_aac_decoder_close(h.decoder)
    C.rtmpjanus_opus_encoder_close(h.encoder)
    C.rtmpjanus_audio_fifo_close(h.fifo)
}

var freqindex = [13]uint32 {
    96000,
    88200,
    64000,
    48000,
    44100,
    32000,
    24000,
    22050,
    16000,
    12000,
    11025,
     8000,
     7350 }

//var counter uint32 = 0

func (h *AudioHandler) Push(audio *flvtag.AudioData) error {

   packet := new(bytes.Buffer)
   packet.ReadFrom(audio.Data)

   if audio.AACPacketType == 0 {

       header := packet.Bytes()

/*
       f, f_err := os.Create("aac-header/data.raw")
       if f_err == nil {
         f.Write(header)
         f.Close()
       }
       */

       // audioObjectType := ((h.header[0] & 0xF1) >> 3) - 1
       audioFreqIndex  := ((header[0] & 0x07) << 1) | ((header[1] & 0x80) >> 7)
       audioChannelConfig := ((header[1] & 0x78) >> 3)

       if audioFreqIndex < 13 {
           h.samplerate = freqindex[audioFreqIndex]
       } else {
           fmt.Printf("error: unsupported audio frequence")
           return errors.New("Unsupported audio frequency")
       }

       if audioChannelConfig > 0 && audioChannelConfig < 8 {
           h.channels = uint32(audioChannelConfig)
       } else {
           fmt.Printf("error: unsupported audio channels")
           return errors.New("Unsupported audio channels")
       }

       C.rtmpjanus_aac_decoder_open(h.decoder,(*C.uint8_t)(&header[0]),C.size_t(len(header)))
       C.rtmpjanus_audio_fifo_open(h.fifo,C.uint32_t(h.samplerate),C.uint32_t(h.channels))

       return nil
   }

   // if we haven't seen the config packet yet, discard
   if h.samplerate != 0 {
       b := packet.Bytes()

/*
       filename := fmt.Sprintf("aac-packets/%05d.raw",counter)
       f, f_err := os.Create(filename)
       if f_err == nil {
           f.Write(b)
           f.Close()
           counter = counter + 1
       }
       */

       dec_frame := C.rtmpjanus_aac_decoder_decode(h.decoder,(*C.uint8_t)(&b[0]),C.size_t(len(b)))
       if dec_frame == nil {
           fmt.Printf("Error decoding AAC audio")
           return errors.New("Error decoding AAC audio")
       }

       C.rtmpjanus_audio_fifo_load(h.fifo,dec_frame)

       for C.rtmpjanus_audio_fifo_size(h.fifo) >= h.encoder.neededSamples {
           raw_frame := C.rtmpjanus_audio_fifo_read(h.fifo,h.encoder.neededSamples)
           packet := C.rtmpjanus_opus_encoder_encode(h.encoder,raw_frame)
           sample := make([]byte, packet.size)
           C.memcpy(
             unsafe.Pointer(&sample[0]),
             unsafe.Pointer(packet.data),
             C.size_t(packet.size))
           h.audioTrack.WriteSample(media.Sample{Data: sample, Samples: uint32(h.encoder.neededSamples)})
       }
   }
   return nil

}
