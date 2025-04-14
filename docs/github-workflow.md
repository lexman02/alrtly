# GitHub Workflow Guidelines

This document outlines our development workflow using GitHub.

## Branch Structure

- `main` - Main development branch
- Feature branches - Individual development work

## Development Workflow

1. **Create Feature Branch**
   ```bash
   git checkout main
   git pull
   git checkout -b feature/my-new-feature
   ```

2. **Develop and Test**
   - Make your changes
   - Commit frequently with meaningful messages
   - Push your branch to GitHub

3. **Open Pull Request**
   - Create a PR from your feature branch to `main`
   - GitHub Actions will build (but not push) a test image
   - Request reviews if working in a team

4. **Merge to Main**
   - After approval, merge your feature into `main`
   - This triggers a build with `dev` and `sha-{hash}` tags
   - The `dev` tag always points to the latest commit on main

5. **Create Release**
   - When ready for a production release, create a GitHub Release with semantic versioning
   - This builds and tags images with:
     - The version number (e.g., `v1.2.3`)
     - `latest` tag for the most recent release

## Manual Deployments

If needed, use the "workflow_dispatch" trigger via GitHub UI to manually build and push a custom-tagged image from any branch or commit.