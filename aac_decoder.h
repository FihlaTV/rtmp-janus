#ifndef AAC_DECODER_H
#define AAC_DECODER_H

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <stdint.h>

typedef struct {
    AVCodecContext *ctx;
    AVFrame *f;
    int got;
} rtmpjanus_aac_decoder_t;

#ifdef __cplusplus
extern "C" {
#endif

int
rtmpjanus_aac_decoder_init(void); /* must be called very early in process */

rtmpjanus_aac_decoder_t *
rtmpjanus_aac_decoder_new(void);

void
rtmpjanus_aac_decoder_close(rtmpjanus_aac_decoder_t *d);

int
rtmpjanus_aac_decoder_open(rtmpjanus_aac_decoder_t *d, uint8_t *header, size_t headerlen);


AVFrame *
rtmpjanus_aac_decoder_decode(rtmpjanus_aac_decoder_t *d, uint8_t *data, size_t len);


#ifdef __cplusplus
}
#endif

#endif
