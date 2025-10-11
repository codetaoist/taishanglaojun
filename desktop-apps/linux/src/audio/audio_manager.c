#include "audio/audio_manager.h"
#include <glib.h>
#include <gio/gio.h>
#include <pulse/pulseaudio.h>
#include <alsa/asoundlib.h>
#include <sndfile.h>
#include <pthread.h>
#include <math.h>

// Audio backend types
typedef enum {
    TAISHANG_AUDIO_BACKEND_NONE,
    TAISHANG_AUDIO_BACKEND_PULSEAUDIO,
    TAISHANG_AUDIO_BACKEND_ALSA
} TaishangAudioBackend;

// Audio sample structure
typedef struct {
    float *data;
    gsize frames;
    gint channels;
    gint sample_rate;
    gchar *name;
    gboolean loaded;
} TaishangAudioSample;

// Audio stream structure
typedef struct {
    gchar *name;
    TaishangAudioSample *sample;
    gboolean playing;
    gboolean looping;
    gdouble volume;
    gdouble position;
    gdouble speed;
    TaishangAudioStreamCallback callback;
    gpointer user_data;
} TaishangAudioStream;

// Main audio manager structure
struct _TaishangAudioManager {
    TaishangAudioBackend backend;
    gboolean initialized;
    
    // PulseAudio
    pa_threaded_mainloop *pa_mainloop;
    pa_context *pa_context;
    pa_stream *pa_stream;
    
    // ALSA
    snd_pcm_t *alsa_handle;
    snd_pcm_hw_params_t *alsa_params;
    
    // Audio settings
    gint sample_rate;
    gint channels;
    gint buffer_size;
    TaishangAudioFormat format;
    
    // Volume and effects
    gdouble master_volume;
    gdouble notification_volume;
    gdouble voice_volume;
    gboolean muted;
    
    // Audio samples and streams
    GHashTable *samples;
    GHashTable *streams;
    
    // Threading
    pthread_mutex_t mutex;
    pthread_t audio_thread;
    gboolean thread_running;
    
    // Statistics
    TaishangAudioStats stats;
    
    // Callbacks
    TaishangAudioDeviceCallback device_callback;
    gpointer device_callback_data;
};

// Global instance
static TaishangAudioManager *g_audio_manager = NULL;

// Forward declarations
static gboolean init_pulseaudio(TaishangAudioManager *manager);
static gboolean init_alsa(TaishangAudioManager *manager);
static void cleanup_pulseaudio(TaishangAudioManager *manager);
static void cleanup_alsa(TaishangAudioManager *manager);
static void *audio_thread_func(void *data);
static void process_audio_streams(TaishangAudioManager *manager);
static TaishangAudioSample *load_audio_file(const char *filename);
static void free_audio_sample(TaishangAudioSample *sample);
static void free_audio_stream(TaishangAudioStream *stream);

// PulseAudio callbacks
static void pa_context_state_callback(pa_context *context, void *userdata) {
    TaishangAudioManager *manager = (TaishangAudioManager *)userdata;
    
    switch (pa_context_get_state(context)) {
        case PA_CONTEXT_READY:
            g_debug("PulseAudio context ready");
            break;
        case PA_CONTEXT_FAILED:
            g_warning("PulseAudio context failed");
            break;
        case PA_CONTEXT_TERMINATED:
            g_debug("PulseAudio context terminated");
            break;
        default:
            break;
    }
    
    pa_threaded_mainloop_signal(manager->pa_mainloop, 0);
}

static void pa_stream_state_callback(pa_stream *stream, void *userdata) {
    TaishangAudioManager *manager = (TaishangAudioManager *)userdata;
    
    switch (pa_stream_get_state(stream)) {
        case PA_STREAM_READY:
            g_debug("PulseAudio stream ready");
            break;
        case PA_STREAM_FAILED:
            g_warning("PulseAudio stream failed");
            break;
        case PA_STREAM_TERMINATED:
            g_debug("PulseAudio stream terminated");
            break;
        default:
            break;
    }
    
    pa_threaded_mainloop_signal(manager->pa_mainloop, 0);
}

