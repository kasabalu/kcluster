package controller

import (
	"context"
	"fmt"
	"github.com/kasabalu/kcluster/pkg/apis/kasabalu.dev/v1alpha1"
	klister "github.com/kasabalu/kcluster/pkg/client/listers/kasabalu.dev/v1alpha1"
	"github.com/kasabalu/kcluster/pkg/do"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"

	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	klientset "github.com/kasabalu/kcluster/pkg/client/clientset/versioned"
	kinformer "github.com/kasabalu/kcluster/pkg/client/informers/externalversions/kasabalu.dev/v1alpha1"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client kubernetes.Interface
	// clientset for custom resource kluster

	klient klientset.Interface

	// way to figure out cache has been synced or updated

	klusterSynced cache.InformerSynced

	// lister
	kLister klister.KlusterLister
	// queue
	wq workqueue.RateLimitingInterface
}

func NewController(client kubernetes.Interface, klient klientset.Interface, klusterInformer kinformer.KlusterInformer) *Controller {
	// to intilize the controller, client set and informers are required, take that as method parameters

	c := &Controller{
		client:        client,
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
	}
	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)
	return c
}

func (c *Controller) Run(ch <-chan struct{}) error {
	// make sure cache that informer maintain initilizes atlease once.
	// Informer maintain cache, making sure Informer cache synced successfully
	if !cache.WaitForCacheSync(ch, c.klusterSynced) {
		fmt.Println("waiting for cache to be synced")
	}

	// run go routine to consume events from queue

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch

	return nil
}

func (c *Controller) worker() {
	for c.processItem() {
		// This loop will terminate when processItem return False.
		//worker func will called every one sec from Util until channel is closed.
	}
}

func (c *Controller) processItem() bool {

	item, shutdown := c.wq.Get() // getting the obj from Queue
	if shutdown {
		return false
	}
	defer c.wq.Forget(item) // this will delete/mark as proecces item from queue, make sure not processed again.
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("getting key from cahce %s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key) // getting namespace and name
	if err != nil {
		fmt.Printf("splitting key into namespace and name %s\n", err.Error())
		return false
	}

	kluster, err := c.kLister.Klusters(ns).Get(name)

	if err != nil {
		// re-try
		fmt.Printf("getting kluster resource from Kluster %s\n", err.Error())
		return false
	}
	log.Printf("Kluster specs %v", kluster.Spec)

	clusterID, err := do.Create(c.client, kluster.Spec)
	if err != nil {
		// do something
		log.Printf("errro %s, creating the cluster", err.Error())
	}

	log.Printf("cluster id that we have is %s\n", clusterID)

	err = c.updateStatus(clusterID, "creating", kluster)
	if err != nil {
		log.Printf("error %s, updating status of the kluster %s\n", err.Error(), kluster.Name)
	}

	return true
}

func (c *Controller) updateStatus(id, progress string, kluster *v1alpha1.Kluster) error {
	// get the latest version of kluster
	k, err := c.klient.KasabaluV1alpha1().Klusters(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	k.Status.KlusterID = id
	k.Status.Progress = progress
	_, err = c.klient.KasabaluV1alpha1().Klusters(kluster.Namespace).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})
	return err
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Printf("handle Add was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handle Del was called")
	c.wq.Add(obj)

}
