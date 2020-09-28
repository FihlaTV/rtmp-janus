#ifndef AUDIO_FIFO_H
#define AUDIO_FIFO_H

#include <libavutil/audio_fifo.h>
#include <libswresample/swresample.h>
#include <stdint.h>

/* handles receiving decoded samples,
 * resample,
 * put into a fifo */

typedef struct {
    SwrContext *resampler;
    AVAudioFifo *fifo;
    AVFrame *frame;
    uint8_t *buffer;
    uint32_t bufferSize;
    uint32_t out_samplerate;
    uint32_t out_channels;
} rtmpjanus_audio_fifo_t;

#ifdef __cplusplus
extern "C" {
#endif

rtmpjanus_audio_fifo_t *
rtmpjanus_audio_fifo_new(uint32_t out_samplerate, uint32_t out_channels);

int
rtmpjanus_audio_fifo_open(rtmpjanus_audio_fifo_t *f, uint32_t in_samplerate, uint32_t in_channels);

int
rtmpjanus_audio_fifo_load(rtmpjanus_audio_fifo_t *f, AVFrame *frame);

uint32_t
rtmpjanus_audio_fifo_size(rtmpjanus_audio_fifo_t *f);

AVFrame *
rtmpjanus_audio_fifo_read(rtmpjanus_audio_fifo_t *f, uint32_t samples);

AVFrame *
rtmpjanus_audio_fifo_flush(rtmpjanus_audio_fifo_t *f);

void
rtmpjanus_audio_fifo_close(rtmpjanus_audio_fifo_t *f);



#ifdef __cplusplus
}
#endif

#endif
