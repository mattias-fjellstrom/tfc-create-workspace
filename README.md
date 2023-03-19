# GitHub Actions for creating workspaces in Terraform Cloud

With this action you can create workspaces in Terraform Cloud as part of your GitHub Actions workflows.

- This action should be preceded by the [mattias-fjellstrom/tfc-setup](https://github.com/mattias-fjellstrom/tfc-setup/blob/main/action.yaml) action to configure required environment variables. See the sample below.
- The action only supports authenticating to Terraform Cloud using an API token, read the [Terraform Cloud documentation](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/api-tokens) about tokens.
- Currently an GitHub App installation in your repository is required for this action, read the [Terraform Cloud documentation](https://developer.hashicorp.com/terraform/cloud-docs/vcs/github-app) on how to create one. The GitHub App must be installed in the same GitHub organization where you use this action.
- You can specify which repository directory your workspace is connected to in the `directory` property. The default value is the root of the repository.
- You can specify which branch the workspace should create infrastructure from using the `branch` property. The default is the repository default branch (e.g. `main`).

## Sample workflow

Below is a full sample workflow that sets up a new workspace in Terraform Cloud when a pull-request is opened, and deletes the workspace once the pull-request is closed. If the pull request is opened for a branch named `feature-1` the resulting workspace will be named `my-workspace-feature-1` in this sample.

```yaml
name: Terraform Cloud workspaces for pull-requests

on:
  pull_request:
    types:
      - opened
      - closed

env:
  ORGANIZATION: my-terraform-cloud-organization
  PROJECT: my-terraform-cloud-project
  WORKSPACE: my-workspace-${{ github.head_ref }}

jobs:
  create-workspace:
    if: ${{ github.event.action == 'opened' }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: mattias-fjellstrom/tfc-setup@v1
        with:
          token: ${{ secrets.TERRAFORM_CLOUD_TOKEN }}
          organization: ${{ env.ORGANIZATION }}
          project: ${{ env.PROJECT }}
          workspace: ${{ env.WORKSPACE }}
      - uses: mattias-fjellstrom/tfc-create-workspace@v1
        with:
          directory: infrastructure
          branch: ${{ github.head_ref }}
  delete-workspace:
    if: ${{ github.event.action == 'closed' }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: mattias-fjellstrom/tfc-setup@v1
        with:
          token: ${{ secrets.TERRAFORM_CLOUD_TOKEN }}
          organization: ${{ env.ORGANIZATION }}
          project: ${{ env.PROJECT }}
          workspace: ${{ env.WORKSPACE }}
      - uses: mattias-fjellstrom/tfc-delete-workspace@v1
```