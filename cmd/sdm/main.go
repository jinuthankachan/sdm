package main

import (
	"context"
	"fmt"
	"os"
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

	var generateCmd = &cobra.Command{
		Use:   "generate [proto_files...]",
		Short: "Generate SDM files from proto definitions",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runGenerate,
	}

	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
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

	for _, file := range response.File {
		name := file.GetName()
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
