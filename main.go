package main

import (
/*
#cgo CFLAGS: -Wall -Wextra -g
#cgo LDFLAGS: -lavcodec -lavutil -lswresample

#include <libavcodec/avcodec.h>
#include "aac_decoder.h"
#include "opus_encoder.h"
*/
    "C"
    "os"
    "io"
    "fmt"
    "log"
    "net"
	"time"
    "github.com/yutopp/go-rtmp"
	janus "github.com/notedit/janus-go"
	"github.com/pion/webrtc/v2"
    "github.com/pion/rtp/codecs"
)

func main() {
    C.avcodec_register_all()
    C.aac_decoder_init()
    C.opus_encoder_init()

    // create mediaengine and register codecs

    mediaEngine := webrtc.MediaEngine{}

    // opus
    mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))

    // custom h264 codec - need to indicate constrained baseline profile
    // with "profile-level-id=42e01f" (default is 42001f - unconstrained baseline)
    mediaEngine.RegisterCodec(webrtc.NewRTPCodec(webrtc.RTPCodecTypeVideo,
          webrtc.H264,
          90000,
          0,
          "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
          webrtc.DefaultPayloadTypeH264,
          &codecs.H264Payloader{}))

    if len(os.Args) < 3 {
        fmt.Printf("Usage: %s <listen-addr> <url>\n",os.Args[0])
        os.Exit(1)
    }

    tcpAddr, err := net.ResolveTCPAddr("tcp",os.Args[1])
    if err != nil {
        fmt.Printf("Failed: %v\n", err)
    }

    listener, err := net.ListenTCP("tcp",tcpAddr)

    if err != nil {
        fmt.Printf("Failed: %v\n", err)
    }


    // connect to janus
    gateway, err := janus.Connect(os.Args[2])
    if err != nil {
        fmt.Printf("Failed to connect to Janus: %s\n",err)
        os.Exit(1)
    }

    // create a session
    session, err := gateway.Create()
    if err != nil {
        fmt.Printf("Failed to create Janus session: %s\n", err)
        os.Exit(1)
    }

    // start a keepliave timer for the session
	go func() {
		for {
			if _, keepAliveErr := session.KeepAlive(); err != nil {
				panic(keepAliveErr)
			}

			time.Sleep(5 * time.Second)
		}
	}()


    srv := rtmp.NewServer(&rtmp.ServerConfig {
        OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
            h := &RtmpHandler{}
            h.session = session
            h.m = mediaEngine

            return conn, &rtmp.ConnConfig{
                Handler: h,
                ControlState: rtmp.StreamControlStateConfig {
                    DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
                },
            }
        },
    })

    log.Println("Listening for RTMP connections on",os.Args[1])

    if err := srv.Serve(listener); err != nil {
        fmt.Printf("Failed: %v\n", err)
    }

    session.Destroy()
    gateway.Close()

}

