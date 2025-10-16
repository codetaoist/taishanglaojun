#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <glib.h>
#include <gtk/gtk.h>
#include <gdk/gdk.h>
#include <epoxy/gl.h>
#include <cairo.h>
#include "../../include/graphics/renderer.h"

// Renderer structure
typedef struct {
    gboolean initialized;
    gboolean opengl_enabled;
    gboolean hardware_acceleration;
    
    // OpenGL context
    GdkGLContext *gl_context;
    GtkWidget *gl_area;
    
    // Rendering settings
    TaishangRenderingQuality quality;
    gboolean vsync_enabled;
    gint max_fps;
    
    // Animation system
    GHashTable *animations;
    guint animation_timer_id;
    gdouble current_time;
    
    // Shaders
    GLuint vertex_shader;
    GLuint fragment_shader;
    GLuint shader_program;
    
    // Buffers
    GLuint vertex_buffer;
    GLuint index_buffer;
    GLuint texture_buffer;
    
    // Matrices
    float projection_matrix[16];
    float view_matrix[16];
    float model_matrix[16];
    
    // Statistics
    TaishangRenderStats stats;
    
    GMutex mutex;
} TaishangRenderer;

// Animation structure
typedef struct {
    gchar *name;
    TaishangAnimationType type;
    TaishangAnimationEasing easing;
    gdouble duration;
    gdouble start_time;
    gdouble start_value;
    gdouble end_value;
    gdouble current_value;
    gboolean loop;
    gboolean reverse;
    TaishangAnimationCallback callback;
    void *user_data;
    gboolean active;
} TaishangAnimation;

static TaishangRenderer *renderer = NULL;

// Forward declarations
static gboolean init_opengl(void);
static void cleanup_opengl(void);
static gboolean compile_shaders(void);
static void setup_buffers(void);
static void cleanup_buffers(void);
static gboolean animation_timer_callback(gpointer user_data);
static void update_animations(gdouble current_time);
static gdouble apply_easing(gdouble t, TaishangAnimationEasing easing);
static void matrix_identity(float *matrix);
static void matrix_multiply(float *result, const float *a, const float *b);
static void matrix_translate(float *matrix, float x, float y, float z);
static void matrix_rotate(float *matrix, float angle, float x, float y, float z);
static void matrix_scale(float *matrix, float x, float y, float z);

// Vertex shader source
static const char *vertex_shader_source = 
    "#version 330 core\n"
    "layout (location = 0) in vec3 aPos;\n"
    "layout (location = 1) in vec2 aTexCoord;\n"
    "layout (location = 2) in vec4 aColor;\n"
    "uniform mat4 uProjection;\n"
    "uniform mat4 uView;\n"
    "uniform mat4 uModel;\n"
    "out vec2 TexCoord;\n"
    "out vec4 Color;\n"
    "void main() {\n"
    "    gl_Position = uProjection * uView * uModel * vec4(aPos, 1.0);\n"
    "    TexCoord = aTexCoord;\n"
    "    Color = aColor;\n"
    "}\n";

// Fragment shader source
static const char *fragment_shader_source = 
    "#version 330 core\n"
    "in vec2 TexCoord;\n"
    "in vec4 Color;\n"
    "out vec4 FragColor;\n"
    "uniform sampler2D uTexture;\n"
    "uniform bool uUseTexture;\n"
    "uniform float uOpacity;\n"
    "void main() {\n"
    "    if (uUseTexture) {\n"
    "        FragColor = texture(uTexture, TexCoord) * Color * uOpacity;\n"
    "    } else {\n"
    "        FragColor = Color * uOpacity;\n"
    "    }\n"
    "}\n";

// Public functions
gboolean taishang_renderer_init(void);
void taishang_renderer_cleanup(void);
TaishangRenderer *taishang_renderer_get_instance(void);

// Rendering functions
gboolean taishang_renderer_begin_frame(void);
gboolean taishang_renderer_end_frame(void);
void taishang_renderer_clear(float r, float g, float b, float a);
void taishang_renderer_set_viewport(gint x, gint y, gint width, gint height);

// Drawing functions
void taishang_renderer_draw_rectangle(float x, float y, float width, float height, 
                                      float r, float g, float b, float a);
void taishang_renderer_draw_circle(float x, float y, float radius, 
                                   float r, float g, float b, float a);
void taishang_renderer_draw_line(float x1, float y1, float x2, float y2, 
                                 float width, float r, float g, float b, float a);
