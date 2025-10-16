#!/bin/bash

# Taishang Laojun AI Platform Deployment Script
# 太上老君AI平台部署脚本

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT="${1:-staging}"
REGION="${2:-us-east-1}"
CLUSTER_NAME="taishanglaojun-${ENVIRONMENT}"
NAMESPACE="taishanglaojun-${ENVIRONMENT}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi
    
    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi
    
    if ! command -v aws &> /dev/null; then
        missing_tools+=("aws")
    fi
    
    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    log_success "All prerequisites are met"
}

# Configure AWS and Kubernetes
configure_aws() {
    log_info "Configuring AWS and Kubernetes access..."
    
    # Update kubeconfig
    aws eks update-kubeconfig --region "$REGION" --name "$CLUSTER_NAME"
    
    # Verify cluster access
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot access Kubernetes cluster"
        exit 1
    fi
    
    log_success "AWS and Kubernetes configured successfully"
}

# Create namespace and secrets
setup_namespace() {
    log_info "Setting up namespace and secrets..."
    
    # Apply namespace
    kubectl apply -f "$PROJECT_ROOT/k8s/namespace.yaml"
    
    # Create secrets if they don't exist
    if ! kubectl get secret app-secrets -n "$NAMESPACE" &> /dev/null; then
        log_warning "Creating placeholder secrets. Please update with actual values!"
        
        kubectl create secret generic app-secrets \
            --from-literal=database-url="postgres://user:pass@localhost:5432/db" \
            --from-literal=redis-url="redis://localhost:6379" \
            --from-literal=jwt-secret="your-jwt-secret-here" \
            --from-literal=openai-api-key="your-openai-key-here" \
            --from-literal=sentry-dsn="your-sentry-dsn-here" \
            -n "$NAMESPACE"
    fi
    
    log_success "Namespace and secrets configured"
}

# Deploy infrastructure components
deploy_infrastructure() {
    log_info "Deploying infrastructure components..."
    
    # Deploy cert-manager if not exists
    if ! kubectl get namespace cert-manager &> /dev/null; then
        log_info "Installing cert-manager..."
        kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
        kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s
    fi
    
    # Deploy ingress-nginx if not exists
    if ! kubectl get namespace ingress-nginx &> /dev/null; then
        log_info "Installing ingress-nginx..."
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/aws/deploy.yaml
        kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=300s
    fi
    
    # Deploy monitoring stack
    if [ -d "$PROJECT_ROOT/monitoring/helm" ]; then
        log_info "Deploying monitoring stack..."
        helm upgrade --install prometheus-stack \
            "$PROJECT_ROOT/monitoring/helm/prometheus-stack" \
            --namespace monitoring \
            --create-namespace \
            --values "$PROJECT_ROOT/monitoring/helm/values-${ENVIRONMENT}.yaml"
    fi
    
    log_success "Infrastructure components deployed"
}

# Build and push Docker images
build_and_push_images() {
    log_info "Building and pushing Docker images..."
    
    local registry="ghcr.io/taishanglaojun"
    local tag="${GITHUB_SHA:-$(git rev-parse HEAD)}"
    
    # Build frontend image
    log_info "Building frontend image..."
    docker build -t "${registry}/taishanglaojun-frontend:${tag}" \
        -f "$PROJECT_ROOT/docker/Dockerfile.frontend" \
        "$PROJECT_ROOT"
    
    # Build backend image
    log_info "Building backend image..."
    docker build -t "${registry}/taishanglaojun-backend:${tag}" \
        -f "$PROJECT_ROOT/core-services/Dockerfile" \
        "$PROJECT_ROOT/core-services"
    
    # Push images
    if [ "${PUSH_IMAGES:-true}" = "true" ]; then
        log_info "Pushing images to registry..."
        docker push "${registry}/taishanglaojun-frontend:${tag}"
        docker push "${registry}/taishanglaojun-backend:${tag}"
        
        # Update image tags in deployment files
        sed -i "s|image: ${registry}/taishanglaojun-frontend:.*|image: ${registry}/taishanglaojun-frontend:${tag}|g" \
            "$PROJECT_ROOT/k8s/${ENVIRONMENT}/frontend-deployment.yaml"
        sed -i "s|image: ${registry}/taishanglaojun-backend:.*|image: ${registry}/taishanglaojun-backend:${tag}|g" \
            "$PROJECT_ROOT/k8s/${ENVIRONMENT}/backend-deployment.yaml"
    fi
    
    log_success "Docker images built and pushed"
}