static void pa_stream_write_callback(pa_stream *stream, size_t nbytes, void *userdata) {
    TaishangAudioManager *manager = (TaishangAudioManager *)userdata;
    
    pthread_mutex_lock(&manager->mutex);
    
    void *buffer;
    pa_stream_begin_write(stream, &buffer, &nbytes);
    
    if (buffer) {
        // Fill buffer with audio data from active streams
        memset(buffer, 0, nbytes);
        process_audio_streams(manager);
        
        // Apply master volume
        if (manager->format == TAISHANG_AUDIO_FORMAT_FLOAT32) {
            float *samples = (float *)buffer;
            gsize sample_count = nbytes / sizeof(float);
            for (gsize i = 0; i < sample_count; i++) {
                samples[i] *= manager->master_volume;
            }
        }
        
        manager->stats.samples_processed += nbytes / (manager->channels * sizeof(float));
    }
    
    pa_stream_write(stream, buffer, nbytes, NULL, 0, PA_SEEK_RELATIVE);
    
    pthread_mutex_unlock(&manager->mutex);
}

// Initialize audio manager
gboolean taishang_audio_manager_init(void) {
    if (g_audio_manager) {
        return TRUE;
    }
    
    g_audio_manager = g_new0(TaishangAudioManager, 1);
    TaishangAudioManager *manager = g_audio_manager;
    
    // Initialize mutex
    pthread_mutex_init(&manager->mutex, NULL);
    
    // Set default settings
    manager->sample_rate = 44100;
    manager->channels = 2;
    manager->buffer_size = 1024;
    manager->format = TAISHANG_AUDIO_FORMAT_FLOAT32;
    manager->master_volume = 1.0;
    manager->notification_volume = 0.8;
    manager->voice_volume = 1.0;
    manager->muted = FALSE;
    
    // Initialize hash tables
    manager->samples = g_hash_table_new_full(g_str_hash, g_str_equal, 
                                             g_free, (GDestroyNotify)free_audio_sample);
    manager->streams = g_hash_table_new_full(g_str_hash, g_str_equal, 
                                             g_free, (GDestroyNotify)free_audio_stream);
    
    // Try to initialize PulseAudio first, then ALSA
    if (init_pulseaudio(manager)) {
        manager->backend = TAISHANG_AUDIO_BACKEND_PULSEAUDIO;
        g_info("Audio manager initialized with PulseAudio backend");
    } else if (init_alsa(manager)) {
        manager->backend = TAISHANG_AUDIO_BACKEND_ALSA;
        g_info("Audio manager initialized with ALSA backend");
    } else {
        g_warning("Failed to initialize any audio backend");
        taishang_audio_manager_cleanup();
        return FALSE;
    }
    
    // Start audio processing thread
    manager->thread_running = TRUE;
    pthread_create(&manager->audio_thread, NULL, audio_thread_func, manager);
    
    manager->initialized = TRUE;
    return TRUE;
}

// Cleanup audio manager
void taishang_audio_manager_cleanup(void) {
    if (!g_audio_manager) {
        return;
    }
    
    TaishangAudioManager *manager = g_audio_manager;
    
    // Stop audio thread
    if (manager->thread_running) {
        manager->thread_running = FALSE;
        pthread_join(manager->audio_thread, NULL);
    }
    
    // Cleanup backend
    switch (manager->backend) {
        case TAISHANG_AUDIO_BACKEND_PULSEAUDIO:
            cleanup_pulseaudio(manager);
            break;
        case TAISHANG_AUDIO_BACKEND_ALSA:
            cleanup_alsa(manager);
            break;
        default:
            break;
    }
    
    // Cleanup hash tables
    if (manager->samples) {
        g_hash_table_destroy(manager->samples);
    }
    if (manager->streams) {
        g_hash_table_destroy(manager->streams);
    }
    
    // Cleanup mutex
    pthread_mutex_destroy(&manager->mutex);
    
    g_free(manager);
    g_audio_manager = NULL;
}