void taishang_renderer_draw_texture(GLuint texture, float x, float y, 
                                    float width, float height, float opacity);

// Matrix functions
void taishang_renderer_set_projection_matrix(const float *matrix);
void taishang_renderer_set_view_matrix(const float *matrix);
void taishang_renderer_set_model_matrix(const float *matrix);
void taishang_renderer_push_matrix(void);
void taishang_renderer_pop_matrix(void);

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
gdouble taishang_renderer_get_animation_value(const char *name);
void taishang_renderer_set_animation_callback(const char *name, 
                                              TaishangAnimationCallback callback, 
                                              void *user_data);

// Settings functions
void taishang_renderer_set_quality(TaishangRenderingQuality quality);
TaishangRenderingQuality taishang_renderer_get_quality(void);
void taishang_renderer_set_vsync(gboolean enabled);
gboolean taishang_renderer_get_vsync(void);
void taishang_renderer_set_max_fps(gint fps);
gint taishang_renderer_get_max_fps(void);

// Statistics functions
TaishangRenderStats taishang_renderer_get_stats(void);
void taishang_renderer_reset_stats(void);

// Implementation
gboolean taishang_renderer_init(void) {
    if (renderer != NULL) {
        g_warning("Renderer already initialized");
        return FALSE;
    }
    
    renderer = g_new0(TaishangRenderer, 1);
    g_mutex_init(&renderer->mutex);
    
    // Initialize settings
    renderer->quality = TAISHANG_RENDER_QUALITY_HIGH;
    renderer->vsync_enabled = TRUE;
    renderer->max_fps = 60;
    
    // Initialize matrices
    matrix_identity(renderer->projection_matrix);
    matrix_identity(renderer->view_matrix);
    matrix_identity(renderer->model_matrix);
    
    // Initialize animations
    renderer->animations = g_hash_table_new_full(g_str_hash, g_str_equal, g_free, g_free);
    
    // Start animation timer
    renderer->animation_timer_id = g_timeout_add(16, animation_timer_callback, NULL); // ~60 FPS
    
    // Initialize OpenGL
    if (!init_opengl()) {
        g_warning("Failed to initialize OpenGL");
        renderer->opengl_enabled = FALSE;
    } else {
        renderer->opengl_enabled = TRUE;
        renderer->hardware_acceleration = TRUE;
    }
    
    renderer->initialized = TRUE;
    g_print("Renderer initialized (OpenGL: %s)\n", 
            renderer->opengl_enabled ? "enabled" : "disabled");
    
    return TRUE;
}

