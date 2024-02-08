package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func UnPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "unpause",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: unPauseCmdRun,
	}

	return cmd
}

func unPauseCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: kubectl pause -n <namespace> {deployment/<deployment-name>|statefulset/<statefulset-name>}")
	}

	namespace, err := cmd.Flags().GetString("namespace")

	pauseArg := strings.Split(args[0], "/")
	if len(pauseArg) != 2 {
		return errors.New("Invalid format for namespace/deployment. Please use <namespace>/<deployment-name>")
	}

	pauseKind, pauseName := pauseArg[0], pauseArg[1]

    // Load kubeconfig
    kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
    config, err := kubeconfig.ClientConfig()
    if err != nil {
        fmt.Printf("Error getting kubeconfig: %s\n", err)
        os.Exit(1)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Printf("Error creating clientset: %s\n", err)
        os.Exit(1)
    }
	
	if pauseKind == "deployment" {
		return unPauseDeployment(clientset, namespace, pauseName)
	} else if pauseKind == "statefulset" {
		return unPauseStatefulSet(clientset, namespace, pauseName)
	} else {
		return errors.New("Invalid kind. Use 'deployment' or 'statefulset'")
	}
}

func unPauseDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) error {
    deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
    if err != nil {
        fmt.Printf("Error getting deployment: %s\n", err)
        os.Exit(1)
    }

    annotations := deployment.GetAnnotations()
    replicasStr, ok := annotations[annotationKey]
    if !ok {
        fmt.Println("No pause annotation found on deployment. Cannot unpause.")
        os.Exit(1)
    }

    var replicas int32
    fmt.Sscanf(replicasStr, "%d", &replicas)

    // Scale up to original replicas
    deployment.Spec.Replicas = &replicas
    _, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error scaling up deployment: %s\n", err)
		os.Exit(1)
	}
	
    deployment, err = clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	// Remove pause annotation
	delete(annotations, annotationKey)
	deployment.SetAnnotations(annotations)
	_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error removing pause annotation: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deployment '%s' in namespace '%s' unpaused.\n", deploymentName, namespace)
	return nil
}

func unPauseStatefulSet(clientset *kubernetes.Clientset, namespace, statefulsetName string) error {
	statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulsetName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting statefulset: %s\n", err)
		os.Exit(1)
	}

	annotations := statefulset.GetAnnotations()
	replicasStr, ok := annotations[annotationKey]
	if !ok {
		fmt.Println("No pause annotation found on statefulset. Cannot unpause.")
		os.Exit(1)
	}

	var replicas int32
	fmt.Sscanf(replicasStr, "%d", &replicas)

	// Scale up to original replicas
	statefulset.Spec.Replicas = &replicas
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error scaling up statefulset: %s\n", err)
		os.Exit(1)
	}

	statefulset, err = clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulsetName, metav1.GetOptions{})
	// Remove pause annotation
	delete(annotations, annotationKey)
	statefulset.SetAnnotations(annotations)
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error removing pause annotation: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Statefulset '%s' in namespace '%s' unpaused.\n", statefulsetName, namespace)
	return nil
}