package main

import (
    "strconv"
    "errors"
    "bytes"
    "io"
    "fmt"
	"log"
    flvtag "github.com/yutopp/go-flv/tag"
    "github.com/yutopp/go-rtmp"
    rtmpmsg "github.com/yutopp/go-rtmp/message"
	janus "github.com/notedit/janus-go"
	"github.com/pion/webrtc/v2"
)

var _ rtmp.Handler = (*RtmpHandler)(nil)

type RtmpHandler struct {
    rtmp.DefaultHandler
    roomId uint64
    pc *webrtc.PeerConnection
    m webrtc.MediaEngine
    audioTrack *webrtc.Track
    videoTrack *webrtc.Track
    videoHandler *VideoHandler
    audioHandler *AudioHandler
    session *janus.Session
    handle *janus.Handle
}

func watchHandle(handle *janus.Handle) {
	// wait for event
	for {
		msg := <-handle.Events
		switch msg := msg.(type) {
		case *janus.SlowLinkMsg:
			log.Println("SlowLinkMsg type ", handle.ID)
		case *janus.MediaMsg:
			log.Println("MediaEvent type", msg.Type, " receiving ", msg.Receiving)
		case *janus.WebRTCUpMsg:
			log.Println("WebRTCUp type ", handle.ID)
		case *janus.HangupMsg:
			log.Println("HangupEvent type ", handle.ID)
		case *janus.EventMsg:
			log.Printf("EventMsg %+v", msg.Plugindata.Data)
		}
	}
}

func connectJanus(h *RtmpHandler) error {
    var err error = nil
    var offer webrtc.SessionDescription
    var msg *janus.EventMsg = nil

    log.Println("Connecting to Videoroom",h.roomId)

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}

	// Create a new RTCPeerConnection
    h.pc, err = webrtc.NewAPI(webrtc.WithMediaEngine(h.m)).NewPeerConnection(config)
	// h.pc, err = webrtc.NewPeerConnection(config)
	if err != nil {
        fmt.Printf("Failed to create peerconnection: %s\n",err)
        return err
	}

	h.pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Println("Connection State has changed", connectionState.String())
	})

	// Create audio track
	h.audioTrack, err = h.pc.NewTrack(webrtc.DefaultPayloadTypeOpus, RandUint32(), MathRandAlpha(16), MathRandAlpha(16))
	if err != nil {
        fmt.Printf("Failed to create audiotrack: %s\n", err)
        return err
	}
	_, err = h.pc.AddTrack(h.audioTrack)
	if err != nil {
        fmt.Printf("Failed to add audiotrack: %s\n", err)
        return err
	}

	// Create video track
	h.videoTrack, err = h.pc.NewTrack(webrtc.DefaultPayloadTypeH264, RandUint32(), MathRandAlpha(16), MathRandAlpha(16))
	if err != nil {
        fmt.Printf("Failed to create videotrack: %s\n", err)
		return err
	}
	_, err = h.pc.AddTrack(h.videoTrack)
	if err != nil {
        fmt.Printf("Failed to add videotrack: %s\n", err)
        return err
	}

	offer, err = h.pc.CreateOffer(nil)
	if err != nil {
        fmt.Printf("Failed to create offer: %s\n",err)
        return err
	}

	err = h.pc.SetLocalDescription(offer)
	if err != nil {
        fmt.Printf("Failed to setlocaldescription",err)
        return err
	}

	h.handle, err = h.session.Attach("janus.plugin.videoroom")
	if err != nil {
        fmt.Printf("Failed to attach",err)
        return err
	}

	go watchHandle(h.handle)

	_, err = h.handle.Message(map[string]interface{}{
		"request": "join",
		"ptype":   "publisher",
		"room":    h.roomId,
		"id":      RandUint32(),
	}, nil)
	if err != nil {
        fmt.Printf("Failed to send handle join: %s\n", err)
        return err
	}

	msg, err = h.handle.Message(map[string]interface{}{
		"request": "publish",
		"audio":   true,
		"video":   true,
		"data":    false,
        "videocodec": "h264",
        "audiocodec": "opus",
	}, map[string]interface{}{
		"type":    "offer",
		"sdp":     offer.SDP,
		"trickle": false,
	})
	if err != nil {
        fmt.Printf("Failed to send handle publish: %s\n", err)
        return err
	}

	if msg.Jsep != nil {
		err = h.pc.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  msg.Jsep["sdp"].(string),
		})
		if err != nil {
            fmt.Printf("Failed to setremotedescription: %s\n", err)
            return err
		}
    }

    h.videoHandler.videoTrack = h.videoTrack
    h.audioHandler.audioTrack = h.audioTrack

    log.Println("Connected to videoroom", h.roomId)

    return nil
}


func (h *RtmpHandler) OnServe(conn *rtmp.Conn) {
    h.videoHandler = NewVideoHandler()
    h.audioHandler = NewAudioHandler()
    h.handle = nil
    h.pc = nil
}

func (h *RtmpHandler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
    if len(cmd.Command.App) == 0 {
        return errors.New("no app given")
    }
    return nil
}

func (h *RtmpHandler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
    return nil
}

func (h *RtmpHandler) OnPublish(timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {

    var s uint64 = 0
    var err error = nil

    if len(cmd.PublishingName) == 0 {
        fmt.Printf("no publishing name given\n")
        return errors.New("no publishing name given")
    }

    if s, err = strconv.ParseUint(cmd.PublishingName, 10, 64); err != nil {
        fmt.Printf("publishing name not uint64\n")
        return errors.New("publishing name not uint64")
    }

    h.roomId = s

    if err := connectJanus(h); err != nil {
        fmt.Printf("Failure in connectJanus: %s\n",err)
        return err
    }

    return nil
}

func (h *RtmpHandler) OnSetDataFrame(timestamp uint32, data *rtmpmsg.NetStreamSetDataFrame) error {
    r := bytes.NewReader(data.Payload)
    var script flvtag.ScriptData
    if err := flvtag.DecodeScriptData(r,&script); err != nil {
        fmt.Printf("Failed to decode script data: %s\n",err)
        return nil
    }

    framerate := script.Objects["onMetaData"]["framerate"]
    h.videoHandler.fps = framerate.(float64)
    h.videoHandler.spf = uint32(90000.0 / h.videoHandler.fps)

    return nil
}

func (h *RtmpHandler) OnAudio(timestamp uint32, payload io.Reader) error {
   var audio flvtag.AudioData
   if err := flvtag.DecodeAudioData(payload,&audio); err != nil {
       return err
   }

   if err := h.audioHandler.Push(&audio); err != nil {
       fmt.Printf("OnAudio failure: %s\n",err)
       return err
   }

   return nil
}


func (h *RtmpHandler) OnVideo(timestamp uint32, payload io.Reader) error {
   var video flvtag.VideoData
   if err := flvtag.DecodeVideoData(payload,&video); err != nil {
       return err
   }

   h.videoHandler.Push(&video)
   return nil
}

func (h *RtmpHandler) OnClose() {
    if h.handle != nil {
        h.handle.Detach()
    }
    h.videoHandler.Close()
    h.audioHandler.Close()
}
