# CI/CD Integration

This guide shows you how to integrate Frank CLI into your CI/CD pipelines for automated deployments.

## GitHub Actions

### Basic Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy with Frank

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Build Frank
        run: |
          go build -o frank .
          chmod +x frank

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'latest'

      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig

      - name: Deploy to development
        if: github.ref == 'refs/heads/develop'
        run: |
          ./frank apply dev --yes

      - name: Deploy to production
        if: github.ref == 'refs/heads/main'
        run: |
          ./frank apply prod --yes
```

### Multi-Environment Workflow

```yaml
# .github/workflows/deploy-multi-env.yml
name: Multi-Environment Deploy

on:
  push:
    branches: [ main, develop, staging ]

jobs:
  deploy-dev:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    environment: development
    steps:
      - uses: actions/checkout@v4
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      - name: Deploy to dev
        run: |
          ./frank apply dev --yes

  deploy-staging:
    if: github.ref == 'refs/heads/staging'
    runs-on: ubuntu-latest
    environment: staging
    needs: deploy-dev
    steps:
      - uses: actions/checkout@v4
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      - name: Deploy to staging
        run: |
          ./frank apply staging --yes

  deploy-prod:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    needs: deploy-staging
    steps:
      - uses: actions/checkout@v4
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      - name: Deploy to production
        run: |
          ./frank apply prod --yes
```

### Pull Request Preview

```yaml
# .github/workflows/pr-preview.yml
name: PR Preview

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      
      - name: Deploy preview
        run: |
          # Create preview namespace
          kubectl create namespace pr-${{ github.event.number }} || true
          
          # Deploy with preview namespace
          NAMESPACE=pr-${{ github.event.number }} ./frank apply dev --yes
      
      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🚀 Preview deployed to namespace: pr-${{ github.event.number }}'
            })
```

## GitLab CI

### Basic Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

variables:
  FRANK_VERSION: "latest"

build:
  stage: build
  image: golang:1.25-alpine
  script:
    - go build -o frank .
    - chmod +x frank
  artifacts:
    paths:
      - frank
    expire_in: 1 hour

deploy-dev:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - ./frank apply dev --yes
  only:
    - develop

deploy-staging:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - ./frank apply staging --yes
  only:
    - staging

deploy-prod:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - ./frank apply prod --yes
  only:
    - main
  when: manual
```

## Jenkins

### Jenkinsfile

```groovy
pipeline {
    agent any
    
    stages {
        stage('Build') {
            steps {
                sh 'go build -o frank .'
                sh 'chmod +x frank'
            }
        }
        
        stage('Deploy Dev') {
            when {
                branch 'develop'
            }
            steps {
                sh './frank apply dev --yes'
            }
        }
        
        stage('Deploy Staging') {
            when {
                branch 'staging'
            }
            steps {
                sh './frank apply staging --yes'
            }
        }
        
        stage('Deploy Prod') {
            when {
                branch 'main'
            }
            steps {
                input message: 'Deploy to production?', ok: 'Deploy'
                sh './frank apply prod --yes'
            }
        }
    }
}
```

## Azure DevOps

### azure-pipelines.yml

```yaml
trigger:
- main
- develop

pool:
  vmImage: 'ubuntu-latest'

stages:
- stage: Build
  jobs:
  - job: BuildFrank
    steps:
    - task: Go@0
      inputs:
        command: 'build'
        arguments: '-o frank .'
    - task: PublishBuildArtifacts@1
      inputs:
        pathToPublish: 'frank'
        artifactName: 'frank-binary'

- stage: DeployDev
  condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/develop'))
  dependsOn: Build
  jobs:
  - deployment: DeployToDev
    environment: 'development'
    strategy:
      runOnce:
        deploy:
          steps:
          - download: current
          - task: Kubectl@1
            inputs:
              command: 'apply'
              arguments: '-f kubeconfig'
          - script: './frank apply dev --yes'

- stage: DeployProd
  condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/main'))
  dependsOn: Build
  jobs:
  - deployment: DeployToProd
    environment: 'production'
    strategy:
      runOnce:
        deploy:
          steps:
          - download: current
          - task: Kubectl@1
            inputs:
              command: 'apply'
              arguments: '-f kubeconfig'
          - script: './frank apply prod --yes'
```

