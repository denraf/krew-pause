package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

const annotationKey = "kubectl-krew-pause/replicas"

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "krew-pause",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: rootCmdRun,
	}
	cmd.PersistentFlags().StringP("namespace", "n", "", "namespace")

	cmd.AddCommand(PauseCmd())
	cmd.AddCommand(UnPauseCmd())

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func rootCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("rootcmd")
	fmt.Println(args)
	fmt.Println(len(args))

	if len(args) != 2 {
		return errors.New("Usage: kubectl -n <namespace> krew-pause {pause|unpause} {deployment/<deployment-name>|statefulset/<statefulset-name>}")
	}
	return nil
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}