void taishang_renderer_cleanup(void) {
    if (renderer == NULL) {
        return;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    // Stop animation timer
    if (renderer->animation_timer_id > 0) {
        g_source_remove(renderer->animation_timer_id);
    }
    
    // Cleanup animations
    if (renderer->animations) {
        g_hash_table_destroy(renderer->animations);
    }
    
    // Cleanup OpenGL
    cleanup_opengl();
    
    g_mutex_unlock(&renderer->mutex);
    g_mutex_clear(&renderer->mutex);
    
    g_free(renderer);
    renderer = NULL;
    
    g_print("Renderer cleaned up\n");
}

TaishangRenderer *taishang_renderer_get_instance(void) {
    return renderer;
}

// Rendering functions
gboolean taishang_renderer_begin_frame(void) {
    if (!renderer || !renderer->opengl_enabled) {
        return FALSE;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    // Make context current
    if (renderer->gl_context) {
        gdk_gl_context_make_current(renderer->gl_context);
    }
    
    // Update time
    renderer->current_time = g_get_monotonic_time() / 1000000.0;
    
    // Update statistics
    renderer->stats.frame_count++;
    
    g_mutex_unlock(&renderer->mutex);
    
    return TRUE;
}

gboolean taishang_renderer_end_frame(void) {
    if (!renderer || !renderer->opengl_enabled) {
        return FALSE;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    // Swap buffers
    if (renderer->gl_area) {
        gtk_gl_area_queue_render(GTK_GL_AREA(renderer->gl_area));
    }
    
    // Update FPS
    static gdouble last_fps_time = 0;
    static gint fps_frame_count = 0;
    
    fps_frame_count++;
    gdouble current_time = g_get_monotonic_time() / 1000000.0;
    
    if (current_time - last_fps_time >= 1.0) {
        renderer->stats.fps = fps_frame_count / (current_time - last_fps_time);
        fps_frame_count = 0;
        last_fps_time = current_time;
    }
    
    g_mutex_unlock(&renderer->mutex);
    
    return TRUE;
}

void taishang_renderer_clear(float r, float g, float b, float a) {
    if (!renderer || !renderer->opengl_enabled) {
        return;
    }
    
    glClearColor(r, g, b, a);
    glClear(GL_COLOR_BUFFER_BIT | GL_DEPTH_BUFFER_BIT);
    
    renderer->stats.draw_calls++;
}

void taishang_renderer_set_viewport(gint x, gint y, gint width, gint height) {
    if (!renderer || !renderer->opengl_enabled) {
        return;
    }
    
    glViewport(x, y, width, height);
    
    // Update projection matrix for 2D rendering
    float left = 0.0f;
    float right = (float)width;
    float bottom = (float)height;
    float top = 0.0f;
    float near_plane = -1.0f;
    float far_plane = 1.0f;
    
    matrix_identity(renderer->projection_matrix);
    
    // Orthographic projection
    renderer->projection_matrix[0] = 2.0f / (right - left);
    renderer->projection_matrix[5] = 2.0f / (top - bottom);
    renderer->projection_matrix[10] = -2.0f / (far_plane - near_plane);
    renderer->projection_matrix[12] = -(right + left) / (right - left);
    renderer->projection_matrix[13] = -(top + bottom) / (top - bottom);
    renderer->projection_matrix[14] = -(far_plane + near_plane) / (far_plane - near_plane);
    renderer->projection_matrix[15] = 1.0f;
}

// Drawing functions
void taishang_renderer_draw_rectangle(float x, float y, float width, float height, 
                                      float r, float g, float b, float a) {
    if (!renderer || !renderer->opengl_enabled) {
        return;
    }
    
    // Define rectangle vertices
    float vertices[] = {
        x,         y,          0.0f, 0.0f, 0.0f, r, g, b, a,  // Bottom-left
        x + width, y,          0.0f, 1.0f, 0.0f, r, g, b, a,  // Bottom-right
        x + width, y + height, 0.0f, 1.0f, 1.0f, r, g, b, a,  // Top-right
        x,         y + height, 0.0f, 0.0f, 1.0f, r, g, b, a   // Top-left
    };
    
    unsigned int indices[] = {
        0, 1, 2,
        2, 3, 0
    };
    
    // Use shader program
    glUseProgram(renderer->shader_program);
    
    // Set uniforms
    GLint projection_loc = glGetUniformLocation(renderer->shader_program, "uProjection");
    GLint view_loc = glGetUniformLocation(renderer->shader_program, "uView");
    GLint model_loc = glGetUniformLocation(renderer->shader_program, "uModel");
    GLint use_texture_loc = glGetUniformLocation(renderer->shader_program, "uUseTexture");
    GLint opacity_loc = glGetUniformLocation(renderer->shader_program, "uOpacity");
    
    glUniformMatrix4fv(projection_loc, 1, GL_FALSE, renderer->projection_matrix);
    glUniformMatrix4fv(view_loc, 1, GL_FALSE, renderer->view_matrix);
    glUniformMatrix4fv(model_loc, 1, GL_FALSE, renderer->model_matrix);
    glUniform1i(use_texture_loc, GL_FALSE);
    glUniform1f(opacity_loc, 1.0f);
    
    // Bind and update vertex buffer
    glBindBuffer(GL_ARRAY_BUFFER, renderer->vertex_buffer);
    glBufferData(GL_ARRAY_BUFFER, sizeof(vertices), vertices, GL_DYNAMIC_DRAW);
    
    // Bind and update index buffer
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, renderer->index_buffer);
    glBufferData(GL_ELEMENT_ARRAY_BUFFER, sizeof(indices), indices, GL_DYNAMIC_DRAW);
    
    // Set vertex attributes
    glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)0);
    glEnableVertexAttribArray(0);
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)(3 * sizeof(float)));
    glEnableVertexAttribArray(1);
    glVertexAttribPointer(2, 4, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)(5 * sizeof(float)));
    glEnableVertexAttribArray(2);
    
    // Draw
    glDrawElements(GL_TRIANGLES, 6, GL_UNSIGNED_INT, 0);
    
    renderer->stats.draw_calls++;
    renderer->stats.triangles_rendered += 2;
}

