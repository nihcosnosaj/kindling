package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

// Versioning (injected via ldflags)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

const banner = `
  _  _____ _   _ _____  _     ___ _   _  ____ 
 | |/ /_ _| \ | |  _  \| |   |_ _| \ | |/ ___|
 | ' / | ||  \| | | |  | |    | ||  \| | |  _ 
 |  \  | || |\  | |_|  | |___ | || |\  | |_| |
 |_| \_\___|_| \_|____/|_____|___|_| \_|\____|
`

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewRootCmd initializes the CLI and binds all subcommands
func NewRootCmd() *cobra.Command {
	var workers int
	var clusterName string

	cmd := &cobra.Command{
		Use:   "kindling",
		Short: "Ephemeral local Kubernetes clusters",
		Run: func(cmd *cobra.Command, args []string) {
			provider := cluster.NewProvider()
			fmt.Print(banner)

			conf := CreateClusterConfig(workers)

			fmt.Printf("Starting up %s (1 CP, %d Workers)...\n", clusterName, workers)
			if err := provider.Create(
				clusterName,
				cluster.CreateWithV1Alpha4Config(conf),
				cluster.CreateWithDisplayUsage(true),
				cluster.CreateWithDisplaySalutation(true),
			); err != nil {
				fmt.Printf("Cluster startup failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\nCluster is hot! Use 'kindling status' to check health.")
		},
	}

	cmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of worker nodes")
	cmd.Flags().StringVarP(&clusterName, "name", "n", "kindling-cluster", "Cluster name")

	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newNukeCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the health of nodes",
		Run: func(cmd *cobra.Command, args []string) {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

			config, err := kubeConfig.ClientConfig()
			if err != nil {
				fmt.Printf("Could not find kubeconfig: %v\n", err)
				return
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				fmt.Printf("Connection error: %v\n", err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to fetch nodes: %v\n", err)
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NODE NAME\tROLE\tSTATUS\tINTERNAL IP")
			for _, node := range nodes.Items {
				role := "worker"
				if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
					role = "control-plane"
				}

				status := "NotReady"
				for _, cond := range node.Status.Conditions {
					if cond.Type == "Ready" && cond.Status == "True" {
						status = "Ready"
					}
				}

				ip := "Unknown"
				for _, addr := range node.Status.Addresses {
					if addr.Type == "InternalIP" {
						ip = addr.Address
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", node.Name, role, status, ip)
			}
			w.Flush()
		},
	}
}

func newNukeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nuke",
		Short: "Brings down all active Kindling clusters",
		Run: func(cmd *cobra.Command, args []string) {
			provider := cluster.NewProvider()
			clusters, err := provider.List()
			if err != nil {
				fmt.Printf("Failed to list clusters: %v\n", err)
				return
			}

			if len(clusters) == 0 {
				fmt.Println("No active clusters found. Nothing to do.")
				return
			}

			fmt.Printf("🧨 Nuking %d cluster(s)...\n", len(clusters))
			for _, c := range clusters {
				fmt.Printf("Deleting: %s\n", c)
				_ = provider.Delete(c, "")
			}
			fmt.Println("All active clusters nuked.")
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kindling Version: %s\n", Version)
			fmt.Printf("Git Commit:       %s\n", Commit)
			fmt.Printf("Build Date:       %s\n", Date)
		},
	}
}

func CreateClusterConfig(workerCount int) *v1alpha4.Cluster {
	conf := &v1alpha4.Cluster{
		Nodes: []v1alpha4.Node{{Role: v1alpha4.ControlPlaneRole}},
	}
	for i := 0; i < workerCount; i++ {
		conf.Nodes = append(conf.Nodes, v1alpha4.Node{Role: v1alpha4.WorkerRole})
	}
	return conf
}
