FROM python:3.11-slim

WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .

# 安装Python依赖
RUN pip install --no-cache-dir -r requirements.txt

# 复制应用代码
COPY . .

# 生成gRPC代码
RUN python -m grpc_tools.protoc --python_out=. --grpc_python_out=. --proto_path=app/proto app/proto/ai_service.proto

# 创建非root用户
RUN useradd --create-home --shell /bin/bash app \
    && chown -R app:app /app
USER app

# 暴露端口
EXPOSE 8083 50051 50052 8000

# 启动命令
CMD ["python", "grpc_server.py"]