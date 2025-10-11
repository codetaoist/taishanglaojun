#ifndef TAISHANG_RENDERER_H
#define TAISHANG_RENDERER_H

#include <glib.h>
#include <gtk/gtk.h>
#include <epoxy/gl.h>

G_BEGIN_DECLS

// Forward declarations
typedef struct _TaishangRenderer TaishangRenderer;
typedef struct _TaishangAnimation TaishangAnimation;

// Rendering quality levels
typedef enum {
    TAISHANG_RENDER_QUALITY_LOW,
    TAISHANG_RENDER_QUALITY_MEDIUM,
    TAISHANG_RENDER_QUALITY_HIGH,
    TAISHANG_RENDER_QUALITY_ULTRA
} TaishangRenderingQuality;

// Animation types
typedef enum {
    TAISHANG_ANIMATION_FADE,
    TAISHANG_ANIMATION_SLIDE,
    TAISHANG_ANIMATION_SCALE,
    TAISHANG_ANIMATION_ROTATE,
    TAISHANG_ANIMATION_CUSTOM
} TaishangAnimationType;

// Animation easing functions
typedef enum {
    TAISHANG_EASING_LINEAR,
    TAISHANG_EASING_EASE_IN,
    TAISHANG_EASING_EASE_OUT,
    TAISHANG_EASING_EASE_IN_OUT,
    TAISHANG_EASING_BOUNCE,
    TAISHANG_EASING_ELASTIC,
    TAISHANG_EASING_BACK,
    TAISHANG_EASING_CUBIC_BEZIER
} TaishangAnimationEasing;

// Rendering statistics
typedef struct {
    guint64 frame_count;
    gdouble fps;
    guint32 draw_calls;
    guint32 triangles_rendered;
    guint32 textures_bound;
    gsize memory_used;
    gdouble frame_time;
    gdouble cpu_time;
    gdouble gpu_time;
} TaishangRenderStats;

// Vertex structure
typedef struct {
    float x, y, z;          // Position
    float u, v;             // Texture coordinates
    float r, g, b, a;       // Color
} TaishangVertex;

// Texture information
typedef struct {
    GLuint id;
    gint width;
    gint height;
    gint channels;
    GLenum format;
    GLenum type;
    gchar *name;
} TaishangTexture;

// Shader information
typedef struct {
    GLuint program;
    GLuint vertex_shader;
    GLuint fragment_shader;
    GHashTable *uniforms;
    gchar *name;
} TaishangShader;

// Callback function types
typedef void (*TaishangAnimationCallback)(const char *name, 
                                          gdouble value, 
                                          gdouble progress, 
                                          gpointer user_data);

typedef void (*TaishangRenderCallback)(gpointer user_data);

// Initialization and cleanup
gboolean taishang_renderer_init(void);
void taishang_renderer_cleanup(void);
TaishangRenderer *taishang_renderer_get_instance(void);

// Context management
gboolean taishang_renderer_create_context(GtkWidget *widget);
gboolean taishang_renderer_make_current(void);
void taishang_renderer_swap_buffers(void);

// Rendering functions
gboolean taishang_renderer_begin_frame(void);
gboolean taishang_renderer_end_frame(void);
void taishang_renderer_clear(float r, float g, float b, float a);
void taishang_renderer_set_viewport(gint x, gint y, gint width, gint height);

// Drawing functions
void taishang_renderer_draw_rectangle(float x, float y, float width, float height, 
                                      float r, float g, float b, float a);

void taishang_renderer_draw_rounded_rectangle(float x, float y, float width, float height, 
                                              float radius, float r, float g, float b, float a);

void taishang_renderer_draw_circle(float x, float y, float radius, 
                                   float r, float g, float b, float a);

void taishang_renderer_draw_ellipse(float x, float y, float width, float height, 
                                    float r, float g, float b, float a);

void taishang_renderer_draw_line(float x1, float y1, float x2, float y2, 
                                 float width, float r, float g, float b, float a);

void taishang_renderer_draw_polyline(const float *points, gint count, 
                                     float width, float r, float g, float b, float a);

void taishang_renderer_draw_polygon(const float *points, gint count, 
                                    float r, float g, float b, float a);

void taishang_renderer_draw_texture(GLuint texture, float x, float y, 
                                    float width, float height, float opacity);

void taishang_renderer_draw_texture_region(GLuint texture, 
                                           float src_x, float src_y, float src_width, float src_height,
                                           float dst_x, float dst_y, float dst_width, float dst_height,
                                           float opacity);

void taishang_renderer_draw_text(const char *text, float x, float y, 
                                 float size, float r, float g, float b, float a);

// Matrix functions
void taishang_renderer_set_projection_matrix(const float *matrix);
void taishang_renderer_set_view_matrix(const float *matrix);
void taishang_renderer_set_model_matrix(const float *matrix);
void taishang_renderer_push_matrix(void);
void taishang_renderer_pop_matrix(void);
void taishang_renderer_translate(float x, float y, float z);
void taishang_renderer_rotate(float angle, float x, float y, float z);
void taishang_renderer_scale(float x, float y, float z);

// Texture management
TaishangTexture *taishang_renderer_create_texture(const char *name, 
                                                  gint width, gint height, 
                                                  gint channels, 
                                                  const void *data);

