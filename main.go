package main

import (
	"flag"
	"fmt"
	"github.com/sijoma/cave-canem/views"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
)

func main() {
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
		panic(err.Error())
	}

	// creates the in-cluster config
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	log.Println(clientset)

	startTime := time.Now()
	watchList := cache.NewListWatchFromClient(clientset.RbacV1().RESTClient(), "rolebindings", "", fields.Everything())
	_, controller := cache.NewInformer(watchList, &v1.RoleBinding{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("rolebinding added: %s \n", obj.(*v1.RoleBinding).Subjects)
				// Only send notifications about rolebindings that are added 10 seconds after startup
				if time.Since(startTime) > (time.Second * 10) {
					views.AddRoleBinding("added", "test", obj.(*v1.RoleBinding))
				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("rolebinding deleted: %s \n", obj)
				if time.Since(startTime) > (time.Second * 10) {
					views.AddRoleBinding("deleted", "test", obj.(*v1.RoleBinding))
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("rolebinding changed \n")
				if time.Since(startTime) > (time.Second * 10) {
					views.AddRoleBinding("modified - OLD", "test", oldObj.(*v1.RoleBinding))
					views.AddRoleBinding("modified - NEW", "test", newObj.(*v1.RoleBinding))
				}
			},
		},
	)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second * 1)
	}
}
