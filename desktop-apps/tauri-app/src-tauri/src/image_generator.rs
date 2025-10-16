use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::fs;
use std::path::Path;
use base64::{Engine as _, engine::general_purpose};

use crate::ai_service::AIService;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImageGenerationRequest {
    pub prompt: String,
    pub negative_prompt: Option<String>,
    pub style: Option<String>,
    pub width: Option<u32>,
    pub height: Option<u32>,
    pub steps: Option<u32>,
    pub guidance_scale: Option<f32>,
    pub seed: Option<i64>,
    pub batch_size: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImageGenerationResult {
    pub success: bool,
    pub images: Vec<GeneratedImage>,
    pub generation_time: u64,
    pub parameters: ImageGenerationRequest,
    pub error: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GeneratedImage {
    pub id: String,
    pub base64_data: String,
    pub format: String,
    pub width: u32,
    pub height: u32,
    pub file_size: u64,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImageEditRequest {
    pub image_path: String,
    pub operation: String,
    pub parameters: Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImageAnalysisResult {
    pub description: String,
    pub objects: Vec<DetectedObject>,
    pub colors: Vec<ColorInfo>,
    pub style: String,
    pub quality_score: f32,
    pub metadata: Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DetectedObject {
    pub name: String,
    pub confidence: f32,
    pub bbox: BoundingBox,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BoundingBox {
    pub x: f32,
    pub y: f32,
    pub width: f32,
    pub height: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColorInfo {
    pub name: String,
    pub hex: String,
    pub rgb: (u8, u8, u8),
    pub percentage: f32,
}

pub struct ImageGenerator {
    ai_service: AIService,
}

impl ImageGenerator {
    pub async fn new() -> Result<Self> {
        let ai_service = AIService::new().await?;
        
        Ok(Self {
            ai_service,
        })
    }

    // 生成图像
    pub async fn generate_image(&self, request: ImageGenerationRequest) -> Result<ImageGenerationResult> {
        let start_time = std::time::Instant::now();
        
        // 构建多模态请求
        let multimodal_data = serde_json::json!({
            "text": {
                "prompt": request.prompt.clone(),
                "negative_prompt": request.negative_prompt.clone().unwrap_or_default(),
            },
            "generation_params": {
                "style": request.style.clone().unwrap_or_else(|| "realistic".to_string()),
                "width": request.width.unwrap_or(512),
                "height": request.height.unwrap_or(512),
                "steps": request.steps.unwrap_or(20),
                "guidance_scale": request.guidance_scale.unwrap_or(7.5),
                "seed": request.seed.unwrap_or(-1),
                "batch_size": request.batch_size.unwrap_or(1),
            }
        });

        let response = self.ai_service.process_multimodal(
            multimodal_data
        ).await?;

        let generation_time = start_time.elapsed().as_millis() as u64;

        if response.success {
            let mut images = Vec::new();
            
            if let Some(generated_images) = response.result.get("images") {
                if let Some(images_array) = generated_images.as_array() {
                    for (index, image_data) in images_array.iter().enumerate() {
                        if let Some(image_obj) = image_data.as_object() {
                            let base64_data = image_obj.get("data")
                                .and_then(|v| v.as_str())
                                .unwrap_or("")
                                .to_string();
                            
                            let width = image_obj.get("width")
                                .and_then(|v| v.as_u64())
                                .unwrap_or(512) as u32;
                            
                            let height = image_obj.get("height")
                                .and_then(|v| v.as_u64())
                                .unwrap_or(512) as u32;

                            let generated_image = GeneratedImage {
                                id: format!("img_{}_{}", chrono::Utc::now().timestamp(), index),
                                base64_data,
                                format: "png".to_string(),
                                width,
                                height,
                                file_size: 0, // 计算实际大小
                                created_at: chrono::Utc::now().to_rfc3339(),
                            };
                            
                            images.push(generated_image);
                        }
                    }
                }
            }

            Ok(ImageGenerationResult {
                success: true,
                images,
                generation_time,
                parameters: request,
                error: None,
            })
        } else {
            Ok(ImageGenerationResult {
                success: false,
                images: Vec::new(),
                generation_time,
                parameters: request,
                error: response.error,
            })
        }
    }

    // 保存生成的图像
    pub async fn save_image(&self, image: &GeneratedImage, save_path: &str) -> Result<String> {
        // 解码base64数据
        let image_data = general_purpose::STANDARD.decode(&image.base64_data)
            .map_err(|e| anyhow!("Base64解码失败: {}", e))?;

        // 确保目录存在
        if let Some(parent) = Path::new(save_path).parent() {
            fs::create_dir_all(parent)?;
        }

        // 写入文件
        fs::write(save_path, image_data)?;

        Ok(save_path.to_string())
    }

    // 图像编辑
    pub async fn edit_image(&self, request: ImageEditRequest) -> Result<GeneratedImage> {
        // 读取原始图像
        let image_data = fs::read(&request.image_path)?;
        let base64_image = general_purpose::STANDARD.encode(&image_data);

        let multimodal_data = serde_json::json!({
            "image": {
                "data": base64_image,
                "format": "auto"
            },
            "operation": request.operation,
            "parameters": request.parameters
        });

        let response = self.ai_service.process_multimodal(
            multimodal_data
        ).await?;

        if response.success {
            if let Some(edited_image) = response.result.get("edited_image") {
                if let Some(image_obj) = edited_image.as_object() {
                    let base64_data = image_obj.get("data")
                        .and_then(|v| v.as_str())
                        .unwrap_or("")
                        .to_string();
                    
                    let width = image_obj.get("width")
                        .and_then(|v| v.as_u64())
                        .unwrap_or(512) as u32;
                    
                    let height = image_obj.get("height")
                        .and_then(|v| v.as_u64())
                        .unwrap_or(512) as u32;

                    return Ok(GeneratedImage {
                        id: format!("edited_{}", chrono::Utc::now().timestamp()),
                        base64_data,
                        format: "png".to_string(),
                        width,
                        height,
                        file_size: 0,
                        created_at: chrono::Utc::now().to_rfc3339(),
                    });
                }
            }
        }

        Err(anyhow!("图像编辑失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
    }

    // 图像分析
    pub async fn analyze_image(&self, image_path: String) -> Result<ImageAnalysisResult> {
        // 读取图像文件
        let image_data = fs::read(&image_path)?;
        let base64_image = general_purpose::STANDARD.encode(&image_data);

        let multimodal_data = serde_json::json!({
            "image": {
                "data": base64_image,
                "format": "auto"
            }
        });

        let response = self.ai_service.process_multimodal(
            multimodal_data
        ).await?;

        if response.success {
            let description = response.result.get("description")
                .and_then(|v| v.as_str())
                .unwrap_or("无法生成描述")
                .to_string();

            let mut objects = Vec::new();
            if let Some(detected_objects) = response.result.get("objects") {
                if let Some(objects_array) = detected_objects.as_array() {
                    for obj in objects_array {
                        if let Some(obj_data) = obj.as_object() {
                            let name = obj_data.get("name")
                                .and_then(|v| v.as_str())
                                .unwrap_or("unknown")
                                .to_string();
                            
                            let confidence = obj_data.get("confidence")
                                .and_then(|v| v.as_f64())
                                .unwrap_or(0.0) as f32;

                            let bbox = if let Some(bbox_data) = obj_data.get("bbox") {
                                if let Some(bbox_obj) = bbox_data.as_object() {
                                    BoundingBox {
                                        x: bbox_obj.get("x").and_then(|v| v.as_f64()).unwrap_or(0.0) as f32,
                                        y: bbox_obj.get("y").and_then(|v| v.as_f64()).unwrap_or(0.0) as f32,
                                        width: bbox_obj.get("width").and_then(|v| v.as_f64()).unwrap_or(0.0) as f32,
                                        height: bbox_obj.get("height").and_then(|v| v.as_f64()).unwrap_or(0.0) as f32,
                                    }
                                } else {
                                    BoundingBox { x: 0.0, y: 0.0, width: 0.0, height: 0.0 }
                                }
                            } else {
                                BoundingBox { x: 0.0, y: 0.0, width: 0.0, height: 0.0 }
                            };

                            objects.push(DetectedObject {
                                name,
                                confidence,
                                bbox,
                            });
                        }
                    }
                }
            }

            let mut colors = Vec::new();
            if let Some(color_info) = response.result.get("colors") {
                if let Some(colors_array) = color_info.as_array() {
                    for color in colors_array {
                        if let Some(color_obj) = color.as_object() {
                            let name = color_obj.get("name")
                                .and_then(|v| v.as_str())
                                .unwrap_or("unknown")
                                .to_string();
                            
                            let hex = color_obj.get("hex")
                                .and_then(|v| v.as_str())
                                .unwrap_or("#000000")
                                .to_string();

                            let rgb = if let Some(rgb_array) = color_obj.get("rgb").and_then(|v| v.as_array()) {
                                (
                                    rgb_array.get(0).and_then(|v| v.as_u64()).unwrap_or(0) as u8,
                                    rgb_array.get(1).and_then(|v| v.as_u64()).unwrap_or(0) as u8,
                                    rgb_array.get(2).and_then(|v| v.as_u64()).unwrap_or(0) as u8,
                                )
                            } else {
                                (0, 0, 0)
                            };

                            let percentage = color_obj.get("percentage")
                                .and_then(|v| v.as_f64())
                                .unwrap_or(0.0) as f32;

                            colors.push(ColorInfo {
                                name,
                                hex,
                                rgb,
                                percentage,
                            });
                        }
                    }
                }
            }

            let style = response.result.get("style")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown")
                .to_string();

            let quality_score = response.result.get("quality_score")
                .and_then(|v| v.as_f64())
                .unwrap_or(0.0) as f32;

            let metadata = response.result.get("metadata")
                .cloned()
                .unwrap_or(Value::Object(serde_json::Map::new()));

            Ok(ImageAnalysisResult {
                description,
                objects,
                colors,
                style,
                quality_score,
                metadata,
            })
        } else {
            Err(anyhow!("图像分析失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
        }
    }

    // 图像风格转换
    pub async fn style_transfer(&self, image_path: String, style: String) -> Result<GeneratedImage> {
        let image_data = fs::read(&image_path)?;
        let base64_image = general_purpose::STANDARD.encode(&image_data);

        let multimodal_data = serde_json::json!({
            "image": {
                "data": base64_image,
                "format": "auto"
            },
            "style": style
        });

        let response = self.ai_service.process_multimodal(
            multimodal_data
        ).await?;

        if response.success {
            if let Some(styled_image) = response.result.get("styled_image") {
                if let Some(image_obj) = styled_image.as_object() {
                    let base64_data = image_obj.get("data")
                        .and_then(|v| v.as_str())
                        .unwrap_or("")
                        .to_string();
                    
                    let width = image_obj.get("width")
                        .and_then(|v| v.as_u64())
                        .unwrap_or(512) as u32;
                    
                    let height = image_obj.get("height")
                        .and_then(|v| v.as_u64())
                        .unwrap_or(512) as u32;

                    return Ok(GeneratedImage {
                        id: format!("styled_{}", chrono::Utc::now().timestamp()),
                        base64_data,
                        format: "png".to_string(),
                        width,
                        height,
                        file_size: 0,
                        created_at: chrono::Utc::now().to_rfc3339(),
                    });
                }
            }
        }

        Err(anyhow!("风格转换失败: {}", response.error.unwrap_or_else(|| "未知错误".to_string())))
    }

    // 获取支持的图像格式
    pub fn supported_formats() -> Vec<String> {
        vec![
            "jpg".to_string(),
            "jpeg".to_string(),
            "png".to_string(),
            "gif".to_string(),
            "bmp".to_string(),
            "webp".to_string(),
        ]
    }

    // 检查图像格式是否支持
    pub fn is_supported_format(file_path: &str) -> bool {
        let path = Path::new(file_path);
        if let Some(extension) = path.extension().and_then(|ext| ext.to_str()) {
            Self::supported_formats().contains(&extension.to_lowercase())
        } else {
            false
        }
    }

    // 获取预设风格列表
    pub fn get_preset_styles() -> Vec<String> {
        vec![
            "realistic".to_string(),
            "anime".to_string(),
            "cartoon".to_string(),
            "oil_painting".to_string(),
            "watercolor".to_string(),
            "sketch".to_string(),
            "digital_art".to_string(),
            "photographic".to_string(),
            "3d_render".to_string(),
            "pixel_art".to_string(),
        ]
    }
}