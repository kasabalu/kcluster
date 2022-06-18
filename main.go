package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"

	klient "github.com/kasabalu/kcluster/pkg/client/clientset/versioned"
	kfactory "github.com/kasabalu/kcluster/pkg/client/informers/externalversions"
	"github.com/kasabalu/kcluster/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}

	klientset, err := klient.NewForConfig(config)
	if err != nil {
		log.Printf("getting klient set %s\n", err.Error())
	}

	fmt.Println(klientset)
	klusters, err := klientset.KasabaluV1alpha1().Klusters("default").List(context.Background(), metav1.ListOptions{})
	fmt.Println(klusters)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("getting std client %s\n", err.Error())
	}

	// NewCntroller in controller pkg expets cleintset and Informer
	// To create informer, call Informer Factory from  factory.go

	ch := make(chan struct{})

	infoFact := kfactory.NewSharedInformerFactory(klientset, 20*time.Minute)
	c := controller.NewController(client, klientset, infoFact.Kasabalu().V1alpha1().Klusters())
	infoFact.Start(ch)
	err = c.Run(ch)

	if err != nil {
		log.Println("error running ccontrollelr", err.Error())
	}

}