// Get audio manager instance
TaishangAudioManager *taishang_audio_manager_get_instance(void) {
    return g_audio_manager;
}

// Load audio sample
gboolean taishang_audio_manager_load_sample(const char *name, const char *filename) {
    g_return_val_if_fail(g_audio_manager && name && filename, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Check if sample already exists
    if (g_hash_table_contains(manager->samples, name)) {
        pthread_mutex_unlock(&manager->mutex);
        return TRUE;
    }
    
    // Load audio file
    TaishangAudioSample *sample = load_audio_file(filename);
    if (!sample) {
        pthread_mutex_unlock(&manager->mutex);
        return FALSE;
    }
    
    sample->name = g_strdup(name);
    g_hash_table_insert(manager->samples, g_strdup(name), sample);
    
    manager->stats.samples_loaded++;
    
    pthread_mutex_unlock(&manager->mutex);
    
    g_debug("Loaded audio sample: %s", name);
    return TRUE;
}

// Unload audio sample
void taishang_audio_manager_unload_sample(const char *name) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    g_hash_table_remove(manager->samples, name);
    pthread_mutex_unlock(&manager->mutex);
    
    g_debug("Unloaded audio sample: %s", name);
}

// Play sound effect
gboolean taishang_audio_manager_play_sound(const char *sample_name, gdouble volume) {
    g_return_val_if_fail(g_audio_manager && sample_name, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioSample *sample = g_hash_table_lookup(manager->samples, sample_name);
    if (!sample) {
        pthread_mutex_unlock(&manager->mutex);
        g_warning("Audio sample not found: %s", sample_name);
        return FALSE;
    }
    
    // Create a new stream for this sound
    gchar *stream_name = g_strdup_printf("sound_%s_%ld", sample_name, time(NULL));
    
    TaishangAudioStream *stream = g_new0(TaishangAudioStream, 1);
    stream->name = g_strdup(stream_name);
    stream->sample = sample;
    stream->playing = TRUE;
    stream->looping = FALSE;
    stream->volume = volume;
    stream->position = 0.0;
    stream->speed = 1.0;
    
    g_hash_table_insert(manager->streams, g_strdup(stream_name), stream);
    
    manager->stats.sounds_played++;
    
    pthread_mutex_unlock(&manager->mutex);
    
    g_free(stream_name);
    return TRUE;
}

// Play notification sound
gboolean taishang_audio_manager_play_notification(TaishangNotificationSound sound) {
    g_return_val_if_fail(g_audio_manager, FALSE);
    
    const char *sample_name = NULL;
    
    switch (sound) {
        case TAISHANG_NOTIFICATION_MESSAGE:
            sample_name = "notification_message";
            break;
        case TAISHANG_NOTIFICATION_ALERT:
            sample_name = "notification_alert";
            break;
        case TAISHANG_NOTIFICATION_ERROR:
            sample_name = "notification_error";
            break;
        case TAISHANG_NOTIFICATION_SUCCESS:
            sample_name = "notification_success";
            break;
        case TAISHANG_NOTIFICATION_CALL:
            sample_name = "notification_call";
            break;
        default:
            return FALSE;
    }
    
    return taishang_audio_manager_play_sound(sample_name, g_audio_manager->notification_volume);
}

// Create audio stream
gboolean taishang_audio_manager_create_stream(const char *name, const char *sample_name) {
    g_return_val_if_fail(g_audio_manager && name && sample_name, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    // Check if stream already exists
    if (g_hash_table_contains(manager->streams, name)) {
        pthread_mutex_unlock(&manager->mutex);
        return TRUE;
    }
    
    TaishangAudioSample *sample = g_hash_table_lookup(manager->samples, sample_name);
    if (!sample) {
        pthread_mutex_unlock(&manager->mutex);
        g_warning("Audio sample not found: %s", sample_name);
        return FALSE;
    }
    
    TaishangAudioStream *stream = g_new0(TaishangAudioStream, 1);
    stream->name = g_strdup(name);
    stream->sample = sample;
    stream->playing = FALSE;
    stream->looping = FALSE;
    stream->volume = 1.0;
    stream->position = 0.0;
    stream->speed = 1.0;
    
    g_hash_table_insert(manager->streams, g_strdup(name), stream);
    
    pthread_mutex_unlock(&manager->mutex);
    
    g_debug("Created audio stream: %s", name);
    return TRUE;
}

// Control stream playback
gboolean taishang_audio_manager_play_stream(const char *name) {
    g_return_val_if_fail(g_audio_manager && name, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->playing = TRUE;
        stream->position = 0.0;
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return stream != NULL;
}

gboolean taishang_audio_manager_pause_stream(const char *name) {
    g_return_val_if_fail(g_audio_manager && name, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->playing = FALSE;
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return stream != NULL;
}

gboolean taishang_audio_manager_stop_stream(const char *name) {
    g_return_val_if_fail(g_audio_manager && name, FALSE);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->playing = FALSE;
        stream->position = 0.0;
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    return stream != NULL;
}

void taishang_audio_manager_remove_stream(const char *name) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    g_hash_table_remove(manager->streams, name);
    pthread_mutex_unlock(&manager->mutex);
    
    g_debug("Removed audio stream: %s", name);
}

// Volume control
void taishang_audio_manager_set_master_volume(gdouble volume) {
    g_return_if_fail(g_audio_manager);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    manager->master_volume = CLAMP(volume, 0.0, 1.0);
    pthread_mutex_unlock(&manager->mutex);
}

gdouble taishang_audio_manager_get_master_volume(void) {
    g_return_val_if_fail(g_audio_manager, 0.0);
    
    return g_audio_manager->master_volume;
}

void taishang_audio_manager_set_notification_volume(gdouble volume) {
    g_return_if_fail(g_audio_manager);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    manager->notification_volume = CLAMP(volume, 0.0, 1.0);
    pthread_mutex_unlock(&manager->mutex);
}

gdouble taishang_audio_manager_get_notification_volume(void) {
    g_return_val_if_fail(g_audio_manager, 0.0);
    
    return g_audio_manager->notification_volume;
}

void taishang_audio_manager_set_voice_volume(gdouble volume) {
    g_return_if_fail(g_audio_manager);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    manager->voice_volume = CLAMP(volume, 0.0, 1.0);
    pthread_mutex_unlock(&manager->mutex);
}

gdouble taishang_audio_manager_get_voice_volume(void) {
    g_return_val_if_fail(g_audio_manager, 0.0);
    
    return g_audio_manager->voice_volume;
}

void taishang_audio_manager_set_muted(gboolean muted) {
    g_return_if_fail(g_audio_manager);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    manager->muted = muted;
    pthread_mutex_unlock(&manager->mutex);
}

gboolean taishang_audio_manager_is_muted(void) {
    g_return_val_if_fail(g_audio_manager, FALSE);
    
    return g_audio_manager->muted;
}

// Stream properties
void taishang_audio_manager_set_stream_volume(const char *name, gdouble volume) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->volume = CLAMP(volume, 0.0, 1.0);
    }
    
    pthread_mutex_unlock(&manager->mutex);
}

void taishang_audio_manager_set_stream_loop(const char *name, gboolean loop) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->looping = loop;
    }
    
    pthread_mutex_unlock(&manager->mutex);
}

