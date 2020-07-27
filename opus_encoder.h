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
} opus_encoder_t;

#ifdef __cplusplus
extern "C" {
#endif

int
opus_encoder_init(void);

opus_encoder_t *
opus_encoder_new(uint32_t channels);

void
opus_encoder_close(opus_encoder_t *e);

AVPacket *
opus_encoder_encode(opus_encoder_t *e, AVFrame *frame);

#ifdef __cplusplus
}
#endif

#endif
