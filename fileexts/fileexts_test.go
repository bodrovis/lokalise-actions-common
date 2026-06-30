package fileexts

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveFromEnv_UsesFileExtWhenProvided(t *testing.T) {
	t.Setenv("FILE_EXT", "json")
	t.Setenv("FILE_FORMAT", "json_structured")

	got, err := ResolveFromEnv("FILE_EXT", "FILE_FORMAT")
	if err != nil {
		t.Fatalf("ResolveFromEnv returned error: %v", err)
	}

	want := []string{"json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extensions mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestResolveFromEnv_FallsBackToFileFormat(t *testing.T) {
	t.Setenv("FILE_EXT", "")
	t.Setenv("FILE_FORMAT", "json")

	got, err := ResolveFromEnv("FILE_EXT", "FILE_FORMAT")
	if err != nil {
		t.Fatalf("ResolveFromEnv returned error: %v", err)
	}

	want := []string{"json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extensions mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestResolveFromEnv_NormalizesExtensions(t *testing.T) {
	t.Setenv("FILE_EXT", `
.JSON
.strings
  STRINGSdict  
`)
	t.Setenv("FILE_FORMAT", "")

	got, err := ResolveFromEnv("FILE_EXT", "FILE_FORMAT")
	if err != nil {
		t.Fatalf("ResolveFromEnv returned error: %v", err)
	}

	want := []string{"json", "strings", "stringsdict"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extensions mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestResolveFromEnv_TrimsFileFormatFallback(t *testing.T) {
	t.Setenv("FILE_EXT", "")
	t.Setenv("FILE_FORMAT", "  YAML  ")

	got, err := ResolveFromEnv("FILE_EXT", "FILE_FORMAT")
	if err != nil {
		t.Fatalf("ResolveFromEnv returned error: %v", err)
	}

	want := []string{"yaml"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extensions mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestResolveFromEnv_ReturnsErrorWhenCannotInfer(t *testing.T) {
	t.Setenv("FILE_EXT", "")
	t.Setenv("FILE_FORMAT", "   ")

	got, err := ResolveFromEnv("FILE_EXT", "FILE_FORMAT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got != nil {
		t.Fatalf("expected nil extensions on error, got %#v", got)
	}

	if !strings.Contains(err.Error(), "cannot infer file extension") {
		t.Fatalf("expected cannot infer error, got %q", err.Error())
	}

	if !strings.Contains(err.Error(), "FILE_EXT") || !strings.Contains(err.Error(), "FILE_FORMAT") {
		t.Fatalf("expected error to mention env names, got %q", err.Error())
	}
}