void taishang_audio_manager_set_stream_speed(const char *name, gdouble speed) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream) {
        stream->speed = CLAMP(speed, 0.1, 4.0);
    }
    
    pthread_mutex_unlock(&manager->mutex);
}

void taishang_audio_manager_set_stream_position(const char *name, gdouble position) {
    g_return_if_fail(g_audio_manager && name);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    
    TaishangAudioStream *stream = g_hash_table_lookup(manager->streams, name);
    if (stream && stream->sample) {
        stream->position = CLAMP(position, 0.0, (gdouble)stream->sample->frames);
    }
    
    pthread_mutex_unlock(&manager->mutex);
}

// Audio device management
GList *taishang_audio_manager_get_devices(void) {
    g_return_val_if_fail(g_audio_manager, NULL);
    
    // Implementation depends on backend
    // For now, return a simple list
    GList *devices = NULL;
    
    TaishangAudioDevice *default_device = g_new0(TaishangAudioDevice, 1);
    default_device->name = g_strdup("Default");
    default_device->description = g_strdup("Default Audio Device");
    default_device->is_default = TRUE;
    default_device->channels = 2;
    default_device->sample_rate = 44100;
    
    devices = g_list_append(devices, default_device);
    
    return devices;
}

gboolean taishang_audio_manager_set_device(const char *device_name) {
    g_return_val_if_fail(g_audio_manager && device_name, FALSE);
    
    // Implementation depends on backend
    g_debug("Setting audio device: %s", device_name);
    return TRUE;
}

