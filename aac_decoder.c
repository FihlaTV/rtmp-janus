#include "aac_decoder.h"
#include <stddef.h>
#include <libavutil/mem.h>

#ifdef DEBUG
#include <stdio.h>
#define DEBUG_LOG(...) fprintf(...)
#else
#define DEBUG_LOG(...)
#endif

static AVCodec *codec = NULL;

int
aac_decoder_init(void) {
    codec = avcodec_find_decoder(AV_CODEC_ID_AAC);
    return codec == NULL;
}

aac_decoder_t *
aac_decoder_new(void) {
    aac_decoder_t *d = NULL;
    if(codec == NULL) return NULL;

    d = (aac_decoder_t *)av_mallocz(sizeof(aac_decoder_t));
    if(d == NULL) return NULL;

    d->f = av_frame_alloc();
    if(d->f == NULL) {
        av_free(d);
        return NULL;
    }

    d->ctx = avcodec_alloc_context3(codec);
    if(d->ctx == NULL) {
        av_frame_free(&d->f);
        av_free(d);
        return NULL;
    }

    return d;
}


int
aac_decoder_open(aac_decoder_t *d, uint8_t *header, size_t headerLen) {
    d->ctx->extradata = av_mallocz((headerLen + 15) & ~0x07);
    d->ctx->extradata_size = headerLen;
    memcpy(d->ctx->extradata,header,headerLen);

    return avcodec_open2(d->ctx,codec,NULL);
}

void
aac_decoder_close(aac_decoder_t *d) {
    avcodec_free_context(&d->ctx);
    av_frame_free(&d->f);
    av_free(d);
}

AVFrame *
aac_decoder_decode(aac_decoder_t *d, uint8_t *data, size_t len) {
    AVPacket packet;
    int got;
    int read;
    av_init_packet(&packet);

    packet.data = av_mallocz((len + 15) & ~0x07);
    if(packet.data == NULL) return NULL;
    memcpy(packet.data,data,len);

    packet.size = len;

    read = avcodec_decode_audio4(d->ctx,d->f,&got,&packet);

    av_free(packet.data);

    if(read < 0) return NULL;
    if(got) return d->f;
    return NULL;
}

