#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('🔍 开始分析项目依赖...\n');

// 读取package.json
const packageJsonPath = path.join(__dirname, '../package.json');
const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

const dependencies = packageJson.dependencies || {};
const devDependencies = packageJson.devDependencies || {};
const allDeps = { ...dependencies, ...devDependencies };

console.log(`📦 总依赖数量: ${Object.keys(allDeps).length}`);
console.log(`   - 生产依赖: ${Object.keys(dependencies).length}`);
console.log(`   - 开发依赖: ${Object.keys(devDependencies).length}\n`);

// 分析依赖大小
function analyzeDependencySize() {
  console.log('📊 分析依赖大小...');
  
  const nodeModulesPath = path.join(__dirname, '../node_modules');
  if (!fs.existsSync(nodeModulesPath)) {
    console.log('❌ node_modules 目录不存在，请先运行 npm install');
    return;
  }
  
  const depSizes = [];
  
  Object.keys(allDeps).forEach(dep => {
    const depPath = path.join(nodeModulesPath, dep);
    if (fs.existsSync(depPath)) {
      try {
        const size = getDirSize(depPath);
        depSizes.push({
          name: dep,
          size: size,
          formattedSize: formatSize(size),
          type: dependencies[dep] ? 'production' : 'development'
        });
      } catch (error) {
        console.warn(`⚠️  无法分析 ${dep} 的大小:`, error.message);
      }
    }
  });
  
  // 按大小排序
  depSizes.sort((a, b) => b.size - a.size);
  
  console.log('\n🔍 最大的10个依赖:');
  console.log('-'.repeat(60));
  depSizes.slice(0, 10).forEach((dep, index) => {
    const typeIcon = dep.type === 'production' ? '🟢' : '🔵';
    console.log(`${(index + 1).toString().padStart(2)}. ${typeIcon} ${dep.name.padEnd(30)} ${dep.formattedSize.padStart(10)}`);
  });
  
  // 分析大型依赖
  const largeDeps = depSizes.filter(dep => dep.size > 5 * 1024 * 1024); // 5MB
  if (largeDeps.length > 0) {
    console.log('\n⚠️  发现大型依赖 (>5MB):');
    largeDeps.forEach(dep => {
      console.log(`  ${dep.name} - ${dep.formattedSize} (${dep.type})`);
    });
  }
  
  return depSizes;
}

// 检查未使用的依赖
function checkUnusedDependencies() {
  console.log('\n🔍 检查可能未使用的依赖...');
  
  const srcPath = path.join(__dirname, '../src');
  const configFiles = [
    path.join(__dirname, '../vite.config.ts'),
    path.join(__dirname, '../tailwind.config.js'),
    path.join(__dirname, '../postcss.config.js'),
  ];
  
  // 读取所有源文件
  const sourceFiles = getAllFiles(srcPath, ['.ts', '.tsx', '.js', '.jsx']);
  const configContent = configFiles
    .filter(file => fs.existsSync(file))
    .map(file => fs.readFileSync(file, 'utf8'))
    .join('\n');
  
  const allContent = sourceFiles.map(file => {
    try {
      return fs.readFileSync(file, 'utf8');
    } catch (error) {
      return '';
    }
  }).join('\n') + configContent;
  
  const possiblyUnused = [];
  
  Object.keys(allDeps).forEach(dep => {
    // 跳过一些特殊的依赖
    const skipDeps = [
      '@types/',
      'eslint',
      'typescript',
      'vite',
      'terser',
      'autoprefixer',
      'postcss',
      'tailwindcss'
    ];
    
    if (skipDeps.some(skip => dep.includes(skip))) {
      return;
    }
    
    // 检查是否在代码中被引用
    const importPatterns = [
      new RegExp(`from ['"]${dep}['"]`, 'g'),
      new RegExp(`import ['"]${dep}['"]`, 'g'),
      new RegExp(`require\\(['"]${dep}['"]\\)`, 'g'),
      new RegExp(`'${dep}'`, 'g'),
      new RegExp(`"${dep}"`, 'g'),
    ];
    
    const isUsed = importPatterns.some(pattern => pattern.test(allContent));
    
    if (!isUsed) {
      possiblyUnused.push(dep);
    }
  });
  
  if (possiblyUnused.length > 0) {
    console.log('\n⚠️  可能未使用的依赖:');
    possiblyUnused.forEach(dep => {
      const type = dependencies[dep] ? '生产' : '开发';
      console.log(`  ${dep} (${type})`);
    });
    console.log('\n💡 建议手动确认后移除未使用的依赖');
  } else {
    console.log('✅ 未发现明显未使用的依赖');
  }
  
  return possiblyUnused;
}

