#!/usr/bin/env python3
"""
Script to generate Python gRPC code from proto files
"""

import os
import sys
import subprocess
import importlib.util

def check_protoc():
    """Check if protoc is installed"""
    try:
        result = subprocess.run(['protoc', '--version'], capture_output=True, text=True)
        if result.returncode == 0:
            print(f"protoc is installed: {result.stdout.strip()}")
            return True
        else:
            print("protoc is not installed or not in PATH")
            return False
    except FileNotFoundError:
        print("protoc is not found")
        return False

def check_grpcio_tools():
    """Check if grpcio-tools is installed"""
    try:
        import grpc_tools.protoc
        print("grpcio-tools is installed")
        return True
    except ImportError:
        print("grpcio-tools is not installed")
        return False

def generate_python_grpc():
    """Generate Python gRPC code"""
    # Set paths
    proto_dir = "/Users/lida/Documents/work/codetaoist/api/proto"
    python_out_dir = "/Users/lida/Documents/work/codetaoist/services/ai"
    proto_file = "ai_service.proto"
    
    # Create proto directory in the Python service if it doesn't exist
    proto_service_dir = os.path.join(python_out_dir, "proto")
    os.makedirs(proto_service_dir, exist_ok=True)
    
    # Generate Python gRPC code
    print("Generating Python gRPC code...")
    cmd = [
        'python3', '-m', 'grpc_tools.protoc',
        f'--proto_path={proto_dir}',
        f'--python_out={python_out_dir}',
        f'--grpc_python_out={python_out_dir}',
        os.path.join(proto_dir, proto_file)
    ]
    
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
        print("Python gRPC code generated successfully!")
        
        # Check if files were created
        pb2_file = os.path.join(python_out_dir, "ai_service_pb2.py")
        pb2_grpc_file = os.path.join(python_out_dir, "ai_service_pb2_grpc.py")
        
        if os.path.exists(pb2_file) and os.path.exists(pb2_grpc_file):
            print(f"Generated files:")
            print(f"- {pb2_file}")
            print(f"- {pb2_grpc_file}")
            
            # Create __init__.py files if they don't exist
            if not os.path.exists(os.path.join(python_out_dir, "__init__.py")):
                with open(os.path.join(python_out_dir, "__init__.py"), "w") as f:
                    f.write("# Package initialization\n")
                print(f"Created __init__.py in {python_out_dir}")
            
            return True
        else:
            print("Error: Generated files not found")
            return False
    except subprocess.CalledProcessError as e:
        print(f"Error generating Python gRPC code: {e}")
        print(f"stdout: {e.stdout}")
        print(f"stderr: {e.stderr}")
        return False

def update_python_service():
    """Update the Python service to use generated gRPC code"""
    service_file = "/Users/lida/Documents/work/codetaoist/services/ai/ai_grpc_services.py"
    
    # Read the current service file
    with open(service_file, 'r') as f:
        content = f.read()
    
    # Update imports
    content = content.replace(
        "# TODO: Uncomment after generating the protobuf code\n# import app.proto.ai_service_pb2 as ai_service_pb2\n# import app.proto.ai_service_pb2_grpc as ai_service_pb2_grpc",
        "import ai_service_pb2 as ai_service_pb2\nimport ai_service_pb2_grpc as ai_service_pb2_grpc"
    )
    
    # Update class inheritance
    content = content.replace(
        "# TODO: Uncomment after generating the protobuf code\n# class VectorServiceImpl(ai_service_pb2_grpc.VectorServiceServicer):",
        "class VectorServiceImpl(ai_service_pb2_grpc.VectorServiceServicer):"
    )
    
    content = content.replace(
        "# TODO: Uncomment after generating the protobuf code\n# class ModelServiceImpl(ai_service_pb2_grpc.ModelServiceServicer):",
        "class ModelServiceImpl(ai_service_pb2_grpc.ModelServiceServicer):"
    )
    
    # Update service registration
    content = content.replace(
        "# TODO: Add vector service after generating the protobuf code\n# ai_service_pb2_grpc.add_VectorServiceServicer_to_server(VectorServiceImpl(), server)",
        "ai_service_pb2_grpc.add_VectorServiceServicer_to_server(VectorServiceImpl(), server)"
    )
    
    content = content.replace(
        "# TODO: Add model service after generating the protobuf code\n# ai_service_pb2_grpc.add_ModelServiceServicer_to_server(ModelServiceImpl(), server)",
        "ai_service_pb2_grpc.add_ModelServiceServicer_to_server(ModelServiceImpl(), server)"
    )
    
    # Write the updated content back
    with open(service_file, 'w') as f:
        f.write(content)
    
    print(f"Updated {service_file} to use generated gRPC code")

def main():
    """Main function"""
    print("Python gRPC Code Generation Script")
    print("=" * 40)
    
    # Check dependencies
    if not check_protoc():
        print("Error: protoc is not installed. Please install protobuf first.")
        sys.exit(1)
    
    if not check_grpcio_tools():
        print("Error: grpcio-tools is not installed. Please install it first.")
        sys.exit(1)
    
    # Generate Python gRPC code
    if not generate_python_grpc():
        print("Error: Failed to generate Python gRPC code.")
        sys.exit(1)
    
    # Update Python service
    update_python_service()
    
    print("\nPython gRPC code generation completed successfully!")

if __name__ == "__main__":
    main()