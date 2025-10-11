#ifndef TEST_GRAPHICS_H
#define TEST_GRAPHICS_H

#include <glib.h>

G_BEGIN_DECLS

// Test registration function
void register_graphics_tests(void);

// Renderer tests
void test_renderer_init(void);
void test_renderer_cleanup(void);
void test_renderer_context_creation(void);
void test_renderer_frame_operations(void);
void test_renderer_viewport_settings(void);
void test_renderer_clear_operations(void);

// Drawing tests
void test_renderer_draw_rectangle(void);
void test_renderer_draw_rounded_rectangle(void);
void test_renderer_draw_circle(void);
void test_renderer_draw_ellipse(void);
void test_renderer_draw_line(void);
void test_renderer_draw_polyline(void);
void test_renderer_draw_polygon(void);
void test_renderer_draw_texture(void);
void test_renderer_draw_text(void);

// Matrix operations tests
void test_renderer_matrix_operations(void);
void test_renderer_transformations(void);
void test_renderer_matrix_stack(void);

// Texture management tests
void test_renderer_texture_creation(void);
void test_renderer_texture_loading(void);
void test_renderer_texture_binding(void);
void test_renderer_texture_deletion(void);

// Shader management tests
void test_renderer_shader_creation(void);
void test_renderer_shader_loading(void);
void test_renderer_shader_compilation(void);
void test_renderer_shader_uniforms(void);
void test_renderer_shader_deletion(void);

// Animation tests
void test_renderer_animation_creation(void);
void test_renderer_animation_playback(void);
void test_renderer_animation_control(void);
void test_renderer_animation_easing(void);
void test_renderer_animation_callbacks(void);
void test_renderer_animation_cleanup(void);

// Batch rendering tests
void test_renderer_batch_operations(void);
void test_renderer_batch_performance(void);

// Framebuffer tests
void test_renderer_framebuffer_creation(void);
void test_renderer_framebuffer_binding(void);
void test_renderer_render_to_texture(void);

// Settings tests
void test_renderer_quality_settings(void);
void test_renderer_vsync_settings(void);
void test_renderer_fps_settings(void);
void test_renderer_multisampling(void);
void test_renderer_anisotropic_filtering(void);

// Statistics tests
void test_renderer_statistics(void);
void test_renderer_performance_monitoring(void);

// Error handling tests
void test_renderer_error_handling(void);
void test_renderer_opengl_errors(void);
void test_renderer_resource_limits(void);

// Performance tests
void test_renderer_draw_performance(void);
void test_renderer_texture_performance(void);
void test_renderer_shader_performance(void);
void test_renderer_animation_performance(void);

// Integration tests
void test_renderer_gtk_integration(void);
void test_renderer_window_resize(void);
void test_renderer_multi_context(void);

// Mock graphics context
typedef struct _MockGraphicsContext MockGraphicsContext;

MockGraphicsContext *mock_graphics_context_new(void);
void mock_graphics_context_free(MockGraphicsContext *context);
void mock_graphics_context_make_current(MockGraphicsContext *context);
void mock_graphics_context_set_opengl_version(MockGraphicsContext *context, int major, int minor);
void mock_graphics_context_set_extensions(MockGraphicsContext *context, const char **extensions);

// Test helpers
void test_graphics_setup_environment(void);
void test_graphics_cleanup_environment(void);
gboolean test_graphics_create_test_texture(const char *name, int width, int height);
void test_graphics_remove_test_texture(const char *name);
gboolean test_graphics_create_test_shader(const char *name, const char *vertex_src, const char *fragment_src);
void test_graphics_remove_test_shader(const char *name);

// Performance measurement
typedef struct {
    gdouble frame_time;
    gdouble draw_time;
    gdouble gpu_time;
    guint32 draw_calls;
    guint32 triangles;
    gsize memory_used;
} GraphicsPerformanceMetrics;

GraphicsPerformanceMetrics test_graphics_measure_performance(void (*test_func)(void));
void test_graphics_log_performance(const char *test_name, GraphicsPerformanceMetrics metrics);

// Rendering verification
gboolean test_graphics_verify_frame_rendered(void);
gboolean test_graphics_verify_texture_loaded(const char *name);
gboolean test_graphics_verify_shader_compiled(const char *name);
gboolean test_graphics_verify_animation_running(const char *name);

// Test data generators
void test_graphics_generate_test_vertices(float *vertices, int count);
void test_graphics_generate_test_texture_data(unsigned char *data, int width, int height, int channels);
const char *test_graphics_get_test_vertex_shader(void);
const char *test_graphics_get_test_fragment_shader(void);

G_END_DECLS

#endif // TEST_GRAPHICS_H