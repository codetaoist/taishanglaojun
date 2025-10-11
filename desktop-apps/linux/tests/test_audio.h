#ifndef TEST_AUDIO_H
#define TEST_AUDIO_H

#include <glib.h>

G_BEGIN_DECLS

// Test registration function
void register_audio_tests(void);

// Audio manager tests
void test_audio_manager_init(void);
void test_audio_manager_cleanup(void);
void test_audio_manager_backend_detection(void);
void test_audio_manager_device_enumeration(void);

// Sample management tests
void test_audio_sample_loading(void);
void test_audio_sample_unloading(void);
void test_audio_sample_format_support(void);
void test_audio_sample_memory_management(void);

// Sound playback tests
void test_audio_sound_playback(void);
void test_audio_notification_sounds(void);
void test_audio_sound_volume_control(void);
void test_audio_sound_concurrent_playback(void);

// Stream management tests
void test_audio_stream_creation(void);
void test_audio_stream_playback_control(void);
void test_audio_stream_properties(void);
void test_audio_stream_callbacks(void);
void test_audio_stream_cleanup(void);

// Volume control tests
void test_audio_master_volume(void);
void test_audio_notification_volume(void);
void test_audio_voice_volume(void);
void test_audio_mute_functionality(void);

// Device management tests
void test_audio_device_selection(void);
void test_audio_device_switching(void);
void test_audio_device_callbacks(void);
void test_audio_device_hot_plugging(void);

// Audio settings tests
void test_audio_sample_rate_settings(void);
void test_audio_buffer_size_settings(void);
void test_audio_channel_settings(void);
void test_audio_format_settings(void);

// Voice functionality tests
void test_audio_voice_recording(void);
void test_audio_voice_playback(void);
void test_audio_voice_message_handling(void);

// Audio effects tests
void test_audio_echo_effects(void);
void test_audio_noise_reduction(void);
void test_audio_equalizer(void);

// Backend-specific tests
void test_audio_pulseaudio_backend(void);
void test_audio_alsa_backend(void);
void test_audio_backend_switching(void);

// Performance tests
void test_audio_latency_measurement(void);
void test_audio_cpu_usage(void);
void test_audio_memory_usage(void);
void test_audio_concurrent_streams(void);

// Error handling tests
void test_audio_error_handling(void);
void test_audio_device_unavailable(void);
void test_audio_format_unsupported(void);
void test_audio_buffer_underrun(void);

// Statistics tests
void test_audio_statistics_collection(void);
void test_audio_performance_monitoring(void);

// File format tests
void test_audio_file_format_support(void);
void test_audio_file_conversion(void);
void test_audio_file_info_extraction(void);

// Integration tests
void test_audio_system_integration(void);
void test_audio_notification_integration(void);
void test_audio_ui_integration(void);

// Mock audio backend
typedef struct _MockAudioBackend MockAudioBackend;

MockAudioBackend *mock_audio_backend_new(const char *name);
void mock_audio_backend_free(MockAudioBackend *backend);
void mock_audio_backend_set_available(MockAudioBackend *backend, gboolean available);
void mock_audio_backend_set_latency(MockAudioBackend *backend, gdouble latency);
void mock_audio_backend_set_sample_rate(MockAudioBackend *backend, gint sample_rate);
void mock_audio_backend_add_device(MockAudioBackend *backend, const char *device_name);
void mock_audio_backend_remove_device(MockAudioBackend *backend, const char *device_name);

// Test helpers
void test_audio_setup_environment(void);
void test_audio_cleanup_environment(void);
gboolean test_audio_create_test_sample(const char *name, gdouble duration, gint frequency);
void test_audio_remove_test_sample(const char *name);
gboolean test_audio_create_test_file(const char *filename, gdouble duration, gint frequency);
void test_audio_remove_test_file(const char *filename);

// Audio verification
gboolean test_audio_verify_sample_loaded(const char *name);
gboolean test_audio_verify_stream_playing(const char *name);
gboolean test_audio_verify_sound_played(void);
gboolean test_audio_verify_device_available(const char *device_name);

// Performance measurement
typedef struct {
    gdouble latency;
    gdouble cpu_usage;
    gsize memory_usage;
    guint32 buffer_underruns;
    guint32 buffer_overruns;
    guint64 samples_processed;
} AudioPerformanceMetrics;

AudioPerformanceMetrics test_audio_measure_performance(void (*test_func)(void));
void test_audio_log_performance(const char *test_name, AudioPerformanceMetrics metrics);

// Test data generators
void test_audio_generate_sine_wave(float *buffer, gsize frames, gint channels, gdouble frequency, gint sample_rate);
void test_audio_generate_white_noise(float *buffer, gsize frames, gint channels);
void test_audio_generate_silence(float *buffer, gsize frames, gint channels);

// Audio quality verification
gboolean test_audio_verify_no_clipping(const float *buffer, gsize frames, gint channels);
gboolean test_audio_verify_no_silence(const float *buffer, gsize frames, gint channels);
gdouble test_audio_calculate_rms(const float *buffer, gsize frames, gint channels);
gdouble test_audio_calculate_thd(const float *buffer, gsize frames, gint channels, gdouble frequency, gint sample_rate);

// Stress testing
void test_audio_stress_concurrent_playback(gint stream_count, gdouble duration);
void test_audio_stress_rapid_start_stop(gint iterations);
void test_audio_stress_memory_allocation(gint sample_count);

G_END_DECLS

#endif // TEST_AUDIO_H