// 检查重复依赖
function checkDuplicateDependencies() {
  console.log('\n🔍 检查重复功能的依赖...');
  
  const duplicateGroups = [
    {
      category: '日期处理',
      deps: ['dayjs', 'moment', 'date-fns'],
      recommendation: '建议只保留一个日期处理库，推荐 dayjs (体积最小)'
    },
    {
      category: 'HTTP客户端',
      deps: ['axios', 'fetch', 'node-fetch'],
      recommendation: '建议只使用 axios 或原生 fetch'
    },
    {
      category: '样式处理',
      deps: ['styled-components', 'emotion', '@emotion/react'],
      recommendation: '建议选择一个CSS-in-JS解决方案'
    }
  ];
  
  const foundDuplicates = [];
  
  duplicateGroups.forEach(group => {
    const foundDeps = group.deps.filter(dep => allDeps[dep]);
    if (foundDeps.length > 1) {
      foundDuplicates.push({
        ...group,
        foundDeps
      });
    }
  });
  
  if (foundDuplicates.length > 0) {
    console.log('\n⚠️  发现重复功能的依赖:');
    foundDuplicates.forEach(group => {
      console.log(`\n${group.category}:`);
      group.foundDeps.forEach(dep => {
        console.log(`  - ${dep}`);
      });
      console.log(`  💡 ${group.recommendation}`);
    });
  } else {
    console.log('✅ 未发现重复功能的依赖');
  }
  
  return foundDuplicates;
}

// 工具函数
function getDirSize(dirPath) {
  let size = 0;
  const files = fs.readdirSync(dirPath);
  
  files.forEach(file => {
    const filePath = path.join(dirPath, file);
    const stat = fs.statSync(filePath);
    
    if (stat.isDirectory()) {
      size += getDirSize(filePath);
    } else {
      size += stat.size;
    }
  });
  
  return size;
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function getAllFiles(dirPath, extensions) {
  let files = [];
  
  try {
    const items = fs.readdirSync(dirPath);
    
    items.forEach(item => {
      const itemPath = path.join(dirPath, item);
      const stat = fs.statSync(itemPath);
      
      if (stat.isDirectory() && !item.startsWith('.') && item !== 'node_modules') {
        files = files.concat(getAllFiles(itemPath, extensions));
      } else if (stat.isFile() && extensions.some(ext => item.endsWith(ext))) {
        files.push(itemPath);
      }
    });
  } catch (error) {
    console.warn(`⚠️  无法读取目录 ${dirPath}:`, error.message);
  }
  
  return files;
}

// 执行分析
async function runAnalysis() {
  const depSizes = analyzeDependencySize();
  const unusedDeps = checkUnusedDependencies();
  const duplicateDeps = checkDuplicateDependencies();
  
  // 生成报告
  const report = {
    timestamp: new Date().toISOString(),
    totalDependencies: Object.keys(allDeps).length,
    productionDependencies: Object.keys(dependencies).length,
    devDependencies: Object.keys(devDependencies).length,
    dependencySizes: depSizes,
    possiblyUnusedDependencies: unusedDeps,
    duplicateDependencies: duplicateDeps,
    recommendations: []
  };
  
  // 添加建议
  if (unusedDeps.length > 0) {
    report.recommendations.push('移除未使用的依赖以减少项目体积');
  }
  
  if (duplicateDeps.length > 0) {
    report.recommendations.push('合并重复功能的依赖');
  }
  
  const largeDeps = depSizes.filter(dep => dep.size > 5 * 1024 * 1024);
  if (largeDeps.length > 0) {
    report.recommendations.push('考虑替换大型依赖或使用按需导入');
  }
  
  // 保存报告
  fs.writeFileSync(
    path.join(__dirname, '../dependency-analysis.json'),
    JSON.stringify(report, null, 2)
  );
  
  console.log('\n✅ 依赖分析完成! 详细报告已保存到 dependency-analysis.json');
  
  if (report.recommendations.length > 0) {
    console.log('\n💡 优化建议:');
    report.recommendations.forEach((rec, index) => {
      console.log(`${index + 1}. ${rec}`);
    });
  }
}

runAnalysis().catch(console.error);