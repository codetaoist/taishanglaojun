#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('🔍 开始分析打包体积...\n');

// 构建项目
console.log('📦 构建项目...');
try {
  execSync('npm run build', { stdio: 'inherit' });
} catch (error) {
  console.error('❌ 构建失败:', error.message);
  process.exit(1);
}

// 分析dist目录
const distPath = path.join(__dirname, '../dist');
const assetsPath = path.join(distPath, 'assets');

if (!fs.existsSync(distPath)) {
  console.error('❌ dist目录不存在');
  process.exit(1);
}

// 获取文件大小
function getFileSize(filePath) {
  const stats = fs.statSync(filePath);
  return stats.size;
}

// 格式化文件大小
function formatSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 分析文件
function analyzeFiles(dirPath, prefix = '') {
  const files = fs.readdirSync(dirPath);
  const results = [];
  
  files.forEach(file => {
    const filePath = path.join(dirPath, file);
    const stat = fs.statSync(filePath);
    
    if (stat.isDirectory()) {
      results.push(...analyzeFiles(filePath, prefix + file + '/'));
    } else {
      const size = getFileSize(filePath);
      results.push({
        name: prefix + file,
        size: size,
        formattedSize: formatSize(size),
        type: path.extname(file)
      });
    }
  });
  
  return results;
}

// 分析结果
const files = analyzeFiles(distPath);

// 按类型分组
const filesByType = files.reduce((acc, file) => {
  const type = file.type || 'other';
  if (!acc[type]) acc[type] = [];
  acc[type].push(file);
  return acc;
}, {});

// 计算总大小
const totalSize = files.reduce((sum, file) => sum + file.size, 0);

console.log('\n📊 打包分析结果:');
console.log('='.repeat(50));
console.log(`总体积: ${formatSize(totalSize)}`);
console.log('');

// 按类型显示
Object.entries(filesByType).forEach(([type, typeFiles]) => {
  const typeSize = typeFiles.reduce((sum, file) => sum + file.size, 0);
  console.log(`${type.toUpperCase()} 文件 (${formatSize(typeSize)}):`);
  
  // 按大小排序
  typeFiles.sort((a, b) => b.size - a.size);
  
  typeFiles.forEach(file => {
    console.log(`  ${file.name.padEnd(40)} ${file.formattedSize.padStart(10)}`);
  });
  console.log('');
});

// 显示最大的文件
console.log('🔍 最大的10个文件:');
console.log('-'.repeat(50));
files.sort((a, b) => b.size - a.size);
files.slice(0, 10).forEach((file, index) => {
  console.log(`${(index + 1).toString().padStart(2)}. ${file.name.padEnd(35)} ${file.formattedSize.padStart(10)}`);
});

// 检查是否有过大的文件
const largeFiles = files.filter(file => file.size > 500 * 1024); // 500KB
if (largeFiles.length > 0) {
  console.log('\n⚠️  发现过大的文件 (>500KB):');
  largeFiles.forEach(file => {
    console.log(`  ${file.name} - ${file.formattedSize}`);
  });
  console.log('\n💡 建议:');
  console.log('  - 考虑代码分割');
  console.log('  - 使用动态导入');
  console.log('  - 移除未使用的依赖');
}

// 生成报告文件
const report = {
  timestamp: new Date().toISOString(),
  totalSize: totalSize,
  totalSizeFormatted: formatSize(totalSize),
  filesByType: filesByType,
  largestFiles: files.slice(0, 10),
  largeFiles: largeFiles
};

fs.writeFileSync(
  path.join(__dirname, '../bundle-analysis.json'),
  JSON.stringify(report, null, 2)
);

console.log('\n✅ 分析完成! 详细报告已保存到 bundle-analysis.json');