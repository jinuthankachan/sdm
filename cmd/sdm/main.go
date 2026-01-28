package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bufbuild/protocompile"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/jinuthankachan/sdm/pkg/generator"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sdm",
		Short: "SDM Tool",
	}

	var outputDir string

	var generateCmd = &cobra.Command{
		Use:   "generate [proto_files...]",
		Short: "Generate SDM files from proto definitions",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd, args, outputDir)
		},
	}

	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated files")
	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string, outputDir string) error {
	// 1. Compile Proto Files
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: []string{".", "proto", "vendor"}, // Add commonly used paths
		}),
	}

	files, err := compiler.Compile(context.Background(), args...)
	if err != nil {
		return fmt.Errorf("failed to compile protos: %w", err)
	}

	// 2. Construct CodeGeneratorRequest
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: args,
		Parameter:      proto.String("paths=source_relative"), // Default option
	}

	// Collect all descriptors (files to generate + imports)
	// We need to topological sort or just dump them. protogen handles lookup.
	// But CodeGeneratorRequest expects topological order usually, or dependencies first.
	// protocompile returns linked files.

	seen := make(map[string]bool)
	var collect func(f protoreflect.FileDescriptor)
	collect = func(f protoreflect.FileDescriptor) {
		if seen[f.Path()] {
			return
		}
		seen[f.Path()] = true
		for i := 0; i < f.Imports().Len(); i++ {
			collect(f.Imports().Get(i))
		}
		req.ProtoFile = append(req.ProtoFile, protodesc.ToFileDescriptorProto(f))
	}

	for _, f := range files {
		collect(f)
	}

	// 3. Create Plugin
	opts := protogen.Options{}
	gen, err := opts.New(req)
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	// 4. Run Generation Logic
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		generator.GenerateFile(gen, f)
	}

	// 5. Write Responses
	response := gen.Response()
	if response.Error != nil {
		return fmt.Errorf("generator error: %s", response.GetError())
	}

	// 6. Run protoc-gen-go
	if err := runProtocGenGo(req, response); err != nil {
		// Just warn? No, let's return error but make it clear.
		return fmt.Errorf("failed to run protoc-gen-go (ensure it is in PATH): %w", err)
	}

	for _, file := range response.File {
		name := file.GetName()

		if outputDir != "" {
			name = filepath.Join(outputDir, name)
		}

		content := file.GetContent()

		// Ensure dir exists
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", name, err)
		}
		fmt.Printf("Generated: %s\n", name)
	}

	return nil
}

func runProtocGenGo(req *pluginpb.CodeGeneratorRequest, resp *pluginpb.CodeGeneratorResponse) error {
	// Marshal request
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	cmd := exec.Command("protoc-gen-go")
	cmd.Stdin = bytes.NewReader(data)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if os.IsNotExist(err) || err.Error() == "exec: \"protoc-gen-go\": executable file not found in $PATH" {
			// Ignore if not present? No, user wants it.
			return fmt.Errorf("protoc-gen-go executable not found in $PATH")
		}
		return fmt.Errorf("exec error: %v, stderr: %s", err, stderr.String())
	}

	// Unmarshal response
	var goResp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(out.Bytes(), &goResp); err != nil {
		return err
	}

	if goResp.Error != nil {
		return fmt.Errorf("protoc-gen-go error: %s", goResp.GetError())
	}

	// Merge files
	resp.File = append(resp.File, goResp.File...)
	return nil
}