TaishangTexture *taishang_renderer_load_texture(const char *filename);
void taishang_renderer_bind_texture(TaishangTexture *texture);
void taishang_renderer_unbind_texture(void);
void taishang_renderer_delete_texture(TaishangTexture *texture);

// Shader management
TaishangShader *taishang_renderer_create_shader(const char *name,
                                                const char *vertex_source,
                                                const char *fragment_source);

TaishangShader *taishang_renderer_load_shader(const char *name,
                                              const char *vertex_file,
                                              const char *fragment_file);

void taishang_renderer_use_shader(TaishangShader *shader);
void taishang_renderer_set_shader_uniform_float(TaishangShader *shader, 
                                                const char *name, float value);
void taishang_renderer_set_shader_uniform_vec2(TaishangShader *shader, 
                                               const char *name, float x, float y);
void taishang_renderer_set_shader_uniform_vec3(TaishangShader *shader, 
                                               const char *name, float x, float y, float z);
void taishang_renderer_set_shader_uniform_vec4(TaishangShader *shader, 
                                               const char *name, float x, float y, float z, float w);
void taishang_renderer_set_shader_uniform_matrix4(TaishangShader *shader, 
                                                  const char *name, const float *matrix);
void taishang_renderer_delete_shader(TaishangShader *shader);

// Animation functions
TaishangAnimation *taishang_renderer_create_animation(const char *name, 
                                                      TaishangAnimationType type,
                                                      gdouble duration, 
                                                      gdouble start_value, 
                                                      gdouble end_value);

gboolean taishang_renderer_start_animation(const char *name);
gboolean taishang_renderer_stop_animation(const char *name);
gboolean taishang_renderer_pause_animation(const char *name);
gboolean taishang_renderer_resume_animation(const char *name);
gboolean taishang_renderer_remove_animation(const char *name);

gdouble taishang_renderer_get_animation_value(const char *name);
gdouble taishang_renderer_get_animation_progress(const char *name);
gboolean taishang_renderer_is_animation_active(const char *name);

void taishang_renderer_set_animation_easing(const char *name, TaishangAnimationEasing easing);
void taishang_renderer_set_animation_loop(const char *name, gboolean loop);
void taishang_renderer_set_animation_reverse(const char *name, gboolean reverse);
void taishang_renderer_set_animation_callback(const char *name, 
                                              TaishangAnimationCallback callback, 
                                              void *user_data);

// Batch rendering
void taishang_renderer_begin_batch(void);
void taishang_renderer_end_batch(void);
void taishang_renderer_flush_batch(void);

// Render targets
GLuint taishang_renderer_create_framebuffer(gint width, gint height);
void taishang_renderer_bind_framebuffer(GLuint framebuffer);
void taishang_renderer_unbind_framebuffer(void);
void taishang_renderer_delete_framebuffer(GLuint framebuffer);

// Settings functions
void taishang_renderer_set_quality(TaishangRenderingQuality quality);
TaishangRenderingQuality taishang_renderer_get_quality(void);

void taishang_renderer_set_vsync(gboolean enabled);
gboolean taishang_renderer_get_vsync(void);

void taishang_renderer_set_max_fps(gint fps);
gint taishang_renderer_get_max_fps(void);

void taishang_renderer_set_multisampling(gint samples);
gint taishang_renderer_get_multisampling(void);

void taishang_renderer_set_anisotropic_filtering(gint level);
gint taishang_renderer_get_anisotropic_filtering(void);

// Statistics functions
TaishangRenderStats taishang_renderer_get_stats(void);
void taishang_renderer_reset_stats(void);

// Utility functions
gboolean taishang_renderer_is_opengl_available(void);
const char *taishang_renderer_get_opengl_version(void);
const char *taishang_renderer_get_opengl_vendor(void);
const char *taishang_renderer_get_opengl_renderer(void);

gboolean taishang_renderer_check_extension(const char *extension);
void taishang_renderer_print_info(void);

// Error handling
GQuark taishang_renderer_error_quark(void);
#define TAISHANG_RENDERER_ERROR taishang_renderer_error_quark()

typedef enum {
    TAISHANG_RENDERER_ERROR_INIT_FAILED,
    TAISHANG_RENDERER_ERROR_CONTEXT_FAILED,
    TAISHANG_RENDERER_ERROR_SHADER_COMPILE_FAILED,
    TAISHANG_RENDERER_ERROR_TEXTURE_LOAD_FAILED,
    TAISHANG_RENDERER_ERROR_FRAMEBUFFER_FAILED
} TaishangRendererError;

// Convenience macros
#define TAISHANG_COLOR_WHITE 1.0f, 1.0f, 1.0f, 1.0f
#define TAISHANG_COLOR_BLACK 0.0f, 0.0f, 0.0f, 1.0f
#define TAISHANG_COLOR_RED 1.0f, 0.0f, 0.0f, 1.0f
#define TAISHANG_COLOR_GREEN 0.0f, 1.0f, 0.0f, 1.0f
#define TAISHANG_COLOR_BLUE 0.0f, 0.0f, 1.0f, 1.0f
#define TAISHANG_COLOR_TRANSPARENT 0.0f, 0.0f, 0.0f, 0.0f

G_END_DECLS

#endif // TAISHANG_RENDERER_H