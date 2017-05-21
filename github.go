package main

import (
	"context"
	"log"
	"strings"
	"time"

	"os"

	"github.com/google/go-github/github"
	"github.com/mmcdole/gofeed"
	"golang.org/x/oauth2"
)

func newGithubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_KEY")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}

func createGitTag(item *gofeed.Item, version string) error {
	log.Println("New TAG", version)
	client := newGithubClient()
	ctx := context.Background()
	tagRef := "tags/" + version

	ref, _, err := client.Git.GetRef(ctx, githubRepoOwner, githubRepoName, tagRef)
	if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
		return err
	}
	if ref != nil {
		log.Println("SKIP Tag exists")
		return nil
	}

	ref, _, err = client.Git.GetRef(ctx, githubRepoOwner, githubRepoName, "heads/master")
	if err != nil {
		return err
	}

	message := item.Title + "\n" + item.GUID
	taggerTime := time.Now().Truncate(time.Second)
	taggerName := "Florian Kinder"
	taggerEmail := "florian.kinder@fankserver.com"

	tag, _, err := client.Git.CreateTag(ctx, githubRepoOwner, githubRepoName, &github.Tag{
		Message: &message,
		Object:  ref.Object,
		Tag:     &version,
		Tagger: &github.CommitAuthor{
			Date:  &taggerTime,
			Name:  &taggerName,
			Email: &taggerEmail,
		},
	})
	if err != nil {
		return err
	}

	ref, _, err = client.Git.CreateRef(ctx, githubRepoOwner, githubRepoName, &github.Reference{
		Object: tag.Object,
		Ref:    &tagRef,
	})
	if err != nil {
		return err
	}

	log.Println("Tag created")

	return nil
}
