# Jenkins CI/CD

This document explains the Jenkins pipeline used to automate Docker image builds and publishing to GitHub Container Registry (GHCR).

## Overview

The project uses **Jenkins** for CI/CD automation. The pipeline is defined in the `Jenkinsfile` at the project root and is typically run on a self-hosted Jenkins server (often via Docker Compose).

## Pipeline Workflow

### 1. Build & Publish Docker Image

**Purpose**: Build the project Docker image and publish it to GHCR for deployment or use elsewhere.

**File**: [`Jenkinsfile`](../Jenkinsfile)

**Trigger Events**:

- Manual build via Jenkins UI
- SCM polling: **Every Tuesday** (`H H * * 2`)

**Scope**: Builds the Docker image from the repository and pushes both a versioned and `latest` tag to GHCR.

**Key Steps**:

- Clean the workspace before each build to prevent leftover or stale files from interfering with the pipeline
  - Reason: Jenkins can sometimes reuse old workspace files, causing unexpected conflicts or build errors. A cleanup step ensures every run starts from a known good state.
- Run a simple sanity check log message
- Build Docker image and tag with Jenkins build number
- Authenticate to GHCR using a GitHub PAT
- Push both versioned and `latest` tags to GHCR

**Key Benefit**: Ensures a reproducible, versioned Docker image is always available in the registry for deployment or local use

---

## Prerequisites

- Jenkins server (can be run via Docker Compose)
- Docker installed on the Jenkins host
- Jenkins container must have access to the Docker daemon and CLI (see below)
- GitHub PAT with `write:packages` scope stored in Jenkins credentials as `GHCR_PAT`

**Example `docker-compose.yaml`:**

```yaml
services:
  image: jenkins/jenkins:lts
  container_name: jenkins_server
  user: root
  ports:
    - "8080:8080"
    - "50000:50000"
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock
    - jenkins_data:/var/jenkins_home
    - /usr/bin/docker:/usr/bin/docker
  env_file:
    - .env

volumes:
  jenkins_data:
```

For the complete Docker Compose setup, refer to the [`docker-compose.yml`](https://github.com/victoriacheng15/home-server/blob/main/docker-compose.yml) file in the `home-server` repository.

---

## Monitoring & Debugging

### Viewing Pipeline Results

1. Open Jenkins in your browser
2. Select the relevant job
3. Click a build number to view logs and results

### Troubleshooting

- **Workspace contains old or unexpected files**:
  - Use the bash commands:

    ```bash
    docker exec -it jenkins_server bash
    rm -rf /var/jenkins_home/workspace/<job_name>
    ```

  - Or rely on the automated cleanup step included in the pipeline
- **`docker: not found`**: Ensure `/usr/bin/docker` is mounted and Docker is installed on the host
- **Permission errors**: Ensure Jenkins runs as `root` and has access to the Docker socket
- **Credential errors**: Ensure the GitHub PAT is stored as `GHCR_PAT` in Jenkins credentials
