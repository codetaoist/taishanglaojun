import { useState } from 'react';
import { 
  Image as ImageIcon, 
  Upload, 
  Download, 
  Loader2, 
  Wand2,
  Eye,
  Settings,
  Palette
} from 'lucide-react';
import { cn } from '../utils/cn';
import aiService from '../services/aiService';

interface GeneratedImage {
  id: string;
  url: string;
  prompt: string;
  style: string;
  width: number;
  height: number;
  created_at: string;
}

interface ImageAnalysisResult {
  description: string;
  objects: Array<{
    name: string;
    confidence: number;
    bounding_box: {
      x: number;
      y: number;
      width: number;
      height: number;
    };
  }>;
  colors: Array<{
    color: string;
    percentage: number;
  }>;
  tags: string[];
}

export default function ImagePage() {
  const [activeTab, setActiveTab] = useState<'generate' | 'analyze' | 'edit'>('generate');
  
  // Generation state
  const [prompt, setPrompt] = useState('');
  const [negativePrompt, setNegativePrompt] = useState('');
  const [selectedStyle, setSelectedStyle] = useState('realistic');
  const [width, setWidth] = useState(512);
  const [height, setHeight] = useState(512);
  const [generating, setGenerating] = useState(false);
  const [generatedImages, setGeneratedImages] = useState<GeneratedImage[]>([]);
  
  // Analysis state
  const [selectedImagePath, setSelectedImagePath] = useState<string>('');
  const [analyzing, setAnalyzing] = useState(false);
  const [analysisResult, setAnalysisResult] = useState<ImageAnalysisResult | null>(null);
  
  // Edit state
  const [editImagePath, setEditImagePath] = useState<string>('');
  const [editOperation, setEditOperation] = useState('enhance');
  const [editing, setEditing] = useState(false);

  const styles = [
    { id: 'realistic', name: '写实风格', description: '逼真的照片风格' },
    { id: 'anime', name: '动漫风格', description: '日式动漫插画风格' },
    { id: 'oil_painting', name: '油画风格', description: '经典油画艺术风格' },
    { id: 'watercolor', name: '水彩风格', description: '水彩画艺术风格' },
    { id: 'sketch', name: '素描风格', description: '铅笔素描风格' },
    { id: 'digital_art', name: '数字艺术', description: '现代数字艺术风格' },
  ];

  const editOperations = [
    { id: 'enhance', name: '图像增强', description: '提升图像质量和清晰度' },
    { id: 'upscale', name: '图像放大', description: '智能放大图像分辨率' },
    { id: 'remove_background', name: '背景移除', description: '自动移除图像背景' },
    { id: 'style_transfer', name: '风格转换', description: '转换图像艺术风格' },
    { id: 'colorize', name: '图像上色', description: '为黑白图像添加颜色' },
    { id: 'denoise', name: '降噪处理', description: '减少图像噪点' },
  ];

  const generateImage = async () => {
    if (!prompt.trim()) return;

    setGenerating(true);
    try {
      // 使用新的AI服务生成图像
      const response = await aiService.generateImage({
        prompt: prompt.trim(),
        negativePrompt: negativePrompt.trim() || undefined,
        style: selectedStyle,
        size: `${width}x${height}`,
        model: 'stable-diffusion-xl',
        provider: 'local'
      });

      if (response.success) {
        const newImages = response.result.images.map((img: any) => ({
          id: Date.now().toString() + Math.random(),
          url: img.url,
          prompt: prompt,
          style: selectedStyle,
          width,
          height,
          created_at: new Date().toISOString(),
        }));
        
        setGeneratedImages(prev => [...newImages, ...prev]);
      }
    } catch (error) {
      console.error('Failed to generate image:', error);
    } finally {
      setGenerating(false);
    }
  };

  const selectImageForAnalysis = async () => {
    try {
      // 模拟图像选择，使用示例图像
      const mockImagePath = 'data:image/svg+xml;base64,' + btoa(`
        <svg width="400" height="300" xmlns="http://www.w3.org/2000/svg">
          <rect width="100%" height="100%" fill="#f3f4f6"/>
          <circle cx="200" cy="150" r="80" fill="#3b82f6"/>
          <rect x="120" y="200" width="160" height="60" fill="#ef4444"/>
          <text x="200" y="280" text-anchor="middle" font-family="Arial" font-size="12" fill="#374151">
            示例图像用于分析
          </text>
        </svg>
      `);
      setSelectedImagePath(mockImagePath);
      setAnalysisResult(null);
    } catch (error) {
      console.error('Failed to select image:', error);
    }
  };

  const analyzeImage = async () => {
    if (!selectedImagePath) return;

    setAnalyzing(true);
    try {
      // 使用新的AI服务分析图像
      const response = await aiService.analyzeImage({
        imageData: selectedImagePath,
        analysisTypes: ['description', 'objects', 'colors', 'tags'],
        model: 'gpt-4-vision',
        provider: 'openai'
      });

      if (response.success) {
        setAnalysisResult(response.result);
      }
    } catch (error) {
      console.error('Failed to analyze image:', error);
    } finally {
      setAnalyzing(false);
    }
  };

  const selectImageForEdit = async () => {
    try {
      // 在浏览器模式下模拟图像选择
      const mockImagePath = 'data:image/svg+xml;base64,' + btoa(`
        <svg width="300" height="200" xmlns="http://www.w3.org/2000/svg">
          <rect width="100%" height="100%" fill="#e5e7eb"/>
          <text x="50%" y="50%" text-anchor="middle" dy=".3em" font-family="Arial" font-size="14" fill="#6b7280">
            选择的图像文件
          </text>
        </svg>
      `);
      setEditImagePath(mockImagePath);
    } catch (error) {
      console.error('Failed to select image:', error);
    }
  };

  const editImage = async () => {
    if (!editImagePath) return;

    setEditing(true);
    try {
      // 使用AI服务进行图像编辑（模拟）
      console.log('Image editing with AI service:', {
        imagePath: editImagePath,
        operation: editOperation
      });
      
      // 这里可以集成实际的图像编辑AI服务
      // const response = await aiService.editImage({
      //   imagePath: editImagePath,
      //   operation: editOperation,
      //   model: 'stable-diffusion-inpaint',
      //   provider: 'local'
      // });
      
    } catch (error) {
      console.error('Failed to edit image:', error);
    } finally {
      setEditing(false);
    }
  };

  const saveImage = async (imageUrl: string, filename: string) => {
    try {
      // 通过浏览器下载功能保存图像
      const a = document.createElement('a');
      a.href = imageUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (error) {
      console.error('Failed to save image:', error);
    }
  };

  const tabs = [
    { id: 'generate', label: '图像生成', icon: Wand2 },
    { id: 'analyze', label: '图像分析', icon: Eye },
    { id: 'edit', label: '图像编辑', icon: Settings },
  ];

  return (
    <div className="h-full flex flex-col space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">图像处理</h1>
        <p className="text-muted-foreground mt-2">
          AI驱动的图像生成、分析和编辑工具
        </p>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 bg-secondary/20 p-1 rounded-lg">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as any)}
            className={cn(
              'flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors',
              activeTab === tab.id
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <tab.icon className="h-4 w-4" />
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Generate Tab */}
      {activeTab === 'generate' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Generation Controls */}
          <div className="card p-6 space-y-4">
            <h2 className="text-lg font-semibold">图像生成设置</h2>
            
            <div>
              <label className="block text-sm font-medium mb-2">描述提示词</label>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder="描述您想要生成的图像..."
                className="input w-full h-24 resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">负面提示词（可选）</label>
              <textarea
                value={negativePrompt}
                onChange={(e) => setNegativePrompt(e.target.value)}
                placeholder="描述您不想要的元素..."
                className="input w-full h-16 resize-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">艺术风格</label>
              <div className="grid grid-cols-2 gap-2">
                {styles.map((style) => (
                  <button
                    key={style.id}
                    onClick={() => setSelectedStyle(style.id)}
                    className={cn(
                      'p-3 text-left border rounded-lg transition-colors',
                      selectedStyle === style.id
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50'
                    )}
                  >
                    <p className="font-medium text-sm">{style.name}</p>
                    <p className="text-xs text-muted-foreground">{style.description}</p>
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">宽度</label>
                <select
                  value={width}
                  onChange={(e) => setWidth(Number(e.target.value))}
                  className="input w-full"
                >
                  <option value={512}>512px</option>
                  <option value={768}>768px</option>
                  <option value={1024}>1024px</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">高度</label>
                <select
                  value={height}
                  onChange={(e) => setHeight(Number(e.target.value))}
                  className="input w-full"
                >
                  <option value={512}>512px</option>
                  <option value={768}>768px</option>
                  <option value={1024}>1024px</option>
                </select>
              </div>
            </div>

            <button
              onClick={generateImage}
              disabled={!prompt.trim() || generating}
              className={cn(
                'btn-primary w-full',
                (!prompt.trim() || generating) && 'opacity-50 cursor-not-allowed'
              )}
            >
              {generating ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  生成中...
                </>
              ) : (
                <>
                  <Wand2 className="h-4 w-4 mr-2" />
                  生成图像
                </>
              )}
            </button>
          </div>

          {/* Generated Images */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4">生成的图像</h2>
            
            {generatedImages.length === 0 ? (
              <div className="flex items-center justify-center h-64 text-muted-foreground">
                <div className="text-center">
                  <ImageIcon className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>还没有生成的图像</p>
                  <p className="text-sm mt-2">输入提示词开始生成</p>
                </div>
              </div>
            ) : (
              <div className="space-y-4 max-h-96 overflow-y-auto">
                {generatedImages.map((image) => (
                  <div key={image.id} className="border border-border rounded-lg p-4">
                    <img
                      src={image.url}
                      alt={image.prompt}
                      className="w-full h-48 object-cover rounded-lg mb-3"
                    />
                    <div className="space-y-2">
                      <p className="text-sm font-medium">{image.prompt}</p>
                      <p className="text-xs text-muted-foreground">
                        {image.style} • {image.width}×{image.height}
                      </p>
                      <button
                        onClick={() => saveImage(image.url, `generated_${image.id}.png`)}
                        className="btn-secondary text-sm w-full"
                      >
                        <Download className="h-4 w-4 mr-2" />
                        保存图像
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Analyze Tab */}
      {activeTab === 'analyze' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4">图像分析</h2>
            
            <div className="space-y-4">
              <button
                onClick={selectImageForAnalysis}
                className="btn-secondary w-full"
              >
                <Upload className="h-4 w-4 mr-2" />
                选择图像文件
              </button>

              {selectedImagePath && (
                <div className="space-y-4">
                  <div className="border border-border rounded-lg p-4">
                    <p className="text-sm text-muted-foreground mb-2">已选择文件：</p>
                    <p className="text-sm font-medium">{selectedImagePath.split('\\').pop()}</p>
                  </div>

                  <button
                    onClick={analyzeImage}
                    disabled={analyzing}
                    className={cn(
                      'btn-primary w-full',
                      analyzing && 'opacity-50 cursor-not-allowed'
                    )}
                  >
                    {analyzing ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        分析中...
                      </>
                    ) : (
                      <>
                        <Eye className="h-4 w-4 mr-2" />
                        开始分析
                      </>
                    )}
                  </button>
                </div>
              )}
            </div>
          </div>

          {analysisResult && (
            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4">分析结果</h2>
              
              <div className="space-y-4">
                <div>
                  <h3 className="font-medium mb-2">图像描述</h3>
                  <p className="text-sm text-muted-foreground bg-secondary/20 p-3 rounded">
                    {analysisResult.description}
                  </p>
                </div>

                {analysisResult.objects.length > 0 && (
                  <div>
                    <h3 className="font-medium mb-2">检测到的对象</h3>
                    <div className="space-y-2">
                      {analysisResult.objects.map((obj, index) => (
                        <div key={index} className="flex justify-between items-center p-2 bg-secondary/20 rounded">
                          <span className="text-sm">{obj.name}</span>
                          <span className="text-xs text-muted-foreground">
                            {(obj.confidence * 100).toFixed(1)}%
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {analysisResult.tags.length > 0 && (
                  <div>
                    <h3 className="font-medium mb-2">标签</h3>
                    <div className="flex flex-wrap gap-2">
                      {analysisResult.tags.map((tag, index) => (
                        <span
                          key={index}
                          className="px-2 py-1 bg-primary/10 text-primary text-xs rounded"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Edit Tab */}
      {activeTab === 'edit' && (
        <div className="card p-6">
          <h2 className="text-lg font-semibold mb-4">图像编辑</h2>
          
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="space-y-4">
              <button
                onClick={selectImageForEdit}
                className="btn-secondary w-full"
              >
                <Upload className="h-4 w-4 mr-2" />
                选择要编辑的图像
              </button>

              {editImagePath && (
                <div className="border border-border rounded-lg p-4">
                  <p className="text-sm text-muted-foreground mb-2">已选择文件：</p>
                  <p className="text-sm font-medium">{editImagePath.split('\\').pop()}</p>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium mb-2">编辑操作</label>
                <div className="space-y-2">
                  {editOperations.map((operation) => (
                    <button
                      key={operation.id}
                      onClick={() => setEditOperation(operation.id)}
                      className={cn(
                        'w-full p-3 text-left border rounded-lg transition-colors',
                        editOperation === operation.id
                          ? 'border-primary bg-primary/10'
                          : 'border-border hover:border-primary/50'
                      )}
                    >
                      <p className="font-medium text-sm">{operation.name}</p>
                      <p className="text-xs text-muted-foreground">{operation.description}</p>
                    </button>
                  ))}
                </div>
              </div>

              <button
                onClick={editImage}
                disabled={!editImagePath || editing}
                className={cn(
                  'btn-primary w-full',
                  (!editImagePath || editing) && 'opacity-50 cursor-not-allowed'
                )}
              >
                {editing ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    处理中...
                  </>
                ) : (
                  <>
                    <Palette className="h-4 w-4 mr-2" />
                    开始编辑
                  </>
                )}
              </button>
            </div>

            <div className="flex items-center justify-center h-64 border-2 border-dashed border-border rounded-lg">
              <div className="text-center text-muted-foreground">
                <ImageIcon className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>编辑结果将在这里显示</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}