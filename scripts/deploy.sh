#!/bin/bash

# Gruvit Music Application Deployment Script
# This script handles both Docker Compose and Kubernetes deployments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="development"
NAMESPACE="gruvit"
REGISTRY=""

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if [ "$ENVIRONMENT" = "production" ]; then
        if ! command_exists kubectl; then
            print_error "kubectl is required for production deployment"
            exit 1
        fi
        
        if ! command_exists docker; then
            print_error "Docker is required for building images"
            exit 1
        fi
    else
        if ! command_exists docker; then
            print_error "Docker is required"
            exit 1
        fi
        
        if ! command_exists docker-compose; then
            print_error "docker-compose is required"
            exit 1
        fi
    fi
    
    print_status "Prerequisites check passed"
}

# Function to build Docker images
build_images() {
    print_status "Building Docker images..."
    
    # Build frontend
    print_status "Building frontend image..."
    docker build -t gruvit/frontend:latest ./client
    
    # Build Java service
    print_status "Building Java auth service image..."
    docker build -t gruvit/java-auth-service:latest ./server/java
    
    # Build Go service
    print_status "Building Go music service image..."
    docker build -t gruvit/go-music-service:latest ./server/go
    
    print_status "All images built successfully"
}

# Function to push images to registry
push_images() {
    if [ -n "$REGISTRY" ]; then
        print_status "Pushing images to registry: $REGISTRY"
        
        docker tag gruvit/frontend:latest $REGISTRY/gruvit/frontend:latest
        docker tag gruvit/java-auth-service:latest $REGISTRY/gruvit/java-auth-service:latest
        docker tag gruvit/go-music-service:latest $REGISTRY/gruvit/go-music-service:latest
        
        docker push $REGISTRY/gruvit/frontend:latest
        docker push $REGISTRY/gruvit/java-auth-service:latest
        docker push $REGISTRY/gruvit/go-music-service:latest
        
        print_status "Images pushed successfully"
    fi
}

# Function to deploy with Docker Compose
deploy_docker_compose() {
    print_status "Deploying with Docker Compose..."
    
    # Check if .env file exists
    if [ ! -f .env ]; then
        print_warning ".env file not found, creating from .env.example"
        cp .env.example .env
        print_warning "Please update .env file with your actual values"
    fi
    
    # Start services
    docker-compose up -d
    
    print_status "Services started successfully"
    print_status "Frontend: http://localhost:3000"
    print_status "API Gateway: http://localhost:80"
    print_status "Java Service: http://localhost:8080"
    print_status "Go Service: http://localhost:8081"
}

# Function to deploy with Kubernetes
deploy_kubernetes() {
    print_status "Deploying with Kubernetes..."
    
    # Create namespace
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply configurations
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/secrets.yaml
    
    # Deploy services
    kubectl apply -f k8s/mongodb-deployment.yaml
    kubectl apply -f k8s/redis-deployment.yaml
    kubectl apply -f k8s/java-service-deployment.yaml
    kubectl apply -f k8s/go-service-deployment.yaml
    kubectl apply -f k8s/frontend-deployment.yaml
    kubectl apply -f k8s/nginx-deployment.yaml
    
    print_status "Kubernetes deployment completed"
    print_status "Waiting for services to be ready..."
    
    # Wait for deployments
    kubectl wait --for=condition=available --timeout=300s deployment/java-auth-service -n $NAMESPACE
    kubectl wait --for=condition=available --timeout=300s deployment/go-music-service -n $NAMESPACE
    kubectl wait --for=condition=available --timeout=300s deployment/frontend -n $NAMESPACE
    kubectl wait --for=condition=available --timeout=300s deployment/nginx-gateway -n $NAMESPACE
    
    print_status "All services are ready!"
}

# Function to show status
show_status() {
    if [ "$ENVIRONMENT" = "production" ]; then
        print_status "Kubernetes deployment status:"
        kubectl get pods -n $NAMESPACE
        kubectl get services -n $NAMESPACE
    else
        print_status "Docker Compose status:"
        docker-compose ps
    fi
}

# Function to show logs
show_logs() {
    if [ "$ENVIRONMENT" = "production" ]; then
        print_status "Showing logs for all services..."
        kubectl logs -l app=nginx-gateway -n $NAMESPACE --tail=50
        kubectl logs -l app=frontend -n $NAMESPACE --tail=50
        kubectl logs -l app=java-auth-service -n $NAMESPACE --tail=50
        kubectl logs -l app=go-music-service -n $NAMESPACE --tail=50
    else
        print_status "Showing Docker Compose logs..."
        docker-compose logs --tail=50
    fi
}

# Function to cleanup
cleanup() {
    if [ "$ENVIRONMENT" = "production" ]; then
        print_status "Cleaning up Kubernetes resources..."
        kubectl delete namespace $NAMESPACE
    else
        print_status "Cleaning up Docker Compose resources..."
        docker-compose down -v
    fi
}

# Main function
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            --build-only)
                check_prerequisites
                build_images
                push_images
                exit 0
                ;;
            --status)
                show_status
                exit 0
                ;;
            --logs)
                show_logs
                exit 0
                ;;
            --cleanup)
                cleanup
                exit 0
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  -e, --environment    Environment (development|production) [default: development]"
                echo "  -n, --namespace      Kubernetes namespace [default: gruvit]"
                echo "  -r, --registry       Docker registry URL"
                echo "  --build-only         Only build and push images"
                echo "  --status             Show deployment status"
                echo "  --logs               Show service logs"
                echo "  --cleanup            Clean up resources"
                echo "  -h, --help           Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Main deployment flow
    check_prerequisites
    
    if [ "$ENVIRONMENT" = "production" ]; then
        build_images
        push_images
        deploy_kubernetes
    else
        deploy_docker_compose
    fi
    
    show_status
    print_status "Deployment completed successfully!"
}

# Run main function
main "$@"