void taishang_renderer_draw_circle(float x, float y, float radius, 
                                   float r, float g, float b, float a) {
    if (!renderer || !renderer->opengl_enabled) {
        return;
    }
    
    const int segments = 32;
    const float angle_step = 2.0f * M_PI / segments;
    
    // Create circle vertices
    float *vertices = g_malloc((segments + 2) * 9 * sizeof(float));
    unsigned int *indices = g_malloc(segments * 3 * sizeof(unsigned int));
    
    // Center vertex
    vertices[0] = x;
    vertices[1] = y;
    vertices[2] = 0.0f;
    vertices[3] = 0.5f; // tex coord u
    vertices[4] = 0.5f; // tex coord v
    vertices[5] = r;
    vertices[6] = g;
    vertices[7] = b;
    vertices[8] = a;
    
    // Circle vertices
    for (int i = 0; i <= segments; i++) {
        float angle = i * angle_step;
        int vertex_index = (i + 1) * 9;
        
        vertices[vertex_index + 0] = x + cosf(angle) * radius;
        vertices[vertex_index + 1] = y + sinf(angle) * radius;
        vertices[vertex_index + 2] = 0.0f;
        vertices[vertex_index + 3] = 0.5f + cosf(angle) * 0.5f;
        vertices[vertex_index + 4] = 0.5f + sinf(angle) * 0.5f;
        vertices[vertex_index + 5] = r;
        vertices[vertex_index + 6] = g;
        vertices[vertex_index + 7] = b;
        vertices[vertex_index + 8] = a;
    }
    
    // Create indices
    for (int i = 0; i < segments; i++) {
        indices[i * 3 + 0] = 0;
        indices[i * 3 + 1] = i + 1;
        indices[i * 3 + 2] = i + 2;
    }
    
    // Use shader program
    glUseProgram(renderer->shader_program);
    
    // Set uniforms
    GLint projection_loc = glGetUniformLocation(renderer->shader_program, "uProjection");
    GLint view_loc = glGetUniformLocation(renderer->shader_program, "uView");
    GLint model_loc = glGetUniformLocation(renderer->shader_program, "uModel");
    GLint use_texture_loc = glGetUniformLocation(renderer->shader_program, "uUseTexture");
    GLint opacity_loc = glGetUniformLocation(renderer->shader_program, "uOpacity");
    
    glUniformMatrix4fv(projection_loc, 1, GL_FALSE, renderer->projection_matrix);
    glUniformMatrix4fv(view_loc, 1, GL_FALSE, renderer->view_matrix);
    glUniformMatrix4fv(model_loc, 1, GL_FALSE, renderer->model_matrix);
    glUniform1i(use_texture_loc, GL_FALSE);
    glUniform1f(opacity_loc, 1.0f);
    
    // Bind and update buffers
    glBindBuffer(GL_ARRAY_BUFFER, renderer->vertex_buffer);
    glBufferData(GL_ARRAY_BUFFER, (segments + 2) * 9 * sizeof(float), vertices, GL_DYNAMIC_DRAW);
    
    glBindBuffer(GL_ELEMENT_ARRAY_BUFFER, renderer->index_buffer);
    glBufferData(GL_ELEMENT_ARRAY_BUFFER, segments * 3 * sizeof(unsigned int), indices, GL_DYNAMIC_DRAW);
    
    // Set vertex attributes
    glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)0);
    glEnableVertexAttribArray(0);
    glVertexAttribPointer(1, 2, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)(3 * sizeof(float)));
    glEnableVertexAttribArray(1);
    glVertexAttribPointer(2, 4, GL_FLOAT, GL_FALSE, 9 * sizeof(float), (void*)(5 * sizeof(float)));
    glEnableVertexAttribArray(2);
    
    // Draw
    glDrawElements(GL_TRIANGLES, segments * 3, GL_UNSIGNED_INT, 0);
    
    g_free(vertices);
    g_free(indices);
    
    renderer->stats.draw_calls++;
    renderer->stats.triangles_rendered += segments;
}

