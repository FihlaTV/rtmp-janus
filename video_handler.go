package main

import (
    "bytes"
	"github.com/pion/webrtc/v2"
    "github.com/pion/webrtc/v2/pkg/media"
    flvtag "github.com/yutopp/go-flv/tag"
)

var nalu_header  = [4]byte { 0x00, 0x00, 0x00, 0x01 }

type VideoHandler struct {
    buffer *bytes.Buffer
    videoTrack *webrtc.Track
    nextLen uint32
    version uint8
    profile uint8
    compatibility uint8
    level uint8
    nalLen uint8
    sps *bytes.Buffer
    pps *bytes.Buffer
    nal *bytes.Buffer
    fps float64
    spf uint32 // samples-per-frame
    //f *os.File
}

func NewVideoHandler() *VideoHandler {
    //var err error
    p := new(VideoHandler)
    p.buffer = new(bytes.Buffer)
    p.sps = new(bytes.Buffer)
    p.pps = new(bytes.Buffer)
    p.nal = new(bytes.Buffer)
    p.sps.Write(nalu_header[:])
    p.pps.Write(nalu_header[:])
    p.nextLen = 0
    p.version = 0
    p.profile = 0
    p.compatibility = 0
    p.level = 0
    p.nalLen = 0
    p.fps = 0.0
    p.spf = 0
    return p
}

func (p *VideoHandler) Close() {
    //p.f.Close()
}

func (p *VideoHandler) NextLen() uint32 {
    var i uint = 0

    if p.nextLen > 0 {
        return p.nextLen
    }

    if p.buffer.Len() < int(p.nalLen) {
        return 0
    }

    t := p.buffer.Next(int(p.nalLen))

    p.nextLen = 0

    for i < uint(p.nalLen) {
        p.nextLen = p.nextLen << 8
        p.nextLen = p.nextLen + uint32(t[i])
        i = i + 1
    }

    return p.nextLen
}

func (p *VideoHandler) ProcessFrames() {
    var err error
    p.nal.Reset()

    for p.NextLen() > 0 && uint32(p.buffer.Len()) >= p.nextLen {

        nal := p.buffer.Next(int(p.nextLen))
        nal_type := nal[0] & 0x1F

        if nal_type == 5 {
            p.nal.Write(p.sps.Bytes())
        }
        if nal_type == 1 || nal_type == 5 {
            p.nal.Write(p.pps.Bytes())
            p.nal.Write(nalu_header[:])
            p.nal.Write(nal[:])
        }
        p.nextLen = 0
    }
    err = p.videoTrack.WriteSample(media.Sample{Data: p.nal.Bytes(), Samples: p.spf})
    if err != nil  {
        panic(err)
    }
}

func (p *VideoHandler) Push(video *flvtag.VideoData) error {
    if(video.AVCPacketType == 0) {
        /*
        bits    
        8   version ( always 0x01 )
        --- 0
        8   avc profile ( sps[0][1] )
        --- 1
        8   avc compatibility ( sps[0][2] )
        --- 2
        8   avc level ( sps[0][3] )
        --- 3
        6   reserved ( all bits on )
        2   NALULengthSizeMinusOne
        --- 4
        3   reserved ( all bits on )
        5   number of SPS NALUs (usually 1)
        --- 5
        repeated once per SPS:
            16         SPS size
            variable   SPS NALU data
        8   number of PPS NALUs (usually 1)
        repeated once per PPS
          16         PPS size
          variable   PPS NALU data
        */
        var tmpLen uint16 = 0
        var offset uint32 = 0
        var i uint8 = 0
        packet := new(bytes.Buffer)
        packet.ReadFrom(video.Data)

        b := packet.Bytes()

        offset = 0
        p.version = b[offset]
        offset = offset + 1

        p.profile = b[offset]
        offset = offset + 1

        p.compatibility = b[offset]
        offset = offset + 1

        p.level = b[offset]
        offset = offset + 1

        p.nalLen = (b[offset] & 0x03) + 1
        offset = offset + 1

        spsNum := (b[offset] & 0x1F)
        offset = offset + 1

        for i < spsNum {
            tmpLen = uint16(b[offset]) << 8
            offset = offset + 1
            tmpLen = uint16(b[offset]) + tmpLen
            offset = offset + 1
            if i == 0 { // only read first sps
                p.sps.Write(b[offset:offset + uint32(tmpLen)])
            }
            offset = offset + uint32(tmpLen)

            i = i + 1
        }

        ppsNum := b[offset]
        offset = offset + 1

        i = 0

        for i < ppsNum {
            tmpLen = uint16(b[offset]) << 8
            offset = offset + 1
            tmpLen = uint16(b[offset]) + tmpLen
            offset = offset + 1
            if i == 0 { // only read first pps
                p.pps.Write(b[offset:offset + uint32(tmpLen)])
            }
            offset = offset + uint32(tmpLen)

            i = i + 1
        }

    } else {
        // have not seen the config packet yet, discard
        if p.version != 0 {
            p.buffer.ReadFrom(video.Data)
            p.ProcessFrames()
        }
    }

    return nil
}