## Environment Variables

### Required Environment Variables

```bash
# Kubernetes configuration
KUBECONFIG=/path/to/kubeconfig

# Frank configuration
FRANK_LOG_LEVEL=info

# Application secrets (don't put in config files)
DOCKER_REGISTRY_PASSWORD=secret123
API_KEY=your-api-key
```

### GitHub Secrets

Set these in your GitHub repository settings:

- `KUBE_CONFIG` - Base64 encoded kubeconfig file
- `DOCKER_REGISTRY_PASSWORD` - Docker registry password
- `API_KEY` - Application API key

### GitLab Variables

Set these in your GitLab project settings:

- `KUBECONFIG` - Kubernetes configuration
- `DOCKER_REGISTRY_PASSWORD` - Docker registry password
- `API_KEY` - Application API key

## Security Best Practices

### 1. Use Secrets Management

```yaml
# Don't put secrets in config files
# config/prod/config.yaml
context: prod-cluster
namespace: myapp-prod
# NO SECRETS HERE!

# Use environment variables instead
# In your CI/CD pipeline
env:
  - name: DOCKER_REGISTRY_PASSWORD
    valueFrom:
      secretKeyRef:
        name: docker-secret
        key: password
```

### 2. Use Service Accounts

```yaml
# Create service account for CI/CD
apiVersion: v1
kind: ServiceAccount
metadata:
  name: frank-ci
  namespace: myapp-prod
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: frank-ci
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: frank-ci
  namespace: myapp-prod
```

### 3. Use Namespace Isolation

```yaml
# Deploy to specific namespaces
# config/dev/config.yaml
namespace: myapp-dev

# config/prod/config.yaml
namespace: myapp-prod
```

## Monitoring and Alerting

### Deployment Status

```yaml
# .github/workflows/deploy-with-status.yml
name: Deploy with Status

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      
      - name: Deploy
        run: |
          ./frank apply prod --yes
      
      - name: Check deployment status
        run: |
          kubectl get deployments -l app.kubernetes.io/managed-by=frank
          kubectl get pods -l app.kubernetes.io/managed-by=frank
      
      - name: Notify on success
        if: success()
        run: |
          echo "✅ Deployment successful"
      
      - name: Notify on failure
        if: failure()
        run: |
          echo "❌ Deployment failed"
```

### Slack Notifications

```yaml
# .github/workflows/deploy-with-slack.yml
name: Deploy with Slack

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Frank
        run: |
          go build -o frank .
          chmod +x frank
      
      - name: Deploy
        run: |
          ./frank apply prod --yes
      
      - name: Notify Slack
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          channel: '#deployments'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()
```

## Troubleshooting

### Common CI/CD Issues

**"kubectl: command not found"**
- Install kubectl in your CI environment
- Use a Docker image with kubectl pre-installed

**"context not found"**
- Ensure kubeconfig is properly configured
- Check context names in your configuration

**"permission denied"**
- Verify service account permissions
- Check RBAC settings

**"timeout waiting for resource"**
- Increase timeout values in configuration
- Check cluster resources

### Debug Commands

```bash
# Enable debug logging in CI
FRANK_LOG_LEVEL=debug ./frank apply dev --yes

# Check kubectl configuration
kubectl config get-contexts
kubectl get nodes

# Check deployment status
kubectl get deployments -l app.kubernetes.io/managed-by=frank
kubectl describe deployment myapp
```

## Next Steps

- [Multi-Environment Setup](multi-environment.md) - Configure multiple environments
- [Debugging](debugging.md) - Troubleshoot deployment issues
- [Best Practices](best-practices.md) - Learn advanced patterns
