package featuredetection

import (
	"net/http"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghinstance"
)

type Detector interface {
	IssueFeatures() (IssueFeatures, error)
	PullRequestFeatures() (PullRequestFeatures, error)
	RepositoryFeatures() (RepositoryFeatures, error)
}

type IssueFeatures struct {
	StateReason bool
}

var allIssueFeatures = IssueFeatures{
	StateReason: true,
}

type PullRequestFeatures struct {
	MergeQueue bool
}

var allPullRequestFeatures = PullRequestFeatures{
	MergeQueue: true,
}

type RepositoryFeatures struct {
	PullRequestTemplateQuery bool
	VisibilityField          bool
	AutoMerge                bool
}

var allRepositoryFeatures = RepositoryFeatures{
	PullRequestTemplateQuery: true,
	VisibilityField:          true,
	AutoMerge:                true,
}

type detector struct {
	host       string
	httpClient *http.Client
}

func NewDetector(httpClient *http.Client, host string) Detector {
	return &detector{
		httpClient: httpClient,
		host:       host,
	}
}

func (d *detector) IssueFeatures() (IssueFeatures, error) {
	if !ghinstance.IsEnterprise(d.host) {
		return allIssueFeatures, nil
	}

	features := IssueFeatures{
		StateReason: false,
	}

	var featureDetection struct {
		Issue struct {
			Fields []struct {
				Name string
			} `graphql:"fields(includeDeprecated: true)"`
		} `graphql:"Issue: __type(name: \"Issue\")"`
	}

	gql := api.NewClientFromHTTP(d.httpClient)
	err := gql.Query(d.host, "Issue_fields", &featureDetection, nil)
	if err != nil {
		return features, err
	}

	for _, field := range featureDetection.Issue.Fields {
		if field.Name == "stateReason" {
			features.StateReason = true
		}
	}

	return features, nil
}

func (d *detector) PullRequestFeatures() (PullRequestFeatures, error) {
	// TODO: reinstate the short-circuit once the APIs are fully available on github.com
	// https://github.com/cli/cli/issues/5778
	//
	// if !ghinstance.IsEnterprise(d.host) {
	// 	return allPullRequestFeatures, nil
	// }

	features := PullRequestFeatures{}

	var featureDetection struct {
		PullRequest struct {
			Fields []struct {
				Name string
			} `graphql:"fields(includeDeprecated: true)"`
		} `graphql:"PullRequest: __type(name: \"PullRequest\")"`
	}

	gql := api.NewClientFromHTTP(d.httpClient)
	err := gql.Query(d.host, "PullRequest_fields", &featureDetection, nil)
	if err != nil {
		return features, err
	}

	for _, field := range featureDetection.PullRequest.Fields {
		if field.Name == "isInMergeQueue" {
			features.MergeQueue = true
		}
	}

	return features, nil
}

func (d *detector) RepositoryFeatures() (RepositoryFeatures, error) {
	if !ghinstance.IsEnterprise(d.host) {
		return allRepositoryFeatures, nil
	}

	features := RepositoryFeatures{}

	var featureDetection struct {
		Repository struct {
			Fields []struct {
				Name string
			} `graphql:"fields(includeDeprecated: true)"`
		} `graphql:"Repository: __type(name: \"Repository\")"`
	}

	gql := api.NewClientFromHTTP(d.httpClient)

	err := gql.Query(d.host, "Repository_fields", &featureDetection, nil)
	if err != nil {
		return features, err
	}

	for _, field := range featureDetection.Repository.Fields {
		if field.Name == "pullRequestTemplates" {
			features.PullRequestTemplateQuery = true
		}
		if field.Name == "visibility" {
			features.VisibilityField = true
		}
		if field.Name == "autoMergeAllowed" {
			features.AutoMerge = true
		}
	}

	return features, nil
}
