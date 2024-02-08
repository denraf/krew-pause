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

func PauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pause",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: pauseCmdRun,
	}

	return cmd
}

func pauseCmdRun(cmd *cobra.Command, args []string) error {

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
	
	if pauseKind == "deployment" || pauseKind == "deployments" {
		return pauseDeployment(clientset, namespace, pauseName)
	} else if pauseKind == "statefulset" {
		return pauseStatefulSet(clientset, namespace, pauseName)
	} else {
		return errors.New("Invalid kind. Use 'deployment' or 'statefulset'")
	}
}

func pauseDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string) error {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting deployment: %s\n", err)
		os.Exit(1)
	}
    replicas := *deployment.Spec.Replicas
    annotations := deployment.GetAnnotations()
    if annotations == nil {
        annotations = make(map[string]string)
    }
    annotations[annotationKey] = fmt.Sprintf("%d", replicas)
    deployment.SetAnnotations(annotations)

    _, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
    if err != nil {
        fmt.Printf("Error updating deployment annotations: %s\n", err)
        os.Exit(1)
    }
	
	deployment, err = clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
    // Scale down replicas to 0
    *deployment.Spec.Replicas = 0
    _, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
    if err != nil {
        fmt.Printf("Error scaling down deployment: %s\n", err)
        os.Exit(1)
    }

    fmt.Printf("Deployment '%s' in namespace '%s' paused.\n", deploymentName, namespace)
	return nil
}

func pauseStatefulSet(clientset *kubernetes.Clientset, namespace, statefulsetName string) error {
	statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulsetName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting statefulset: %s\n", err)
		os.Exit(1)
	}
	replicas := *statefulset.Spec.Replicas
	annotations := statefulset.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = fmt.Sprintf("%d", replicas)
	statefulset.SetAnnotations(annotations)

	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error updating statefulset annotations: %s\n", err)
		os.Exit(1)
	}

	statefulset, err = clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulsetName, metav1.GetOptions{})
	// Scale down replicas to 0
	*statefulset.Spec.Replicas = 0
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("Error scaling down statefulset: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("StatefulSet '%s' in namespace '%s' paused.\n", statefulsetName, namespace)
	return nil
}