// Statistics
TaishangAudioStats taishang_audio_manager_get_stats(void) {
    g_return_val_if_fail(g_audio_manager, (TaishangAudioStats){0});
    
    return g_audio_manager->stats;
}

void taishang_audio_manager_reset_stats(void) {
    g_return_if_fail(g_audio_manager);
    
    TaishangAudioManager *manager = g_audio_manager;
    
    pthread_mutex_lock(&manager->mutex);
    memset(&manager->stats, 0, sizeof(TaishangAudioStats));
    pthread_mutex_unlock(&manager->mutex);
}

// Private helper functions
static gboolean init_pulseaudio(TaishangAudioManager *manager) {
    // Create mainloop
    manager->pa_mainloop = pa_threaded_mainloop_new();
    if (!manager->pa_mainloop) {
        return FALSE;
    }
    
    // Create context
    pa_mainloop_api *api = pa_threaded_mainloop_get_api(manager->pa_mainloop);
    manager->pa_context = pa_context_new(api, "Taishang Audio Manager");
    if (!manager->pa_context) {
        pa_threaded_mainloop_free(manager->pa_mainloop);
        return FALSE;
    }
    
    // Set callbacks
    pa_context_set_state_callback(manager->pa_context, pa_context_state_callback, manager);
    
    // Start mainloop
    pa_threaded_mainloop_start(manager->pa_mainloop);
    
    // Connect to server
    pa_threaded_mainloop_lock(manager->pa_mainloop);
    
    if (pa_context_connect(manager->pa_context, NULL, PA_CONTEXT_NOFLAGS, NULL) < 0) {
        pa_threaded_mainloop_unlock(manager->pa_mainloop);
        cleanup_pulseaudio(manager);
        return FALSE;
    }
    
    // Wait for connection
    while (pa_context_get_state(manager->pa_context) != PA_CONTEXT_READY) {
        if (pa_context_get_state(manager->pa_context) == PA_CONTEXT_FAILED ||
            pa_context_get_state(manager->pa_context) == PA_CONTEXT_TERMINATED) {
            pa_threaded_mainloop_unlock(manager->pa_mainloop);
            cleanup_pulseaudio(manager);
            return FALSE;
        }
        pa_threaded_mainloop_wait(manager->pa_mainloop);
    }
    
    // Create stream
    pa_sample_spec spec = {
        .format = PA_SAMPLE_FLOAT32LE,
        .rate = manager->sample_rate,
        .channels = manager->channels
    };
    
    manager->pa_stream = pa_stream_new(manager->pa_context, "Taishang Audio Stream", &spec, NULL);
    if (!manager->pa_stream) {
        pa_threaded_mainloop_unlock(manager->pa_mainloop);
        cleanup_pulseaudio(manager);
        return FALSE;
    }
    
    pa_stream_set_state_callback(manager->pa_stream, pa_stream_state_callback, manager);
    pa_stream_set_write_callback(manager->pa_stream, pa_stream_write_callback, manager);
    
    // Connect stream
    pa_buffer_attr buffer_attr = {
        .maxlength = manager->buffer_size * manager->channels * sizeof(float) * 4,
        .tlength = manager->buffer_size * manager->channels * sizeof(float),
        .prebuf = 0,
        .minreq = manager->buffer_size * manager->channels * sizeof(float) / 4,
        .fragsize = -1
    };
    
    if (pa_stream_connect_playback(manager->pa_stream, NULL, &buffer_attr,
                                   PA_STREAM_ADJUST_LATENCY, NULL, NULL) < 0) {
        pa_threaded_mainloop_unlock(manager->pa_mainloop);
        cleanup_pulseaudio(manager);
        return FALSE;
    }
    
    // Wait for stream to be ready
    while (pa_stream_get_state(manager->pa_stream) != PA_STREAM_READY) {
        if (pa_stream_get_state(manager->pa_stream) == PA_STREAM_FAILED ||
            pa_stream_get_state(manager->pa_stream) == PA_STREAM_TERMINATED) {
            pa_threaded_mainloop_unlock(manager->pa_mainloop);
            cleanup_pulseaudio(manager);
            return FALSE;
        }
        pa_threaded_mainloop_wait(manager->pa_mainloop);
    }
    
    pa_threaded_mainloop_unlock(manager->pa_mainloop);
    
    return TRUE;
}

