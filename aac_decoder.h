#ifndef AAC_DECODER_H
#define AAC_DECODER_H

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <stdint.h>

typedef struct {
    AVCodecContext *ctx;
    AVFrame *f;
    int got;
} aac_decoder_t;

#ifdef __cplusplus
extern "C" {
#endif

int
aac_decoder_init(void); /* must be called very early in process */

aac_decoder_t *
aac_decoder_new(void);

void
aac_decoder_close(aac_decoder_t *d);

int
aac_decoder_open(aac_decoder_t *d, uint8_t *header, size_t headerlen);


AVFrame *
aac_decoder_decode(aac_decoder_t *d, uint8_t *data, size_t len);


#ifdef __cplusplus
}
#endif

#endif
