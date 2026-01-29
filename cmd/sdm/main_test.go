package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigCmd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sdm-test-config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config command failed: %v", err)
	}

	if _, err := os.Stat("sdm.cfg.yaml"); os.IsNotExist(err) {
		t.Error("sdm.cfg.yaml not created")
	}
}

func TestGenerateCmd(t *testing.T) {
	// Locate project root to find proto files
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	protoDir := filepath.Join(projectRoot, "proto")

	tmpDir, err := os.MkdirTemp("", "sdm-test-gen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// User proto
	userProtoContent := `
syntax = "proto3";
package test;
import "sdm/annotations.proto";
option go_package = "example/test";

message TestMessage {
  string id = 1 [(sdm.primary_key) = true];
}
`
	testProtoPath := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(testProtoPath, []byte(userProtoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Config file pointing to existing protos
	// We need config to specify sdm-proto if we want to rely on setup logic,
	// OR we rely on flags.
	// But runGenerate logic: "sdmProtoDir := cfg.SdmProto ... ImportPaths := ... sdmProtoDir"
	// So we need to provide a config that points to the sdm protos.

	configContent := "sdm-proto: " + protoDir + "\n"
	configPath := filepath.Join(tmpDir, "sdm.cfg.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Use the command
	// We need to change dir to tmpDir potentially, so "." resolves correctly?
	// Or just use absolute paths.
	// runGenerate uses "files, err := compiler.Compile(..., filesToGenerate...)"
	// and "ImportPaths... sdmProtoDir".

	// We run the command from the test function, so we might need to reset flags?
	// Cobra flags are persistent on the command struct. newRootCmd creates valid fresh commands.

	cmd := newRootCmd()
	// Set flags
	cmd.SetArgs([]string{
		"generate",
		"--proto", testProtoPath,
		"--out", filepath.Join(tmpDir, "gen"),
		"--cfg", configPath,
	})

	// Also we need `protoc-gen-go` to be found.
	// As we improved the search logic, it should find it in GOPATH if set.
	// But in test environment, we assume `go test` runs in an environment where things are set.

	if err := cmd.Execute(); err != nil {
		t.Fatalf("generate command failed: %v", err)
	}

	// Check generated files
	// Since we passed basename 'test.proto' and used paths=source_relative,
	// the output matches the input structure relative to the import path.
	expectedFile := filepath.Join(tmpDir, "gen", "test.pb.go")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected generated file %s not found", expectedFile)
	}
}
