package main

import (
	"context"
	"flag"
	"github.com/sijoma/cave-canem/views"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
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

	mux := &sync.RWMutex{}
	synced := false

	log.Println("starting controller watch")
	watchList := cache.NewListWatchFromClient(clientset.RbacV1().RESTClient(), "rolebindings", "", fields.Everything())
	_, controller := cache.NewInformer(watchList, &v1.RoleBinding{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Printf("rolebinding added: %s \n", obj.(*v1.RoleBinding).Subjects)
				views.AddRoleBinding("added", "test", obj.(*v1.RoleBinding))
			},
			DeleteFunc: func(obj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Printf("rolebinding deleted: %s \n", obj)
				views.AddRoleBinding("deleted", "test", obj.(*v1.RoleBinding))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Printf("rolebinding changed \n")
				views.AddRoleBinding("modified - OLD", "test", oldObj.(*v1.RoleBinding))
				views.AddRoleBinding("modified - NEW", "test", newObj.(*v1.RoleBinding))
			},
		},
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go controller.Run(ctx.Done())

	isSynced := cache.WaitForCacheSync(ctx.Done(), controller.HasSynced)
	mux.Lock()
	synced = isSynced
	mux.Unlock()

	if !isSynced {
		log.Fatal("failed to sync controller cache")
	}

	<-ctx.Done()
}
