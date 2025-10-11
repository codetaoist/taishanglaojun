"""
太上老君AI平台 Python SDK 安装配置
"""

from setuptools import setup, find_packages
import os

# 读取README文件
def read_readme():
    readme_path = os.path.join(os.path.dirname(__file__), 'README.md')
    if os.path.exists(readme_path):
        with open(readme_path, 'r', encoding='utf-8') as f:
            return f.read()
    return ''

# 读取requirements文件
def read_requirements():
    requirements_path = os.path.join(os.path.dirname(__file__), 'requirements.txt')
    if os.path.exists(requirements_path):
        with open(requirements_path, 'r', encoding='utf-8') as f:
            return [line.strip() for line in f if line.strip() and not line.startswith('#')]
    return []

setup(
    name='taishanglaojun-sdk',
    version='1.0.0',
    description='太上老君AI平台官方Python SDK',
    long_description=read_readme(),
    long_description_content_type='text/markdown',
    author='太上老君AI团队',
    author_email='dev@taishanglaojun.com',
    url='https://github.com/taishanglaojun/python-sdk',
    project_urls={
        'Documentation': 'https://docs.taishanglaojun.com/sdk/python',
        'Source': 'https://github.com/taishanglaojun/python-sdk',
        'Tracker': 'https://github.com/taishanglaojun/python-sdk/issues',
        'Homepage': 'https://taishanglaojun.com',
    },
    packages=find_packages(exclude=['tests*', 'examples*']),
    include_package_data=True,
    package_data={
        'taishanglaojun': ['py.typed'],
    },
    python_requires='>=3.7',
    install_requires=[
        'requests>=2.25.0',
        'aiohttp>=3.8.0',
        'pydantic>=1.8.0',
        'typing-extensions>=4.0.0',
        'websockets>=10.0',
    ],
    extras_require={
        'dev': [
            'pytest>=7.0.0',
            'pytest-asyncio>=0.21.0',
            'pytest-cov>=4.0.0',
            'black>=22.0.0',
            'isort>=5.10.0',
            'flake8>=5.0.0',
            'mypy>=1.0.0',
            'pre-commit>=2.20.0',
        ],
        'docs': [
            'sphinx>=5.0.0',
            'sphinx-rtd-theme>=1.2.0',
            'sphinx-autodoc-typehints>=1.19.0',
        ],
        'examples': [
            'flask>=2.2.0',
            'fastapi>=0.95.0',
            'uvicorn>=0.20.0',
            'streamlit>=1.25.0',
        ],
    },
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: 3.10',
        'Programming Language :: Python :: 3.11',
        'Programming Language :: Python :: 3.12',
        'Topic :: Software Development :: Libraries :: Python Modules',
        'Topic :: Scientific/Engineering :: Artificial Intelligence',
        'Topic :: Internet :: WWW/HTTP :: Dynamic Content',
        'Topic :: Communications :: Chat',
    ],
    keywords=[
        'taishanglaojun',
        'ai',
        'artificial-intelligence',
        'chatbot',
        'nlp',
        'machine-learning',
        'sdk',
        'api',
        'python',
        'async',
    ],
    license='MIT',
    zip_safe=False,
    entry_points={
        'console_scripts': [
            'taishanglaojun=taishanglaojun.cli:main',
        ],
    },
)