# Deploy application
deploy_application() {
    log_info "Deploying application to ${ENVIRONMENT}..."
    
    # Apply all Kubernetes manifests
    kubectl apply -f "$PROJECT_ROOT/k8s/${ENVIRONMENT}/"
    
    # Wait for deployments to be ready
    log_info "Waiting for deployments to be ready..."
    kubectl rollout status deployment/frontend -n "$NAMESPACE" --timeout=600s
    kubectl rollout status deployment/backend -n "$NAMESPACE" --timeout=600s
    
    log_success "Application deployed successfully"
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."
    
    # Get service URLs
    local frontend_url
    local backend_url
    
    if [ "$ENVIRONMENT" = "production" ]; then
        frontend_url="https://taishanglaojun.ai"
        backend_url="https://api.taishanglaojun.ai"
    else
        frontend_url="https://staging.taishanglaojun.ai"
        backend_url="https://api-staging.taishanglaojun.ai"
    fi
    
    # Wait for services to be available
    log_info "Waiting for services to be available..."
    sleep 30
    
    # Check frontend health
    if curl -f -s "${frontend_url}/health" > /dev/null; then
        log_success "Frontend health check passed"
    else
        log_warning "Frontend health check failed"
    fi
    
    # Check backend health
    if curl -f -s "${backend_url}/health" > /dev/null; then
        log_success "Backend health check passed"
    else
        log_warning "Backend health check failed"
    fi
    
    # Run smoke tests if available
    if [ -f "$PROJECT_ROOT/tests/e2e/package.json" ]; then
        log_info "Running smoke tests..."
        cd "$PROJECT_ROOT/tests/e2e"
        npm run test:smoke -- --base-url="$frontend_url" || log_warning "Smoke tests failed"
    fi
}

# Cleanup old resources
cleanup() {
    log_info "Cleaning up old resources..."
    
    # Remove old ReplicaSets
    kubectl delete replicaset -l app=taishanglaojun -n "$NAMESPACE" \
        --field-selector='status.replicas=0' || true
    
    # Clean up completed jobs older than 7 days
    kubectl delete job -l app=taishanglaojun -n "$NAMESPACE" \
        --field-selector='status.conditions[0].type=Complete' \
        --field-selector='metadata.creationTimestamp<$(date -d "7 days ago" -u +%Y-%m-%dT%H:%M:%SZ)' || true
    
    log_success "Cleanup completed"
}

# Rollback function
rollback() {
    log_warning "Rolling back deployment..."
    
    kubectl rollout undo deployment/frontend -n "$NAMESPACE"
    kubectl rollout undo deployment/backend -n "$NAMESPACE"
    
    kubectl rollout status deployment/frontend -n "$NAMESPACE" --timeout=300s
    kubectl rollout status deployment/backend -n "$NAMESPACE" --timeout=300s
    
    log_success "Rollback completed"
}

# Main deployment function
main() {
    log_info "Starting deployment to ${ENVIRONMENT} environment in ${REGION} region"
    
    # Validate environment
    if [[ ! "$ENVIRONMENT" =~ ^(staging|production)$ ]]; then
        log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'"
        exit 1
    fi
    
    # Set up error handling
    trap 'log_error "Deployment failed! Check the logs above."; exit 1' ERR
    
    # Run deployment steps
    check_prerequisites
    configure_aws
    setup_namespace
    deploy_infrastructure
    
    if [ "${SKIP_BUILD:-false}" != "true" ]; then
        build_and_push_images
    fi
    
    deploy_application
    run_health_checks
    cleanup
    
    log_success "Deployment to ${ENVIRONMENT} completed successfully!"
    
    # Print useful information
    echo ""
    log_info "Deployment Summary:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Region: $REGION"
    echo "  Namespace: $NAMESPACE"
    echo "  Frontend URL: $([ "$ENVIRONMENT" = "production" ] && echo "https://taishanglaojun.ai" || echo "https://staging.taishanglaojun.ai")"
    echo "  Backend URL: $([ "$ENVIRONMENT" = "production" ] && echo "https://api.taishanglaojun.ai" || echo "https://api-staging.taishanglaojun.ai")"
    echo ""
    log_info "To monitor the deployment:"
    echo "  kubectl get pods -n $NAMESPACE"
    echo "  kubectl logs -f deployment/backend -n $NAMESPACE"
    echo "  kubectl logs -f deployment/frontend -n $NAMESPACE"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "rollback")
        rollback
        ;;
    "health")
        run_health_checks
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        echo "Usage: $0 [deploy|rollback|health|cleanup] [environment] [region]"
        echo "  deploy   - Deploy the application (default)"
        echo "  rollback - Rollback to previous version"
        echo "  health   - Run health checks"
        echo "  cleanup  - Clean up old resources"
        echo ""
        echo "Examples:"
        echo "  $0 deploy staging us-east-1"
        echo "  $0 deploy production eu-central-1"
        echo "  $0 rollback staging"
        exit 1
        ;;
esac