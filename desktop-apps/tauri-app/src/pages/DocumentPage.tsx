import { useState } from 'react';
import { 
  Upload, 
  FileText, 
  Download, 
  Loader2, 
  CheckCircle,
  AlertCircle
} from 'lucide-react';
import { cn } from '../utils/cn';
import aiService from '../services/aiService';

interface ProcessingResult {
  operation: string;
  result: string;
  success: boolean;
}

interface DocumentInfo {
  name: string;
  size: number;
  file_type: string;
  path: string;
}

export default function DocumentPage() {
  const [selectedFile, setSelectedFile] = useState<DocumentInfo | null>(null);
  const [processing, setProcessing] = useState(false);
  const [results, setResults] = useState<ProcessingResult[]>([]);
  const [selectedOperations, setSelectedOperations] = useState<string[]>(['summarize']);


  const operations = [
    { id: 'summarize', label: '文档摘要', description: '生成文档的简洁摘要' },
    { id: 'extract_keywords', label: '关键词提取', description: '提取文档中的关键词' },
    { id: 'analyze_sentiment', label: '情感分析', description: '分析文档的情感倾向' },
    { id: 'translate', label: '翻译', description: '翻译文档内容' },
    { id: 'extract_entities', label: '实体提取', description: '提取人名、地名、机构名等' },
    { id: 'generate_outline', label: '生成大纲', description: '为文档生成结构化大纲' },
    { id: 'check_grammar', label: '语法检查', description: '检查并修正语法错误' },
    { id: 'improve_writing', label: '写作改进', description: '提供写作改进建议' },
  ];

  const handleFileSelect = async () => {
    try {
      // 模拟文件选择
      const mockFile: DocumentInfo = {
        name: 'sample_document.txt',
        size: 2048,
        file_type: 'text/plain',
        path: 'data:text/plain;base64,' + btoa('这是一个示例文档内容，用于演示AI文档处理功能。\n\n文档包含多个段落和不同的内容类型，可以用来测试摘要、关键词提取、情感分析等各种AI处理功能。\n\n这个文档展示了如何使用AI服务来处理和分析文档内容。')
      };
      setSelectedFile(mockFile);
      setResults([]);
    } catch (error) {
      console.error('Failed to select file:', error);
    }
  };

  const handleOperationToggle = (operationId: string) => {
    setSelectedOperations(prev => 
      prev.includes(operationId)
        ? prev.filter(id => id !== operationId)
        : [...prev, operationId]
    );
  };

  const processDocument = async () => {
    if (!selectedFile || selectedOperations.length === 0) return;

    setProcessing(true);
    setResults([]);

    try {
      for (const operation of selectedOperations) {
        // 使用新的AI服务处理文档
        const response = await aiService.processDocument({
          filePath: selectedFile.path,
          operation,
          model: 'gpt-4',
          provider: 'openai'
        });

        const result: ProcessingResult = {
          operation,
          result: response.success ? response.result.content : '处理失败，请重试',
          success: response.success
        };

        setResults(prev => [...prev, result]);
      }
    } catch (error) {
      console.error('Failed to process document:', error);
      setResults(prev => [...prev, {
        operation: 'error',
        result: '处理文档时发生错误，请重试。',
        success: false
      }]);
    } finally {
      setProcessing(false);
    }
  };

  const saveResult = async (result: ProcessingResult) => {
    try {
      // 通过浏览器下载功能保存结果
      const fileName = `${selectedFile?.name}_${result.operation}_${Date.now()}.txt`;
      const blob = new Blob([result.result], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to save result:', error);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="h-full flex flex-col space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">文档处理</h1>
        <p className="text-muted-foreground mt-2">
          上传文档并选择处理操作，AI将为您分析和处理文档内容
        </p>
      </div>

      {/* File upload area */}
      <div className="card p-6">
        <h2 className="text-lg font-semibold mb-4">选择文档</h2>
        
        {!selectedFile ? (
          <div
            onClick={handleFileSelect}
            className="border-2 border-dashed border-border rounded-lg p-8 text-center cursor-pointer hover:border-primary transition-colors"
          >
            <Upload className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
            <p className="text-lg font-medium text-foreground mb-2">
              点击选择文档文件
            </p>
            <p className="text-sm text-muted-foreground">
              支持 TXT, MD, HTML, PDF, DOCX 等格式
            </p>
          </div>
        ) : (
          <div className="flex items-center justify-between p-4 bg-secondary/20 rounded-lg">
            <div className="flex items-center space-x-3">
              <FileText className="h-8 w-8 text-primary" />
              <div>
                <p className="font-medium text-foreground">{selectedFile.name}</p>
                <p className="text-sm text-muted-foreground">
                  {formatFileSize(selectedFile.size)} • {selectedFile.file_type.toUpperCase()}
                </p>
              </div>
            </div>
            <button
              onClick={handleFileSelect}
              className="btn-secondary"
            >
              重新选择
            </button>
          </div>
        )}
      </div>

      {/* Operations selection */}
      {selectedFile && (
        <div className="card p-6">
          <h2 className="text-lg font-semibold mb-4">选择处理操作</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {operations.map((operation) => (
              <div
                key={operation.id}
                className={cn(
                  'p-4 border rounded-lg cursor-pointer transition-colors',
                  selectedOperations.includes(operation.id)
                    ? 'border-primary bg-primary/10'
                    : 'border-border hover:border-primary/50'
                )}
                onClick={() => handleOperationToggle(operation.id)}
              >
                <div className="flex items-center space-x-3">
                  <div className={cn(
                    'w-4 h-4 rounded border-2 flex items-center justify-center',
                    selectedOperations.includes(operation.id)
                      ? 'border-primary bg-primary'
                      : 'border-muted-foreground'
                  )}>
                    {selectedOperations.includes(operation.id) && (
                      <CheckCircle className="w-3 h-3 text-primary-foreground" />
                    )}
                  </div>
                  <div>
                    <p className="font-medium text-foreground">{operation.label}</p>
                    <p className="text-sm text-muted-foreground">{operation.description}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <button
            onClick={processDocument}
            disabled={selectedOperations.length === 0 || processing}
            className={cn(
              'btn-primary',
              (selectedOperations.length === 0 || processing) && 'opacity-50 cursor-not-allowed'
            )}
          >
            {processing ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                处理中...
              </>
            ) : (
              '开始处理'
            )}
          </button>
        </div>
      )}

      {/* Results */}
      {results.length > 0 && (
        <div className="card p-6">
          <h2 className="text-lg font-semibold mb-4">处理结果</h2>
          
          <div className="space-y-4">
            {results.map((result, index) => (
              <div key={index} className="border border-border rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    {result.success ? (
                      <CheckCircle className="h-5 w-5 text-green-500" />
                    ) : (
                      <AlertCircle className="h-5 w-5 text-red-500" />
                    )}
                    <h3 className="font-medium text-foreground">
                      {operations.find(op => op.id === result.operation)?.label || result.operation}
                    </h3>
                  </div>
                  
                  {result.success && (
                    <button
                      onClick={() => saveResult(result)}
                      className="btn-ghost text-sm"
                    >
                      <Download className="h-4 w-4 mr-1" />
                      保存
                    </button>
                  )}
                </div>
                
                <div className="bg-secondary/20 rounded p-3">
                  <pre className="whitespace-pre-wrap text-sm text-foreground">
                    {result.result}
                  </pre>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}