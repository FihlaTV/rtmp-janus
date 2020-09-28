#ifndef OPUS_ENCODER_H
#define OPUS_ENCODER_H

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <stdint.h>

typedef struct {
    AVCodecContext *ctx;
    AVDictionary *opts;
    AVPacket *packet;
    uint32_t neededSamples;
    uint8_t *packetData;
    size_t   packetSize;
} rtmpjanus_opus_encoder_t;

#ifdef __cplusplus
extern "C" {
#endif

int
rtmpjanus_opus_encoder_init(void);

rtmpjanus_opus_encoder_t *
rtmpjanus_opus_encoder_new(uint32_t channels);

void
rtmpjanus_opus_encoder_close(rtmpjanus_opus_encoder_t *e);

AVPacket *
rtmpjanus_opus_encoder_encode(rtmpjanus_opus_encoder_t *e, AVFrame *frame);

#ifdef __cplusplus
}
#endif

#endif
