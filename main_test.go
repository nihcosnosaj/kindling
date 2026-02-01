package main

import (
	"testing"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestCreateClusterConfig(t *testing.T) {
	tests := []struct {
		name          string
		workerCount   int
		expectedNodes int
	}{
		{"Single Node", 0, 1},
		{"Standard Cluster", 2, 3},
		{"Large Cluster", 10, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CreateClusterConfig(tt.workerCount)

			// Check total nodes (control plane plus workers)
			if len(config.Nodes) != tt.expectedNodes {
				t.Errorf("got %d nodes, want %d", len(config.Nodes), tt.expectedNodes)

			}

			// Ensure first node is always control plane node.
			if config.Nodes[0].Role != v1alpha4.ControlPlaneRole {
				t.Errorf("first node role is %s, want %s", config.Nodes[0].Role, v1alpha4.ControlPlaneRole)
			}
		})
	}
}

func TestFlagBinding(t *testing.T) {
	cmd := NewRootCmd()

	cmd.SetArgs([]string{"--workers", "5", "--name", "test-cluster"})

	// Mock the run function to prevent actual cluster creation.
	cmd.Run = func(cmd *cobra.Command, args []string) {}

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command exeuction failed: %v", err)
	}

	// Verify 'workers' flag was parsed correctly.
	w, err := cmd.Flags().GetInt("workers")
	if err != nil {
		t.Fatal(err)
	}
	if w != 5 {
		t.Errorf("expected 5 workers, got %d", w)
	}

	// Verify 'name' flag was parsed correctly.
	n, err := cmd.Flags().GetString("name")
	if err != nil {
		t.Fatal(err)
	}
	if n != "test-cluster" {
		t.Errorf("expected name 'test-cluster', got %s", n)
	}
}
