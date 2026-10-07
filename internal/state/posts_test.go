package state

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func postTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(context.Background(), ":memory:", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func samplePost() PostInit {
	published := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return PostInit{
		PostMetadataInit: PostMetadataInit{Slug: "sample", Title: "Sample", Description: "Original description", PublishedAt: &published, Tags: []string{"go", "blog"}},
		Body:             "original body", BodyType: PostBodyHtmlTemplate,
		Assets: []PostAssetInit{{Path: "assets/old.svg", ContentType: "image/svg+xml", Data: []byte("old image")}},
	}
}

func revisionCount(t *testing.T, s *Store) int {
	t.Helper()
	var count int
	if err := s.db.QueryRow("SELECT count(*) FROM post_revisions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestUpsertPostReplacesAndSnapshots(t *testing.T) {
	s := postTestStore(t)
	ctx := context.Background()
	input := samplePost()
	original, err := s.UpsertPost(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if revisionCount(t, s) != 0 {
		t.Fatal("creation should not snapshot a nonexistent version")
	}
	input.Body = "edited body"
	input.Description = "Edited description"
	input.PublishedAt = nil
	input.Tags = []string{"new"}
	input.Assets = []PostAssetInit{{Path: "assets/new.svg", ContentType: "image/svg+xml", Data: []byte("new image")}}
	before := time.Now().UTC()
	edited, err := s.UpsertPost(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if edited.ID != original.ID || !edited.CreatedAt.Equal(original.CreatedAt) || !edited.PublishedAt.Equal(original.PublishedAt) {
		t.Fatalf("identity or original dates changed: %+v", edited.PostMetadata)
	}
	if edited.Body != input.Body || edited.Description != input.Description || edited.UpdatedAt.Before(before) {
		t.Fatalf("edit was not applied: %+v", edited)
	}
	if len(edited.Tags) != 1 || len(edited.Assets) != 1 || edited.Assets[0].Path != "assets/new.svg" {
		t.Fatalf("associations not replaced: %+v", edited)
	}
	if _, err := s.GetPostAsset(ctx, input.Slug, "assets/old.svg"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("removed asset still associated: %v", err)
	}
	var snapshot Post
	metadata, err := scanPostMetadata(s.db.QueryRow(`SELECT post_id, slug, title, description, body_type,
		created_at, published_at, updated_at, body FROM post_revisions WHERE post_id = ?`, original.ID), &snapshot.Body)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.PostMetadata = metadata
	if snapshot.PostMetadata != original.PostMetadata || snapshot.Body != original.Body {
		t.Fatalf("snapshot metadata/body differs: %+v", snapshot)
	}
	var count int
	if err := s.db.QueryRow(`SELECT count(*) FROM post_revision_tags rt JOIN post_revisions r ON r.id = rt.revision_id WHERE r.post_id = ?`, original.ID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("snapshot tags = %d, err = %v", count, err)
	}
	var image []byte
	var localPath string
	if err := s.db.QueryRow(`SELECT a.data, ra.local_path FROM post_revision_assets ra JOIN assets a ON a.id = ra.asset_id`).Scan(&image, &localPath); err != nil {
		t.Fatal(err)
	}
	if string(image) != "old image" || localPath != "assets/old.svg" {
		t.Fatal("previous asset not retained")
	}
	if revisionCount(t, s) != 1 {
		t.Fatal("expected one revision")
	}
}

func TestUpsertPostTitleChange(t *testing.T) {
	s := postTestStore(t)
	ctx := context.Background()
	input := samplePost()
	original, err := s.UpsertPost(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Title = "Renamed"
	input.Body = "replacement"
	renamed, err := s.UpsertPost(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if renamed.ID != original.ID || renamed.Title != "Renamed" || revisionCount(t, s) != 1 {
		t.Fatal("title change was not saved and snapshotted")
	}
}
