#ifndef TAISHANG_AUDIO_MANAGER_H
#define TAISHANG_AUDIO_MANAGER_H

#include <glib.h>
#include <gio/gio.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangAudioManager TaishangAudioManager;

// Audio format types
typedef enum {
    TAISHANG_AUDIO_FORMAT_S16LE,
    TAISHANG_AUDIO_FORMAT_S24LE,
    TAISHANG_AUDIO_FORMAT_S32LE,
    TAISHANG_AUDIO_FORMAT_FLOAT32,
    TAISHANG_AUDIO_FORMAT_FLOAT64
} TaishangAudioFormat;

// Notification sound types
typedef enum {
    TAISHANG_NOTIFICATION_MESSAGE,
    TAISHANG_NOTIFICATION_ALERT,
    TAISHANG_NOTIFICATION_ERROR,
    TAISHANG_NOTIFICATION_SUCCESS,
    TAISHANG_NOTIFICATION_CALL
} TaishangNotificationSound;

// Audio device information
typedef struct {
    gchar *name;
    gchar *description;
    gboolean is_default;
    gint channels;
    gint sample_rate;
    gboolean is_input;
    gboolean is_output;
} TaishangAudioDevice;

// Audio statistics
typedef struct {
    guint64 samples_processed;
    guint32 samples_loaded;
    guint32 sounds_played;
    guint32 streams_active;
    gdouble cpu_usage;
    gdouble latency;
    guint32 buffer_underruns;
    guint32 buffer_overruns;
} TaishangAudioStats;

// Callback function types
typedef void (*TaishangAudioStreamCallback)(const char *stream_name, 
                                            gdouble position, 
                                            gpointer user_data);

typedef void (*TaishangAudioDeviceCallback)(const char *device_name, 
                                            gboolean connected, 
                                            gpointer user_data);

// Initialization and cleanup
gboolean taishang_audio_manager_init(void);
void taishang_audio_manager_cleanup(void);
TaishangAudioManager *taishang_audio_manager_get_instance(void);

// Audio sample management
gboolean taishang_audio_manager_load_sample(const char *name, const char *filename);
void taishang_audio_manager_unload_sample(const char *name);
gboolean taishang_audio_manager_is_sample_loaded(const char *name);

// Sound effects
gboolean taishang_audio_manager_play_sound(const char *sample_name, gdouble volume);
gboolean taishang_audio_manager_play_notification(TaishangNotificationSound sound);
void taishang_audio_manager_stop_all_sounds(void);

// Audio streams
gboolean taishang_audio_manager_create_stream(const char *name, const char *sample_name);
gboolean taishang_audio_manager_play_stream(const char *name);
gboolean taishang_audio_manager_pause_stream(const char *name);
gboolean taishang_audio_manager_stop_stream(const char *name);
void taishang_audio_manager_remove_stream(const char *name);

gboolean taishang_audio_manager_is_stream_playing(const char *name);
gdouble taishang_audio_manager_get_stream_position(const char *name);
gdouble taishang_audio_manager_get_stream_duration(const char *name);

// Stream properties
void taishang_audio_manager_set_stream_volume(const char *name, gdouble volume);
gdouble taishang_audio_manager_get_stream_volume(const char *name);

void taishang_audio_manager_set_stream_loop(const char *name, gboolean loop);
gboolean taishang_audio_manager_get_stream_loop(const char *name);

void taishang_audio_manager_set_stream_speed(const char *name, gdouble speed);
gdouble taishang_audio_manager_get_stream_speed(const char *name);

void taishang_audio_manager_set_stream_position(const char *name, gdouble position);

void taishang_audio_manager_set_stream_callback(const char *name, 
                                                TaishangAudioStreamCallback callback, 
                                                gpointer user_data);

// Volume control
void taishang_audio_manager_set_master_volume(gdouble volume);
gdouble taishang_audio_manager_get_master_volume(void);

void taishang_audio_manager_set_notification_volume(gdouble volume);
gdouble taishang_audio_manager_get_notification_volume(void);

void taishang_audio_manager_set_voice_volume(gdouble volume);
gdouble taishang_audio_manager_get_voice_volume(void);

void taishang_audio_manager_set_muted(gboolean muted);
gboolean taishang_audio_manager_is_muted(void);

// Audio device management
GList *taishang_audio_manager_get_devices(void);
gboolean taishang_audio_manager_set_device(const char *device_name);
const char *taishang_audio_manager_get_current_device(void);

void taishang_audio_manager_set_device_callback(TaishangAudioDeviceCallback callback, 
                                                gpointer user_data);

// Audio settings
void taishang_audio_manager_set_sample_rate(gint sample_rate);
gint taishang_audio_manager_get_sample_rate(void);

