#!/bin/bash

# AI Services Startup Script

# Set paths
PYTHON_SERVICE_DIR="/Users/lida/Documents/work/codetaoist/services/ai"
GO_SERVICE_DIR="/Users/lida/Documents/work/codetaoist/api"

# Function to check if a port is in use
function is_port_in_use() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null ; then
        return 0
    else
        return 1
    fi
}

# Function to start Go gRPC service
function start_go_service() {
    echo "Starting Go gRPC service..."
    cd $GO_SERVICE_DIR
    go run ai_grpc_services.go &
    GO_PID=$!
    echo "Go gRPC service started with PID: $GO_PID"
    echo $GO_PID > /tmp/go_grpc_service.pid
}

# Function to start Python vector service
function start_python_vector_service() {
    echo "Starting Python vector service..."
    cd $PYTHON_SERVICE_DIR
    python3 ai_grpc_services.py vector &
    VECTOR_PID=$!
    echo "Python vector service started with PID: $VECTOR_PID"
    echo $VECTOR_PID > /tmp/python_vector_service.pid
}

# Function to start Python model service
function start_python_model_service() {
    echo "Starting Python model service..."
    cd $PYTHON_SERVICE_DIR
    python3 ai_grpc_services.py model &
    MODEL_PID=$!
    echo "Python model service started with PID: $MODEL_PID"
    echo $MODEL_PID > /tmp/python_model_service.pid
}

# Function to stop services
function stop_services() {
    echo "Stopping services..."
    
    # Stop Go service
    if [ -f /tmp/go_grpc_service.pid ]; then
        GO_PID=$(cat /tmp/go_grpc_service.pid)
        if kill -0 $GO_PID 2>/dev/null; then
            echo "Stopping Go gRPC service (PID: $GO_PID)..."
            kill $GO_PID
        fi
        rm -f /tmp/go_grpc_service.pid
    fi
    
    # Stop Python vector service
    if [ -f /tmp/python_vector_service.pid ]; then
        VECTOR_PID=$(cat /tmp/python_vector_service.pid)
        if kill -0 $VECTOR_PID 2>/dev/null; then
            echo "Stopping Python vector service (PID: $VECTOR_PID)..."
            kill $VECTOR_PID
        fi
        rm -f /tmp/python_vector_service.pid
    fi
    
    # Stop Python model service
    if [ -f /tmp/python_model_service.pid ]; then
        MODEL_PID=$(cat /tmp/python_model_service.pid)
        if kill -0 $MODEL_PID 2>/dev/null; then
            echo "Stopping Python model service (PID: $MODEL_PID)..."
            kill $MODEL_PID
        fi
        rm -f /tmp/python_model_service.pid
    fi
    
    echo "All services stopped."
}

# Function to check service status
function check_status() {
    echo "Checking service status..."
    
    # Check Go service
    if [ -f /tmp/go_grpc_service.pid ]; then
        GO_PID=$(cat /tmp/go_grpc_service.pid)
        if kill -0 $GO_PID 2>/dev/null; then
            echo "Go gRPC service is running (PID: $GO_PID)"
        else
            echo "Go gRPC service is not running"
        fi
    else
        echo "Go gRPC service is not running"
    fi
    
    # Check Python vector service
    if [ -f /tmp/python_vector_service.pid ]; then
        VECTOR_PID=$(cat /tmp/python_vector_service.pid)
        if kill -0 $VECTOR_PID 2>/dev/null; then
            echo "Python vector service is running (PID: $VECTOR_PID)"
        else
            echo "Python vector service is not running"
        fi
    else
        echo "Python vector service is not running"
    fi
    
    # Check Python model service
    if [ -f /tmp/python_model_service.pid ]; then
        MODEL_PID=$(cat /tmp/python_model_service.pid)
        if kill -0 $MODEL_PID 2>/dev/null; then
            echo "Python model service is running (PID: $MODEL_PID)"
        else
            echo "Python model service is not running"
        fi
    else
        echo "Python model service is not running"
    fi
}

# Main script logic
case "$1" in
    start)
        echo "Starting AI services..."
        
        # Check if ports are already in use
        if is_port_in_use 50051; then
            echo "Port 50051 is already in use. Please stop existing services first."
            exit 1
        fi
        
        if is_port_in_use 50052; then
            echo "Port 50052 is already in use. Please stop existing services first."
            exit 1
        fi
        
        # Start services
        start_go_service
        sleep 2  # Give Go service time to start
        
        start_python_vector_service
        start_python_model_service
        
        echo "All services started."
        echo "Go gRPC service is running on port 50051"
        echo "Python vector service is running on port 50051"
        echo "Python model service is running on port 50052"
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        sleep 2
        start_go_service
        sleep 2
        start_python_vector_service
        start_python_model_service
        echo "All services restarted."
        ;;
    status)
        check_status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac

exit 0