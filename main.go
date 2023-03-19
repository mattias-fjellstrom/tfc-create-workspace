package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
)

const (
	ENV_TERRAFORM_CLOUD_TOKEN        = "TERRAFORM_CLOUD_TOKEN"
	ENV_TERRAFORM_CLOUD_ORGANIZATION = "TERRAFORM_CLOUD_ORGANIZATION"
	ENV_TERRAFORM_CLOUD_PROJECT      = "TERRAFORM_CLOUD_PROJECT"
	ENV_TERRAFORM_CLOUD_WORKSPACE    = "TERRAFORM_CLOUD_WORKSPACE"
	ENV_REPOSITORY_NAME              = "GITHUB_REPOSITORY"
)

var organizationName string
var projectName string
var workspaceName string
var repositoryName string
var workingDirectory string
var branchName string

func init() {
	flag.StringVar(&organizationName, "organization", "", "Terraform Cloud organization name")
	flag.StringVar(&projectName, "project", "", "Terraform Cloud project name")
	flag.StringVar(&workspaceName, "workspace", "", "Desired Terraform Cloud workspace name")
	flag.StringVar(&repositoryName, "repository", "", "Git repository to connect the Terraform Cloud workspace to")
	flag.StringVar(&workingDirectory, "working_directory", "", "Directory in repository containing the Terraform configuration")
	flag.StringVar(&branchName, "branch", "", "Branch name Terraform runs should trigger on")
}

func main() {
	flag.Parse()

	if organizationName == "" {
		log.Println("No organization name provided as input argument, will fall back to environment variable")
		_, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_ORGANIZATION)
		if !ok {
			log.Fatalf("The organization name must be provided either as an input parameter or in the %s environment variable", ENV_TERRAFORM_CLOUD_ORGANIZATION)
		}
		organizationName = os.Getenv(ENV_TERRAFORM_CLOUD_ORGANIZATION)
		log.Println("Organization name read from environment variable")
	}

	if workspaceName == "" {
		log.Println("No workspace name provided as input argument, will fall back to environment variable")
		_, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_WORKSPACE)
		if !ok {
			log.Fatalf("A workspace name must be provided either as an input parameter or in the %s environment variable", ENV_TERRAFORM_CLOUD_WORKSPACE)
		}
		workspaceName = os.Getenv(ENV_TERRAFORM_CLOUD_WORKSPACE)
		log.Println("Workspace name read from environment variable")
	}

	if repositoryName == "" {
		log.Println("No repository name provided as input argument, will fall back to environment variable")
		_, ok := os.LookupEnv(ENV_REPOSITORY_NAME)
		if !ok {
			log.Fatalf("The repository name could not be read from the %s environment variable and no value was provided as an input parameter", ENV_REPOSITORY_NAME)
		}
		repositoryName = os.Getenv(ENV_REPOSITORY_NAME)
		log.Println("Current repository read from environment variable")
	}

	if projectName == "" {
		log.Println("No project name provided as input argument, will fall back to environment variable or use default project")
		_, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_PROJECT)
		if ok {
			projectName = os.Getenv(ENV_TERRAFORM_CLOUD_PROJECT)
			log.Println("Project name read from environment variable")
		}
	}

	token, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_TOKEN)
	if !ok || token == "" {
		log.Fatalf("%s environment variable must be set with a valid token", ENV_TERRAFORM_CLOUD_TOKEN)
	}

	config := &tfe.Config{
		Token:             token,
		RetryServerErrors: true,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// get github app installation id
	// search for one named as the github organization
	githubOrganization := strings.Split(repositoryName, "/")[0]
	log.Printf("GitHub organization is set to %s", githubOrganization)
	gitHubApplications, err := client.GHAInstallations.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Discovered %d GitHub Application Installations", len(gitHubApplications.Items))

	var gitHubApplication *tfe.GHAInstallation
	for _, gitHubAppItem := range gitHubApplications.Items {
		log.Printf("Found application with name %s", *gitHubAppItem.Name)
		if *gitHubAppItem.Name == githubOrganization {
			gitHubApplication = gitHubAppItem
			log.Printf("Set active github app to ID %s", *gitHubApplication.ID)
		}
	}

	// find the project ID
	projects, err := client.Projects.List(ctx, organizationName, &tfe.ProjectListOptions{
		ListOptions: tfe.ListOptions{
			PageNumber: 1,
			PageSize:   100,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Discovered %d projects", len(projects.Items))
	var project *tfe.Project = nil
	for _, projectItem := range projects.Items {
		if projectItem.Name == projectName {
			project = projectItem
		}
	}

	_, err = client.Workspaces.Create(ctx, organizationName, tfe.WorkspaceCreateOptions{
		Type:             "workspaces",
		Name:             tfe.String(workspaceName),
		AutoApply:        tfe.Bool(true),
		WorkingDirectory: tfe.String(workingDirectory),
		VCSRepo: &tfe.VCSRepoOptions{
			Branch:            tfe.String(branchName),
			Identifier:        tfe.String(repositoryName),
			GHAInstallationID: gitHubApplication.ID,
		},
		Project: project,
	})
	if err != nil {
		log.Fatal(err)
	}
}
