package controllers

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/sijoma/cave-canem/views"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func RolebindingWatcher(clientset *kubernetes.Clientset, log logr.Logger) {
	mux := &sync.RWMutex{}
	synced := false
	log.Info("starting controller watch")
	watchList := cache.NewListWatchFromClient(clientset.RbacV1().RESTClient(), "rolebindings", "", fields.Everything())
	_, controller := cache.NewInformer(watchList, &rbacv1.RoleBinding{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Info("rolebinding added: %s \n", "name", obj.(*rbacv1.RoleBinding).Subjects[0].Name)
				views.AddRoleBinding("added", "test", obj.(*rbacv1.RoleBinding))
			},
			DeleteFunc: func(obj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Info("rolebinding deleted: %s \n", obj.(*rbacv1.RoleBinding).Subjects[0].Name)
				views.AddRoleBinding("deleted", "test", obj.(*rbacv1.RoleBinding))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				mux.RLock()
				defer mux.RUnlock()
				if !synced {
					return
				}

				log.Info("rolebinding changed \n")
				views.AddRoleBinding("modified - OLD", "test", oldObj.(*rbacv1.RoleBinding))
				views.AddRoleBinding("modified - NEW", "test", newObj.(*rbacv1.RoleBinding))
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
		log.Info("failed to sync controller cache")
	}

	<-ctx.Done()
}
