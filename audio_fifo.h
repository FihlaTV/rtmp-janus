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
} audio_fifo_t;

#ifdef __cplusplus
extern "C" {
#endif

audio_fifo_t *
audio_fifo_new(uint32_t out_samplerate, uint32_t out_channels);

int
audio_fifo_open(audio_fifo_t *f, uint32_t in_samplerate, uint32_t in_channels);

int
audio_fifo_load(audio_fifo_t *f, AVFrame *frame);

uint32_t
audio_fifo_size(audio_fifo_t *f);

AVFrame *
audio_fifo_read(audio_fifo_t *f, uint32_t samples);

AVFrame *
audio_fifo_flush(audio_fifo_t *f);

void
audio_fifo_close(audio_fifo_t *f);



#ifdef __cplusplus
}
#endif

#endif
