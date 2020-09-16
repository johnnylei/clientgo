package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path"
	"time"
)

const (
	DefaultNameSpace = "default"
)

func main() {
	configPath := path.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		fmt.Printf("%s", err.Error())
		time.Sleep(time.Minute*100)
		return
	}
	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	result := &corev1.PodList{}
	err = restClient.Get().Namespace(DefaultNameSpace).
		Resource("pods").VersionedParams(&metav1.ListOptions{Limit:500}, scheme.ParameterCodec).
		Do().Into(result)
	if err != nil {
		panic(err)
	}

	for _, item := range result.Items {
		fmt.Printf("NAMESPCE:%v \t NAME:%v \t STATUS:%+v\n", item.Namespace, item.Name, item.Status.Phase)
	}

	nodeList := &corev1.NodeList{}
	err = restClient.Get().Resource("nodes").VersionedParams(&metav1.ListOptions{Limit:500}, scheme.ParameterCodec).
		Do().Into(nodeList)
	for _, item := range nodeList.Items {
		fmt.Printf("NAME:%v \t STATUS:Ready:%+v\n", item.Name, item.Status.Conditions[len(item.Status.Conditions)-1].Status)
	}

	node := &corev1.Node{}
	err = restClient.Get().Resource("nodes").Name("node3").
		Do().Into(node)
	fmt.Printf("Name:%v\t CPU:%v\n", node.Name, node.Status.Capacity.Cpu())

	clientSet, _ := kubernetes.NewForConfig(config)
	podList, _ := clientSet.CoreV1().Pods(corev1.NamespaceDefault).List(metav1.ListOptions{Limit:10})

	for _, item := range podList.Items {
		fmt.Printf("NAMESPCE:%v \t NAME:%v \t STATUS:%+v\n", item.Namespace, item.Name, item.Status.Phase)
	}

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)
	apiGroups, _, _ := discoveryClient.ServerGroupsAndResources()

	for _, item := range apiGroups {
		fmt.Printf("%s\n", item.Name)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformers := informers.NewSharedInformerFactory(clientSet, time.Minute)
	informer := sharedInformers.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mObject := obj.(metav1.Object)
			fmt.Printf("%s added\n", mObject.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			mObject := obj.(metav1.Object)
			fmt.Printf("%s deleted\n", mObject.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
	})
	informer.Run(stopCh)
}
