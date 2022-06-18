package controller

import (
	klister "github.com/kasabalu/kcluster/pkg/client/listers/v1alpha1/internalversion"
	"google.golang.org/appengine/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	klientset "github.com/kasabalu/kcluster/pkg/client/clientset/versioned"
	kinformer "github.com/kasabalu/kcluster/pkg/client/informers/internalversion/v1alpha1/internalversion"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	// clientset for custom resource kluster

	klient klientset.Interface

	// way to figure out cache has been synced or updated

	klusterSynced cache.InformerSynced

	// lister
	kLister klister.KlusterLister
	// queue
	wq workqueue.RateLimitingInterface
}

func NewController(klient klientset.Interface, klusterInformer kinformer.KlusterInformer) *Controller {
	// to intilize the controller, client set and informers are required, take that as method parameters

	c := &Controller{
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

	return false
}

func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handle Add was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handle Del was called")
	c.wq.Add(obj)

}
