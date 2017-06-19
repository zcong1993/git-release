package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"io"
	"time"
)

// RLS is git-release object
type RLS struct {
	GitHub    GitHub
	outStream io.Writer
}

// CreateRelease is handler of create github release
func (r *RLS) CreateRelease(ctx context.Context, req *github.RepositoryRelease, recreate bool) (*github.RepositoryRelease, error) {

	// When draft release creation is requested,
	// create it without any check (it can).
	if *req.Draft {
		fmt.Fprintln(r.outStream, "==> Create a draft release")
		return r.GitHub.CreateRelease(ctx, req)
	}

	// Check release exists.
	// If release is not found, then create a new release.
	release, err := r.GitHub.GetRelease(ctx, *req.TagName)
	if err != nil {
		if err != fmt.Errorf("release is not found") {
			return nil, errors.Wrap(err, "failed to get release")
		}
		if recreate {
			fmt.Fprintf(r.outStream,
				"WARNING: '-recreate' is specified but release (%s) not found",
				*req.TagName)
		}

		fmt.Fprintln(r.outStream, "==> Create a new release")
		return r.GitHub.CreateRelease(ctx, req)
	}

	// recreate is not true. Then use that existing release.
	if !recreate {
		fmt.Fprintf(r.outStream, "WARNING: found release (%s). Use existing one.\n",
			*req.TagName)
		return release, nil
	}

	// When recreate is requested, delete existing release and create a
	// new release.
	fmt.Fprintln(r.outStream, "==> Recreate a release")
	if err := r.DeleteRelease(ctx, *release.ID, *req.TagName); err != nil {
		return nil, err
	}

	return r.GitHub.CreateRelease(ctx, req)
}

// DeleteRelease is handler of delete github release
func (r *RLS) DeleteRelease(ctx context.Context, ID int, tag string) error {

	err := r.GitHub.DeleteRelease(ctx, ID)
	if err != nil {
		return err
	}

	err = r.GitHub.DeleteTag(ctx, tag)
	if err != nil {
		return err
	}

	// This is because sometimes the process of creating a release on GitHub
	// is faster than deleting a tag.
	time.Sleep(5 * time.Second)

	return nil
}

// GetCommits pass the same function from github.go
func (r *RLS) GetCommits(ctx context.Context, opts *github.CommitsListOptions) ([]*github.RepositoryCommit, error) {
	return r.GitHub.GetCommits(ctx, opts)
}