static gboolean init_alsa(TaishangAudioManager *manager) {
    int err;
    
    // Open PCM device
    err = snd_pcm_open(&manager->alsa_handle, "default", SND_PCM_STREAM_PLAYBACK, 0);
    if (err < 0) {
        g_warning("Cannot open ALSA device: %s", snd_strerror(err));
        return FALSE;
    }
    
    // Allocate hardware parameters
    snd_pcm_hw_params_alloca(&manager->alsa_params);
    
    // Initialize parameters
    err = snd_pcm_hw_params_any(manager->alsa_handle, manager->alsa_params);
    if (err < 0) {
        g_warning("Cannot initialize ALSA parameters: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Set access type
    err = snd_pcm_hw_params_set_access(manager->alsa_handle, manager->alsa_params, 
                                       SND_PCM_ACCESS_RW_INTERLEAVED);
    if (err < 0) {
        g_warning("Cannot set ALSA access type: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Set sample format
    err = snd_pcm_hw_params_set_format(manager->alsa_handle, manager->alsa_params, 
                                       SND_PCM_FORMAT_FLOAT_LE);
    if (err < 0) {
        g_warning("Cannot set ALSA sample format: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Set sample rate
    unsigned int rate = manager->sample_rate;
    err = snd_pcm_hw_params_set_rate_near(manager->alsa_handle, manager->alsa_params, 
                                          &rate, 0);
    if (err < 0) {
        g_warning("Cannot set ALSA sample rate: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Set channels
    err = snd_pcm_hw_params_set_channels(manager->alsa_handle, manager->alsa_params, 
                                         manager->channels);
    if (err < 0) {
        g_warning("Cannot set ALSA channels: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Set buffer size
    snd_pcm_uframes_t buffer_size = manager->buffer_size * 4;
    err = snd_pcm_hw_params_set_buffer_size_near(manager->alsa_handle, manager->alsa_params, 
                                                 &buffer_size);
    if (err < 0) {
        g_warning("Cannot set ALSA buffer size: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Apply parameters
    err = snd_pcm_hw_params(manager->alsa_handle, manager->alsa_params);
    if (err < 0) {
        g_warning("Cannot apply ALSA parameters: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    // Prepare device
    err = snd_pcm_prepare(manager->alsa_handle);
    if (err < 0) {
        g_warning("Cannot prepare ALSA device: %s", snd_strerror(err));
        cleanup_alsa(manager);
        return FALSE;
    }
    
    return TRUE;
}

static void cleanup_pulseaudio(TaishangAudioManager *manager) {
    if (manager->pa_stream) {
        pa_stream_unref(manager->pa_stream);
        manager->pa_stream = NULL;
    }
    
    if (manager->pa_context) {
        pa_context_disconnect(manager->pa_context);
        pa_context_unref(manager->pa_context);
        manager->pa_context = NULL;
    }
    
    if (manager->pa_mainloop) {
        pa_threaded_mainloop_stop(manager->pa_mainloop);
        pa_threaded_mainloop_free(manager->pa_mainloop);
        manager->pa_mainloop = NULL;
    }
}

static void cleanup_alsa(TaishangAudioManager *manager) {
    if (manager->alsa_handle) {
        snd_pcm_close(manager->alsa_handle);
        manager->alsa_handle = NULL;
    }
}

static void *audio_thread_func(void *data) {
    TaishangAudioManager *manager = (TaishangAudioManager *)data;
    
    while (manager->thread_running) {
        // Process audio streams for ALSA backend
        if (manager->backend == TAISHANG_AUDIO_BACKEND_ALSA) {
            pthread_mutex_lock(&manager->mutex);
            
            float buffer[manager->buffer_size * manager->channels];
            memset(buffer, 0, sizeof(buffer));
            
            process_audio_streams(manager);
            
            // Apply master volume
            for (int i = 0; i < manager->buffer_size * manager->channels; i++) {
                buffer[i] *= manager->master_volume;
            }
            
            pthread_mutex_unlock(&manager->mutex);
            
            // Write to ALSA
            if (manager->alsa_handle) {
                snd_pcm_writei(manager->alsa_handle, buffer, manager->buffer_size);
            }
        }
        
        // Sleep for a short time
        usleep(10000); // 10ms
    }
    
    return NULL;
}

static void process_audio_streams(TaishangAudioManager *manager) {
    // This function would mix all active audio streams
    // For now, it's a placeholder
    GHashTableIter iter;
    gpointer key, value;
    
    g_hash_table_iter_init(&iter, manager->streams);
    while (g_hash_table_iter_next(&iter, &key, &value)) {
        TaishangAudioStream *stream = (TaishangAudioStream *)value;
        
        if (stream->playing && stream->sample && stream->sample->loaded) {
            // Process stream audio data
            // This would involve mixing, resampling, effects, etc.
            
            // Update position
            stream->position += manager->buffer_size * stream->speed;
            
            // Check if stream finished
            if (stream->position >= stream->sample->frames) {
                if (stream->looping) {
                    stream->position = 0.0;
                } else {
                    stream->playing = FALSE;
                    // Remove one-shot sound effects
                    if (g_str_has_prefix(stream->name, "sound_")) {
                        g_hash_table_iter_remove(&iter);
                    }
                }
            }
            
            // Call callback if set
            if (stream->callback) {
                stream->callback(stream->name, stream->position, stream->user_data);
            }
        }
    }
}

static TaishangAudioSample *load_audio_file(const char *filename) {
    SF_INFO info;
    SNDFILE *file = sf_open(filename, SFM_READ, &info);
    
    if (!file) {
        g_warning("Failed to open audio file: %s", filename);
        return NULL;
    }
    
    TaishangAudioSample *sample = g_new0(TaishangAudioSample, 1);
    sample->frames = info.frames;
    sample->channels = info.channels;
    sample->sample_rate = info.samplerate;
    sample->data = g_new(float, info.frames * info.channels);
    
    sf_count_t read_frames = sf_readf_float(file, sample->data, info.frames);
    if (read_frames != info.frames) {
        g_warning("Failed to read complete audio file: %s", filename);
        free_audio_sample(sample);
        sf_close(file);
        return NULL;
    }
    
    sf_close(file);
    sample->loaded = TRUE;
    
    return sample;
}

static void free_audio_sample(TaishangAudioSample *sample) {
    if (sample) {
        g_free(sample->data);
        g_free(sample->name);
        g_free(sample);
    }
}

static void free_audio_stream(TaishangAudioStream *stream) {
    if (stream) {
        g_free(stream->name);
        g_free(stream);
    }
}

// Error handling
GQuark taishang_audio_manager_error_quark(void) {
    return g_quark_from_static_string("taishang-audio-manager-error-quark");
}