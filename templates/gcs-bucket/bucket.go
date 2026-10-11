package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Private helpers ------------------------------------------------------------

// defineBucket uses the published GCP provider directly. Premise does not
// generate or manage a nested component library for this project.
func defineBucket(ctx *pulumi.Context, project, location string, artifacts map[string]string) error {
	files, err := collectArtifacts(artifacts)
	if err != nil {
		return err
	}
	bucket, err := storage.NewBucket(ctx, "artifacts", &storage.BucketArgs{
		Project: pulumi.String(project), Location: pulumi.String(location),
		UniformBucketLevelAccess: pulumi.Bool(true), PublicAccessPrevention: pulumi.String("enforced"), ForceDestroy: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	for _, file := range files {
		id := sha256.Sum256([]byte(file.object))
		args := &storage.BucketObjectArgs{
			Bucket: bucket.Name, Name: pulumi.String(file.object),
			Source: pulumi.NewFileAsset(file.path), SourceMd5hash: pulumi.String(file.checksum),
		}
		if contentType := mime.TypeByExtension(filepath.Ext(file.path)); contentType != "" {
			args.ContentType = pulumi.String(contentType)
		}
		if _, err := storage.NewBucketObject(ctx, "artifact-"+hex.EncodeToString(id[:]), args); err != nil {
			return err
		}
	}
	ctx.Export("bucketName", bucket.Name)
	ctx.Export("environment", pulumi.String(ctx.Stack()))
	return nil
}

func collectArtifacts(artifacts map[string]string) ([]artifactFile, error) {
	// Validate every bound build before registering any cloud resource.
	var files []artifactFile
	apps := make([]string, 0, len(artifacts))
	for app := range artifacts {
		apps = append(apps, app)
	}
	sort.Strings(apps)
	for _, app := range apps {
		if app == "" || strings.ContainsAny(app, "/\\") || app == "." || app == ".." {
			return nil, fmt.Errorf("invalid artifact app name %q", app)
		}
		directory := artifacts[app]
		info, err := os.Lstat(directory)
		if err != nil {
			return nil, fmt.Errorf("build artifact for %s: %w", app, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("artifact for %s must be a directory", app)
		}
		start := len(files)
		err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("artifact %s must be a regular file", path)
			}
			relative, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			digest := md5.Sum(data)
			files = append(files, artifactFile{path: path, object: app + "/" + filepath.ToSlash(relative), checksum: base64.StdEncoding.EncodeToString(digest[:])})
			return nil
		})
		if err != nil {
			return nil, err
		}
		if len(files) == start {
			return nil, fmt.Errorf("artifact directory for %s is empty", app)
		}
	}
	return files, nil
}

type artifactFile struct{ path, object, checksum string }