void taishang_audio_manager_set_buffer_size(gint buffer_size);
gint taishang_audio_manager_get_buffer_size(void);

void taishang_audio_manager_set_channels(gint channels);
gint taishang_audio_manager_get_channels(void);

void taishang_audio_manager_set_format(TaishangAudioFormat format);
TaishangAudioFormat taishang_audio_manager_get_format(void);

// Voice functionality (placeholder for future implementation)
gboolean taishang_audio_manager_start_voice_recording(void);
gboolean taishang_audio_manager_stop_voice_recording(void);
gboolean taishang_audio_manager_is_voice_recording(void);

gboolean taishang_audio_manager_play_voice_message(const char *filename);
gboolean taishang_audio_manager_save_voice_message(const char *filename);

// Audio effects (placeholder for future implementation)
void taishang_audio_manager_set_echo_enabled(gboolean enabled);
gboolean taishang_audio_manager_get_echo_enabled(void);

void taishang_audio_manager_set_noise_reduction_enabled(gboolean enabled);
gboolean taishang_audio_manager_get_noise_reduction_enabled(void);

void taishang_audio_manager_set_equalizer_preset(const char *preset);
const char *taishang_audio_manager_get_equalizer_preset(void);

// Statistics and monitoring
TaishangAudioStats taishang_audio_manager_get_stats(void);
void taishang_audio_manager_reset_stats(void);

gdouble taishang_audio_manager_get_cpu_usage(void);
gdouble taishang_audio_manager_get_latency(void);

// Utility functions
gboolean taishang_audio_manager_is_backend_available(const char *backend_name);
const char *taishang_audio_manager_get_current_backend(void);

GList *taishang_audio_manager_get_supported_formats(void);
gboolean taishang_audio_manager_is_format_supported(TaishangAudioFormat format);

// Audio file utilities
gboolean taishang_audio_manager_get_file_info(const char *filename, 
                                              gint *channels, 
                                              gint *sample_rate, 
                                              gdouble *duration);

gboolean taishang_audio_manager_convert_file(const char *input_file, 
                                             const char *output_file, 
                                             TaishangAudioFormat format);

// Memory management utilities
void taishang_audio_device_free(TaishangAudioDevice *device);
void taishang_audio_device_list_free(GList *devices);

// Error handling
GQuark taishang_audio_manager_error_quark(void);
#define TAISHANG_AUDIO_MANAGER_ERROR taishang_audio_manager_error_quark()

typedef enum {
    TAISHANG_AUDIO_MANAGER_ERROR_INIT_FAILED,
    TAISHANG_AUDIO_MANAGER_ERROR_BACKEND_UNAVAILABLE,
    TAISHANG_AUDIO_MANAGER_ERROR_DEVICE_NOT_FOUND,
    TAISHANG_AUDIO_MANAGER_ERROR_SAMPLE_LOAD_FAILED,
    TAISHANG_AUDIO_MANAGER_ERROR_STREAM_CREATE_FAILED,
    TAISHANG_AUDIO_MANAGER_ERROR_PLAYBACK_FAILED,
    TAISHANG_AUDIO_MANAGER_ERROR_RECORDING_FAILED
} TaishangAudioManagerError;

// Convenience macros
#define TAISHANG_AUDIO_DEFAULT_SAMPLE_RATE 44100
#define TAISHANG_AUDIO_DEFAULT_CHANNELS 2
#define TAISHANG_AUDIO_DEFAULT_BUFFER_SIZE 1024
#define TAISHANG_AUDIO_DEFAULT_FORMAT TAISHANG_AUDIO_FORMAT_FLOAT32

#define TAISHANG_AUDIO_MIN_VOLUME 0.0
#define TAISHANG_AUDIO_MAX_VOLUME 1.0
#define TAISHANG_AUDIO_DEFAULT_VOLUME 1.0

#define TAISHANG_AUDIO_MIN_SPEED 0.1
#define TAISHANG_AUDIO_MAX_SPEED 4.0
#define TAISHANG_AUDIO_DEFAULT_SPEED 1.0

// Predefined notification sound files
#define TAISHANG_AUDIO_NOTIFICATION_MESSAGE_FILE "notification_message.ogg"
#define TAISHANG_AUDIO_NOTIFICATION_ALERT_FILE "notification_alert.ogg"
#define TAISHANG_AUDIO_NOTIFICATION_ERROR_FILE "notification_error.ogg"
#define TAISHANG_AUDIO_NOTIFICATION_SUCCESS_FILE "notification_success.ogg"
#define TAISHANG_AUDIO_NOTIFICATION_CALL_FILE "notification_call.ogg"

G_END_DECLS

#endif // TAISHANG_AUDIO_MANAGER_H