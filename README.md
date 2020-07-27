# RTMP to Janus Videoroom

## Warning I am not a Go programmer

This is very much in the work-in-progress/proof-of-concept phase!

I really don't know go, I learned just enough so I could make use of the [`pion/webrtc`](https://github.com/pion/webrtc) library.

This is adapted from the Janus example in Pion's [`example-webrtc-applications`](https://github.com/pion/example-webrtc-applications) repo.

## Usage

```bash
rtmp-janus <listen host:port> ws://janus-host:port

# example: rtmp-janus :1935 ws://127.0.0.1:8188
```

## What does this do?

When launched, this app:

* Starts listening for incoming RTMP sessions.
* Connects to Janus gateway via websocket, establishes a session.

When an RTMP connection is received, it parses the RTMP "key" for a room ID.

Example: `rtmp://127.0.0.1:1935/live/1234` - the `1234` part of that URL becomes the room ID.

The `live` part of the URL (the application name) can be whatever you'd like, it's ignored.

The app then joins the janus videoroom and establishes a WebRTC session.

I'm assuming all incoming RTMP sessions are using H264 for video and AAC for audio.

H264 data is re-packed from FLV format into Annex-B format, but otherwise passed
as-is (no decoding/encoding).

AAC audio is resampled to 48000kHz, stereo audio and encoded to Opus using ffmpeg.

## TODO

* Figure out what events from Janus I should handle (I just blast audio/video).
* Check more sources for timing info (like the SPS NAL).
* Support more audio samplerates (maybe?)

## LICENSE

MIT (see `LICENSE`)