// Animation functions
TaishangAnimation *taishang_renderer_create_animation(const char *name, 
                                                      TaishangAnimationType type,
                                                      gdouble duration, 
                                                      gdouble start_value, 
                                                      gdouble end_value) {
    if (!renderer || !name) {
        return NULL;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    TaishangAnimation *animation = g_new0(TaishangAnimation, 1);
    animation->name = g_strdup(name);
    animation->type = type;
    animation->duration = duration;
    animation->start_value = start_value;
    animation->end_value = end_value;
    animation->current_value = start_value;
    animation->easing = TAISHANG_EASING_LINEAR;
    animation->loop = FALSE;
    animation->reverse = FALSE;
    animation->active = FALSE;
    
    g_hash_table_insert(renderer->animations, g_strdup(name), animation);
    
    g_mutex_unlock(&renderer->mutex);
    
    g_print("Animation created: %s\n", name);
    return animation;
}

gboolean taishang_renderer_start_animation(const char *name) {
    if (!renderer || !name) {
        return FALSE;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    TaishangAnimation *animation = g_hash_table_lookup(renderer->animations, name);
    if (animation) {
        animation->start_time = renderer->current_time;
        animation->active = TRUE;
        g_print("Animation started: %s\n", name);
    }
    
    g_mutex_unlock(&renderer->mutex);
    
    return animation != NULL;
}

gdouble taishang_renderer_get_animation_value(const char *name) {
    if (!renderer || !name) {
        return 0.0;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    TaishangAnimation *animation = g_hash_table_lookup(renderer->animations, name);
    gdouble value = animation ? animation->current_value : 0.0;
    
    g_mutex_unlock(&renderer->mutex);
    
    return value;
}

// Statistics functions
TaishangRenderStats taishang_renderer_get_stats(void) {
    if (!renderer) {
        TaishangRenderStats empty_stats = {0};
        return empty_stats;
    }
    
    return renderer->stats;
}

void taishang_renderer_reset_stats(void) {
    if (!renderer) {
        return;
    }
    
    g_mutex_lock(&renderer->mutex);
    memset(&renderer->stats, 0, sizeof(TaishangRenderStats));
    g_mutex_unlock(&renderer->mutex);
}

// Private functions
static gboolean init_opengl(void) {
    if (!renderer) {
        return FALSE;
    }
    
    // Create GL area widget
    renderer->gl_area = gtk_gl_area_new();
    if (!renderer->gl_area) {
        g_warning("Failed to create GL area");
        return FALSE;
    }
    
    gtk_gl_area_set_required_version(GTK_GL_AREA(renderer->gl_area), 3, 3);
    
    // Get GL context
    gtk_widget_realize(renderer->gl_area);
    renderer->gl_context = gtk_gl_area_get_context(GTK_GL_AREA(renderer->gl_area));
    
    if (!renderer->gl_context) {
        g_warning("Failed to get GL context");
        return FALSE;
    }
    
    gdk_gl_context_make_current(renderer->gl_context);
    
    // Check OpenGL version
    const char *version = (const char*)glGetString(GL_VERSION);
    g_print("OpenGL version: %s\n", version ? version : "unknown");
    
    // Compile shaders
    if (!compile_shaders()) {
        g_warning("Failed to compile shaders");
        return FALSE;
    }
    
    // Setup buffers
    setup_buffers();
    
    // Enable blending
    glEnable(GL_BLEND);
    glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);
    
    g_print("OpenGL initialized successfully\n");
    return TRUE;
}

static void cleanup_opengl(void) {
    if (!renderer) {
        return;
    }
    
    if (renderer->gl_context) {
        gdk_gl_context_make_current(renderer->gl_context);
        
        cleanup_buffers();
        
        if (renderer->shader_program) {
            glDeleteProgram(renderer->shader_program);
        }
        if (renderer->vertex_shader) {
            glDeleteShader(renderer->vertex_shader);
        }
        if (renderer->fragment_shader) {
            glDeleteShader(renderer->fragment_shader);
        }
    }
    
    if (renderer->gl_area) {
        gtk_widget_destroy(renderer->gl_area);
    }
}

static gboolean compile_shaders(void) {
    GLint success;
    GLchar info_log[512];
    
    // Compile vertex shader
    renderer->vertex_shader = glCreateShader(GL_VERTEX_SHADER);
    glShaderSource(renderer->vertex_shader, 1, &vertex_shader_source, NULL);
    glCompileShader(renderer->vertex_shader);
    
    glGetShaderiv(renderer->vertex_shader, GL_COMPILE_STATUS, &success);
    if (!success) {
        glGetShaderInfoLog(renderer->vertex_shader, 512, NULL, info_log);
        g_warning("Vertex shader compilation failed: %s", info_log);
        return FALSE;
    }
    
    // Compile fragment shader
    renderer->fragment_shader = glCreateShader(GL_FRAGMENT_SHADER);
    glShaderSource(renderer->fragment_shader, 1, &fragment_shader_source, NULL);
    glCompileShader(renderer->fragment_shader);
    
    glGetShaderiv(renderer->fragment_shader, GL_COMPILE_STATUS, &success);
    if (!success) {
        glGetShaderInfoLog(renderer->fragment_shader, 512, NULL, info_log);
        g_warning("Fragment shader compilation failed: %s", info_log);
        return FALSE;
    }
    
    // Create shader program
    renderer->shader_program = glCreateProgram();
    glAttachShader(renderer->shader_program, renderer->vertex_shader);
    glAttachShader(renderer->shader_program, renderer->fragment_shader);
    glLinkProgram(renderer->shader_program);
    
    glGetProgramiv(renderer->shader_program, GL_LINK_STATUS, &success);
    if (!success) {
        glGetProgramInfoLog(renderer->shader_program, 512, NULL, info_log);
        g_warning("Shader program linking failed: %s", info_log);
        return FALSE;
    }
    
    g_print("Shaders compiled successfully\n");
    return TRUE;
}

static void setup_buffers(void) {
    glGenBuffers(1, &renderer->vertex_buffer);
    glGenBuffers(1, &renderer->index_buffer);
    glGenBuffers(1, &renderer->texture_buffer);
}

static void cleanup_buffers(void) {
    if (renderer->vertex_buffer) {
        glDeleteBuffers(1, &renderer->vertex_buffer);
    }
    if (renderer->index_buffer) {
        glDeleteBuffers(1, &renderer->index_buffer);
    }
    if (renderer->texture_buffer) {
        glDeleteBuffers(1, &renderer->texture_buffer);
    }
}

static gboolean animation_timer_callback(gpointer user_data) {
    if (!renderer) {
        return G_SOURCE_REMOVE;
    }
    
    gdouble current_time = g_get_monotonic_time() / 1000000.0;
    update_animations(current_time);
    
    return G_SOURCE_CONTINUE;
}

static void update_animations(gdouble current_time) {
    if (!renderer || !renderer->animations) {
        return;
    }
    
    g_mutex_lock(&renderer->mutex);
    
    GHashTableIter iter;
    gpointer key, value;
    
    g_hash_table_iter_init(&iter, renderer->animations);
    while (g_hash_table_iter_next(&iter, &key, &value)) {
        TaishangAnimation *animation = (TaishangAnimation*)value;
        
        if (!animation->active) {
            continue;
        }
        
        gdouble elapsed = current_time - animation->start_time;
        gdouble progress = elapsed / animation->duration;
        
        if (progress >= 1.0) {
            if (animation->loop) {
                animation->start_time = current_time;
                progress = 0.0;
            } else {
                progress = 1.0;
                animation->active = FALSE;
            }
        }
        
        // Apply easing
        gdouble eased_progress = apply_easing(progress, animation->easing);
        
        // Calculate current value
        animation->current_value = animation->start_value + 
            (animation->end_value - animation->start_value) * eased_progress;
        
        // Call callback if set
        if (animation->callback) {
            animation->callback(animation->name, animation->current_value, 
                               progress, animation->user_data);
        }
    }
    
    g_mutex_unlock(&renderer->mutex);
}

static gdouble apply_easing(gdouble t, TaishangAnimationEasing easing) {
    switch (easing) {
        case TAISHANG_EASING_LINEAR:
            return t;
        case TAISHANG_EASING_EASE_IN:
            return t * t;
        case TAISHANG_EASING_EASE_OUT:
            return 1.0 - (1.0 - t) * (1.0 - t);
        case TAISHANG_EASING_EASE_IN_OUT:
            return t < 0.5 ? 2.0 * t * t : 1.0 - 2.0 * (1.0 - t) * (1.0 - t);
        case TAISHANG_EASING_BOUNCE:
            if (t < 1.0 / 2.75) {
                return 7.5625 * t * t;
            } else if (t < 2.0 / 2.75) {
                t -= 1.5 / 2.75;
                return 7.5625 * t * t + 0.75;
            } else if (t < 2.5 / 2.75) {
                t -= 2.25 / 2.75;
                return 7.5625 * t * t + 0.9375;
            } else {
                t -= 2.625 / 2.75;
                return 7.5625 * t * t + 0.984375;
            }
        default:
            return t;
    }
}

static void matrix_identity(float *matrix) {
    memset(matrix, 0, 16 * sizeof(float));
    matrix[0] = matrix[5] = matrix[10] = matrix[15] = 1.0f;
}