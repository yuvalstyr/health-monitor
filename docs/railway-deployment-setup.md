# Railway Deployment Setup Guide

This guide walks you through setting up automated deployment to Railway using GitHub Actions.

## Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **GitHub Repository**: Your code should be in a GitHub repository
3. **Railway Project**: Create a new project in Railway connected to your GitHub repo

## Setup Steps

### 1. Railway Project Configuration

1. **Create Railway Project**:
   - Go to [Railway Dashboard](https://railway.app/dashboard)
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository

2. **Configure Environment Variables**:
   - In your Railway project, click on your service
   - Go to the "Variables" tab
   - Add the following variables:
     ```
     LOG_LEVEL=info
     DB_PATH=/data/health-monitor.db
     ```

3. **Volume Configuration**:
   Railway will automatically handle volume creation based on your `railway.toml` configuration:
   
   ```toml
   [environments.production.volumes]
   data = { mountPath = "/data" }
   ```
   
   This creates a persistent volume mounted at `/data` for your SQLite database. No manual volume setup is required in the Railway UI - it's handled automatically during deployment.

### 2. GitHub Secrets Configuration

1. **Generate Railway Token**:
   - Go to [Railway Account Settings](https://railway.app/account/tokens)
   - Click "Create New Token"
   - Give it a descriptive name (e.g., "GitHub Actions Deploy")
   - Copy the generated token

2. **Add GitHub Secret**:
   
   **Method 1 - Direct search**:
   - Go to your GitHub repository
   - Click on "Settings" tab
   - In the left sidebar, scroll down and look for "Secrets and variables"
   - Click "Secrets and variables" → "Actions"
   
   **Method 2 - If you only see "Code security"**:
   - Go to your repository Settings
   - Scroll down in the left sidebar to find "Actions" section
   - Click "Actions" → "General" 
   - Make sure Actions are enabled for your repository
   - Then look for "Secrets and variables" in the sidebar
   
   **Method 3 - Direct URL**:
   - Navigate to: `https://github.com/[your-username]/[your-repo]/settings/secrets/actions`
   - Replace `[your-username]` and `[your-repo]` with your actual values
   
   Once you find the secrets page:
   - Click "New repository secret" (green button)
   - Name: `RAILWAY_TOKEN`
   - Secret: Paste the Railway token you copied
   - Click "Add secret"

### 3. Railway Configuration File

The `railway.toml` file is already configured in your repository:

```toml
[build]
builder = "nixpacks"
buildCommand = "make railway-build"

[deploy]
startCommand = "make railway-start"
healthcheckPath = "/health"
healthcheckTimeout = 300
restartPolicyType = "on_failure"
restartPolicyMaxRetries = 3

[environments.production]
variables = { LOG_LEVEL = "info", DB_PATH = "/data/health-monitor.db" }

[environments.production.volumes]
data = { mountPath = "/data" }
```

**Key Configuration Details**:
- **Volume**: Automatically creates a persistent volume at `/data` for SQLite database
- **Health Check**: Railway will monitor `/health` endpoint
- **Auto-restart**: Restarts on failure up to 3 times
- **Build Process**: Uses your Makefile targets for consistent builds

## Deployment Process

### Automatic Deployment

Once configured, deployments happen automatically:

1. **Push to main branch** → Triggers deployment
2. **Tests run** → Must pass before deployment
3. **Railway deployment** → Application deployed to Railway
4. **Health checks** → Verifies deployment success
5. **Rollback** → Automatic rollback if deployment fails

### Manual Deployment

You can also trigger deployments manually:

1. Go to your GitHub repository
2. Navigate to "Actions"
3. Select "CI/CD Pipeline"
4. Click "Run workflow"
5. Choose the branch and click "Run workflow"

## Monitoring and Troubleshooting

### Railway Dashboard

Monitor your deployment at:
- **Project Dashboard**: `https://railway.app/project/[your-project-id]`
- **Deployments**: View deployment history and logs
- **Metrics**: Monitor CPU, memory, and request metrics
- **Logs**: Real-time application logs

### Health Check Endpoint

Your application provides a health check at:
```
https://[your-app-domain].railway.app/health
```

Response format:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "database": true,
  "version": "1.0.0"
}
```

### GitHub Actions Logs

View deployment logs in GitHub:
1. Go to "Actions" tab in your repository
2. Click on the latest workflow run
3. Expand job steps to see detailed logs

## Rollback Process

### Automatic Rollback

If deployment fails, the system automatically:
1. Detects the failure
2. Finds the last successful deployment
3. Rolls back to that version
4. Verifies the rollback with health checks

### Manual Rollback

If you need to manually rollback:

1. **Using Railway CLI**:
   ```bash
   railway login
   railway deployments
   railway rollback [deployment-id]
   ```

2. **Using Railway Dashboard**:
   - Go to "Deployments"
   - Find the deployment to rollback to
   - Click "Rollback"

## Cost Management

### Free Tier Limits

Railway Hobby Plan ($5/month):
- **Execution Time**: 500 hours/month
- **Memory**: Up to 8GB
- **Storage**: 100GB
- **Bandwidth**: 100GB

### Usage Monitoring

Monitor usage in Railway dashboard:
- **Metrics** → View resource consumption
- **Usage** → Track against limits
- **Billing** → Monitor costs

## Troubleshooting

### Common Issues

1. **Deployment Timeout**:
   - Check build logs for errors
   - Verify all dependencies are available
   - Ensure database migrations complete successfully

2. **Health Check Failures**:
   - Verify `/health` endpoint is accessible
   - Check database connectivity
   - Review application logs

3. **Database Issues**:
   - Ensure volume is properly mounted at `/data`
   - Check database file permissions
   - Verify migration scripts

4. **Environment Variables**:
   - Confirm all required variables are set
   - Check variable names match exactly
   - Verify sensitive data is in Railway secrets

### Getting Help

- **Railway Documentation**: [docs.railway.app](https://docs.railway.app)
- **Railway Discord**: [discord.gg/railway](https://discord.gg/railway)
- **GitHub Actions Docs**: [docs.github.com/actions](https://docs.github.com/en/actions)

## Security Best Practices

1. **Secrets Management**:
   - Never commit Railway tokens to code
   - Use GitHub secrets for sensitive data
   - Rotate tokens periodically

2. **Access Control**:
   - Limit Railway project access
   - Use principle of least privilege
   - Monitor access logs

3. **Database Security**:
   - Regular backups via Railway volumes
   - Monitor for unusual access patterns
   - Keep application dependencies updated

## Next Steps

After setup is complete:

1. **Test the deployment** by pushing a small change
2. **Monitor the first few deployments** to ensure everything works
3. **Set up monitoring alerts** in Railway dashboard
4. **Document any custom configurations** for your team

Your deployment automation is now ready! 🚀