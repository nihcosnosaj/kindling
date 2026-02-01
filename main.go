package main

import (
	"fmt"
	"os"

	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

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
	var workers int
	var clusterName string

	var rootCmd = &cobra.Command{
		Use:   "kindling",
		Short: "Kindling boots up a configurable k8s cluster via kind.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(banner)

			provider := cluster.NewProvider()

			// base config
			conf := &v1alpha4.Cluster{
				Nodes: []v1alpha4.Node{
					{Role: v1alpha4.ControlPlaneRole},
				},
			}

			// add some worker nodes based on cli args
			for i := 0; i < workers; i++ {
				conf.Nodes = append(conf.Nodes, v1alpha4.Node{Role: v1alpha4.WorkerRole})
			}

			fmt.Printf("Topology: 1 Control Plane + %d Worker Nodes\n", workers)
			fmt.Printf("Buidling '%s' cluster now ...\n\n", clusterName)

			if err := provider.Create(
				clusterName,
				cluster.CreateWithV1Alpha4Config(conf),
				cluster.CreateWithDisplayUsage(true),
			); err != nil {
				fmt.Printf("\nError building and deploying cluster: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\n Cluster built successfully. Access via:")
			fmt.Printf("    kubectl cluster-info --context kind-%s\n", clusterName)
		},
	}

	var nukeCmd = &cobra.Command{
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

			fmt.Printf("Nuking %d cluster(s)...\n", len(clusters))
			for _, c := range clusters {
				fmt.Printf("Deleting: %s\n", c)
				if err := provider.Delete(c, ""); err != nil {
					fmt.Printf("Failed to delete %s: %v\n", c, err)
				}
			}
			fmt.Println("All active clusters deleted.")
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check the health of nodes.",
		Run: func(cmd *cobra.Command, args []string) {
			// Load kubeconfig from default location.
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

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

			nodes, err := clientset.CoreV1().Nodes().List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to fetch nodes: %v\n", err)
				return
			}

			// format output so it looks nice
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

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information for Kindling",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Kindling Version: %s\n", Version)
			fmt.Printf("Git Commit:       %s\n", Commit)
			fmt.Printf("Build Date:       %s\n", Date)
		},
	}

	rootCmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of worker nodes to add")
	rootCmd.Flags().StringVarP(&clusterName, "name", "n", "kindling-cluster", "Name of your cluster")
	rootCmd.AddCommand(nukeCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
