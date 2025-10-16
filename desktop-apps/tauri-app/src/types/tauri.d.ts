// Tauri API 类型声明
declare global {
  interface Window {
    __TAURI__?: {
      invoke: (cmd: string, args?: any) => Promise<any>;
      convertFileSrc: (filePath: string) => string;
    };
  }
}

// AI 服务相关类型
export interface AIRequest {
  id: string;
  message: string;
  type: 'chat' | 'image' | 'document';
  timestamp: number;
}

export interface AIResponse {
  id: string;
  content: string;
  type: 'text' | 'image' | 'file';
  timestamp: number;
}

// 图像生成相关类型
export interface ImageGenerationRequest {
  prompt: string;
  style?: string;
  size?: string;
}

export interface ImageAnalysisRequest {
  imageData: string;
  analysisType: string;
}

// 文件传输相关类型
export interface FileTransferRequest {
  filePath: string;
  targetPath: string;
  transferType: 'upload' | 'download';
}

export interface FileTransferProgress {
  id: string;
  progress: number;
  status: 'pending' | 'transferring' | 'completed' | 'failed';
}

// 桌宠相关类型
export interface DesktopPetState {
  x: number;
  y: number;
  animation: string;
  mood: 'happy' | 'sad' | 'excited' | 'sleeping';
}

export interface DesktopPetAction {
  type: 'move' | 'animate' | 'speak' | 'interact';
  data: any;
}

export {};