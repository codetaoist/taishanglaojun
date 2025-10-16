#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
太上老君AI平台 - 文档验证脚本

用于验证文档的完整性、格式规范和内容质量
"""

import os
import re
import sys
import json
import yaml
import argparse
from pathlib import Path
from typing import List, Dict, Tuple, Optional
from dataclasses import dataclass
from datetime import datetime

@dataclass
class ValidationResult:
    """验证结果数据类"""
    file_path: str
    status: str  # 'pass', 'warning', 'error'
    message: str
    line_number: Optional[int] = None
    suggestion: Optional[str] = None

class DocumentValidator:
    """文档验证器"""
    
    def __init__(self, root_dir: str = "."):
        self.root_dir = Path(root_dir)
        self.results: List[ValidationResult] = []
        self.templates_dir = self.root_dir / "docs" / "templates"
        
        # 必需的章节模式
        self.required_sections = {
            'README.md': [
                r'# .+',  # 标题
                r'## 核心功能',
                r'## 架构设计',
                r'## 快速开始',
                r'## 使用说明'
            ],
            'API文档': [
                r'# .+ API 文档',
                r'## 📋 API概述',
                r'## 🔌 API接口列表',
                r'## 📊 错误码定义'
            ],
            '功能设计': [
                r'# .+ 功能设计',
                r'## 📋 文档信息',
                r'## 🎯 需求概述',
                r'## 📊 需求分析',
                r'## 🏗️ 系统设计'
            ]
        }
        
        # 链接模式
        self.link_patterns = [
            r'\[([^\]]+)\]\(([^)]+)\)',  # Markdown链接
            r'<([^>]+)>',  # 尖括号链接
            r'https?://[^\s]+',  # HTTP链接
        ]
        
    def validate_all(self) -> Dict[str, any]:
        """验证所有文档"""
        print("🔍 开始文档验证...")
        
        # 验证文档结构
        self._validate_structure()
        
        # 验证README文件
        self._validate_readme_files()
        
        # 验证API文档
        self._validate_api_docs()
        
        # 验证模板使用
        self._validate_template_usage()
        
        # 验证链接有效性
        self._validate_links()
        
        # 验证文档元数据
        self._validate_metadata()
        
        # 生成报告
        return self._generate_report()
    
    def _validate_structure(self):
        """验证文档目录结构"""
        print("📁 验证文档目录结构...")
        
        required_dirs = [
            "docs/00-项目概览",
            "docs/01-快速开始", 
            "docs/02-架构设计",
            "docs/03-核心服务",
            "docs/04-前端应用",
            "docs/05-基础设施",
            "docs/06-API文档",
            "docs/07-开发指南",
            "docs/08-部署运维",
            "docs/09-用户手册",
            "docs/10-开发进度",
            "docs/templates"
        ]
        
        for dir_path in required_dirs:
            full_path = self.root_dir / dir_path
            if not full_path.exists():
                self.results.append(ValidationResult(
                    file_path=dir_path,
                    status='error',
                    message=f'缺少必需的目录: {dir_path}',
                    suggestion=f'创建目录: mkdir -p {dir_path}'
                ))
            else:
                self.results.append(ValidationResult(
                    file_path=dir_path,
                    status='pass',
                    message='目录结构正确'
                ))
    
    def _validate_readme_files(self):
        """验证README文件"""
        print("📝 验证README文件...")
        
        # 查找所有README文件
        readme_files = list(self.root_dir.rglob("README.md"))
        
        for readme_file in readme_files:
            # 跳过模板目录
            if 'templates' in str(readme_file):
                continue
                
            self._validate_readme_content(readme_file)
    
    def _validate_readme_content(self, file_path: Path):
        """验证README文件内容"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # 检查必需章节
            missing_sections = []
            for pattern in self.required_sections['README.md']:
                if not re.search(pattern, content, re.MULTILINE):
                    missing_sections.append(pattern)
            
            if missing_sections:
                self.results.append(ValidationResult(
                    file_path=str(file_path.relative_to(self.root_dir)),
                    status='warning',
                    message=f'缺少推荐章节: {", ".join(missing_sections)}',
                    suggestion='参考README模板添加缺失章节'
                ))
            else:
                self.results.append(ValidationResult(
                    file_path=str(file_path.relative_to(self.root_dir)),
                    status='pass',
                    message='README结构完整'
                ))
            
            # 检查内容质量
            self._check_content_quality(file_path, content)
            
        except Exception as e:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='error',
                message=f'读取文件失败: {str(e)}'
            ))
    
    def _validate_api_docs(self):
        """验证API文档"""
        print("🔌 验证API文档...")
        
        api_docs_dir = self.root_dir / "docs" / "06-API文档"
        if not api_docs_dir.exists():
            return
        
        api_files = list(api_docs_dir.rglob("*.md"))
        
        for api_file in api_files:
            self._validate_api_content(api_file)
    
    def _validate_api_content(self, file_path: Path):
        """验证API文档内容"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # 检查API文档必需章节
            missing_sections = []
            for pattern in self.required_sections['API文档']:
                if not re.search(pattern, content, re.MULTILINE):
                    missing_sections.append(pattern)
            
            if missing_sections:
                self.results.append(ValidationResult(
                    file_path=str(file_path.relative_to(self.root_dir)),
                    status='warning',
                    message=f'API文档缺少章节: {", ".join(missing_sections)}',
                    suggestion='参考API文档模板添加缺失章节'
                ))
            
            # 检查API接口格式
            self._check_api_format(file_path, content)
            
        except Exception as e:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='error',
                message=f'读取API文档失败: {str(e)}'
            ))
    
    def _check_api_format(self, file_path: Path, content: str):
        """检查API接口格式"""
        # 检查HTTP方法格式
        http_methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']
        api_patterns = []
        
        for method in http_methods:
            pattern = rf'```http\s*\n{method}\s+/[^\n]*\n```'
            if re.search(pattern, content, re.MULTILINE | re.IGNORECASE):
                api_patterns.append(method)
        
        if api_patterns:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='pass',
                message=f'发现API接口: {", ".join(api_patterns)}'
            ))
        else:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='warning',
                message='未发现标准格式的API接口定义',
                suggestion='使用标准的HTTP代码块格式定义API'
            ))
    
    def _validate_template_usage(self):
        """验证模板使用情况"""
        print("📋 验证模板使用情况...")
        
        if not self.templates_dir.exists():
            self.results.append(ValidationResult(
                file_path="docs/templates",
                status='error',
                message='模板目录不存在'
            ))
            return
        
        templates = list(self.templates_dir.glob("*.md"))
        
        for template in templates:
            template_name = template.stem
            self.results.append(ValidationResult(
                file_path=str(template.relative_to(self.root_dir)),
                status='pass',
                message=f'模板可用: {template_name}'
            ))
    
    def _validate_links(self):
        """验证文档链接"""
        print("🔗 验证文档链接...")
        
        md_files = list(self.root_dir.rglob("*.md"))
        
        for md_file in md_files:
            if 'templates' in str(md_file):
                continue
            self._check_file_links(md_file)
    
    def _check_file_links(self, file_path: Path):
        """检查文件中的链接"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            lines = content.split('\n')
            
            for line_num, line in enumerate(lines, 1):
                # 检查Markdown链接
                markdown_links = re.findall(r'\[([^\]]+)\]\(([^)]+)\)', line)
                
                for link_text, link_url in markdown_links:
                    self._validate_link(file_path, line_num, link_text, link_url)
                    
        except Exception as e:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='error',
                message=f'检查链接失败: {str(e)}'
            ))
    
    def _validate_link(self, file_path: Path, line_num: int, link_text: str, link_url: str):
        """验证单个链接"""
        # 跳过外部链接和锚点链接
        if link_url.startswith(('http://', 'https://', '#', 'mailto:')):
            return
        
        # 处理相对路径
        if link_url.startswith('./') or link_url.startswith('../'):
            target_path = (file_path.parent / link_url).resolve()
        else:
            target_path = self.root_dir / link_url
        
        if not target_path.exists():
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='error',
                message=f'链接目标不存在: {link_url}',
                line_number=line_num,
                suggestion=f'检查链接路径或创建目标文件: {target_path}'
            ))
    
    def _validate_metadata(self):
        """验证文档元数据"""
        print("📊 验证文档元数据...")
        
        md_files = list(self.root_dir.rglob("*.md"))
        
        for md_file in md_files:
            if 'templates' in str(md_file):
                continue
            self._check_metadata(md_file)
    
    def _check_metadata(self, file_path: Path):
        """检查文档元数据"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # 检查是否有标题
            if not re.search(r'^# .+', content, re.MULTILINE):
                self.results.append(ValidationResult(
                    file_path=str(file_path.relative_to(self.root_dir)),
                    status='warning',
                    message='文档缺少主标题',
                    suggestion='添加一级标题 (# 标题)'
                ))
            
            # 检查是否有状态徽章
            badge_pattern = r'\[\!\[.*?\]\(.*?\)\]\(.*?\)'
            if not re.search(badge_pattern, content):
                self.results.append(ValidationResult(
                    file_path=str(file_path.relative_to(self.root_dir)),
                    status='info',
                    message='建议添加状态徽章',
                    suggestion='参考模板添加状态徽章'
                ))
                
        except Exception as e:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='error',
                message=f'检查元数据失败: {str(e)}'
            ))
    
    def _check_content_quality(self, file_path: Path, content: str):
        """检查内容质量"""
        # 检查内容长度
        if len(content.strip()) < 100:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='warning',
                message='文档内容过短，可能需要补充',
                suggestion='添加更多详细信息和说明'
            ))
        
        # 检查代码块
        code_blocks = re.findall(r'```(\w+)?\n(.*?)\n```', content, re.DOTALL)
        if code_blocks:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='pass',
                message=f'包含 {len(code_blocks)} 个代码示例'
            ))
        
        # 检查图片
        images = re.findall(r'!\[([^\]]*)\]\(([^)]+)\)', content)
        if images:
            self.results.append(ValidationResult(
                file_path=str(file_path.relative_to(self.root_dir)),
                status='pass',
                message=f'包含 {len(images)} 个图片'
            ))
    
    def _generate_report(self) -> Dict[str, any]:
        """生成验证报告"""
        print("📋 生成验证报告...")
        
        # 统计结果
        stats = {
            'total': len(self.results),
            'pass': len([r for r in self.results if r.status == 'pass']),
            'warning': len([r for r in self.results if r.status == 'warning']),
            'error': len([r for r in self.results if r.status == 'error']),
            'info': len([r for r in self.results if r.status == 'info'])
        }
        
        # 按状态分组
        grouped_results = {
            'pass': [r for r in self.results if r.status == 'pass'],
            'warning': [r for r in self.results if r.status == 'warning'],
            'error': [r for r in self.results if r.status == 'error'],
            'info': [r for r in self.results if r.status == 'info']
        }
        
        # 生成报告
        report = {
            'timestamp': datetime.now().isoformat(),
            'summary': stats,
            'results': grouped_results,
            'recommendations': self._generate_recommendations(stats, grouped_results)
        }
        
        return report
    
    def _generate_recommendations(self, stats: Dict, grouped_results: Dict) -> List[str]:
        """生成改进建议"""
        recommendations = []
        
        if stats['error'] > 0:
            recommendations.append(f"🔴 发现 {stats['error']} 个错误，需要立即修复")
        
        if stats['warning'] > 0:
            recommendations.append(f"🟡 发现 {stats['warning']} 个警告，建议优化")
        
        if stats['error'] == 0 and stats['warning'] == 0:
            recommendations.append("✅ 文档质量良好，无需修复")
        
        # 具体建议
        error_files = set(r.file_path for r in grouped_results['error'])
        if error_files:
            recommendations.append(f"优先修复以下文件: {', '.join(list(error_files)[:5])}")
        
        warning_files = set(r.file_path for r in grouped_results['warning'])
        if warning_files:
            recommendations.append(f"建议优化以下文件: {', '.join(list(warning_files)[:5])}")
        
        return recommendations

def print_report(report: Dict):
    """打印验证报告"""
    print("\n" + "="*60)
    print("📋 文档验证报告")
    print("="*60)
    
    # 打印统计信息
    stats = report['summary']
    print(f"\n📊 统计信息:")
    print(f"  总计: {stats['total']}")
    print(f"  ✅ 通过: {stats['pass']}")
    print(f"  🟡 警告: {stats['warning']}")
    print(f"  🔴 错误: {stats['error']}")
    print(f"  ℹ️  信息: {stats['info']}")
    
    # 打印错误
    if report['results']['error']:
        print(f"\n🔴 错误 ({len(report['results']['error'])}):")
        for result in report['results']['error'][:10]:  # 只显示前10个
            line_info = f" (行 {result.line_number})" if result.line_number else ""
            print(f"  ❌ {result.file_path}{line_info}: {result.message}")
            if result.suggestion:
                print(f"     💡 建议: {result.suggestion}")
    
    # 打印警告
    if report['results']['warning']:
        print(f"\n🟡 警告 ({len(report['results']['warning'])}):")
        for result in report['results']['warning'][:10]:  # 只显示前10个
            line_info = f" (行 {result.line_number})" if result.line_number else ""
            print(f"  ⚠️  {result.file_path}{line_info}: {result.message}")
            if result.suggestion:
                print(f"     💡 建议: {result.suggestion}")
    
    # 打印建议
    if report['recommendations']:
        print(f"\n💡 改进建议:")
        for rec in report['recommendations']:
            print(f"  • {rec}")
    
    print(f"\n⏰ 验证时间: {report['timestamp']}")
    print("="*60)

def save_report(report: Dict, output_file: str):
    """保存验证报告"""
    try:
        # 转换结果为可序列化格式
        serializable_report = {
            'timestamp': report['timestamp'],
            'summary': report['summary'],
            'recommendations': report['recommendations'],
            'results': {}
        }
        
        for status, results in report['results'].items():
            serializable_report['results'][status] = [
                {
                    'file_path': r.file_path,
                    'status': r.status,
                    'message': r.message,
                    'line_number': r.line_number,
                    'suggestion': r.suggestion
                }
                for r in results
            ]
        
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(serializable_report, f, ensure_ascii=False, indent=2)
        
        print(f"📄 报告已保存到: {output_file}")
        
    except Exception as e:
        print(f"❌ 保存报告失败: {str(e)}")

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='太上老君AI平台文档验证工具')
    parser.add_argument('--root', '-r', default='.', help='项目根目录路径')
    parser.add_argument('--output', '-o', help='输出报告文件路径')
    parser.add_argument('--format', '-f', choices=['text', 'json'], default='text', help='输出格式')
    parser.add_argument('--verbose', '-v', action='store_true', help='详细输出')
    
    args = parser.parse_args()
    
    # 创建验证器
    validator = DocumentValidator(args.root)
    
    # 执行验证
    report = validator.validate_all()
    
    # 输出报告
    if args.format == 'text':
        print_report(report)
    
    # 保存报告
    if args.output:
        save_report(report, args.output)
    
    # 返回退出码
    if report['summary']['error'] > 0:
        sys.exit(1)
    elif report['summary']['warning'] > 0:
        sys.exit(2)
    else:
        sys.exit(0)

if __name__ == '__main__':
    main()