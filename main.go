package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/sijoma/cave-canem/controllers"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
)

var log logr.Logger

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Error(err, "error building config")
		return
	}

	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "error building client")
	}

	log.Info("starting controllers")
	controllers.RolebindingWatcher(clientset, log)
}
