package main

import (
	"fmt"

	"github.com/jinghzhu/KubernetesCRD/pkg/config"
	"github.com/jinghzhu/KubernetesCRD/pkg/types"
	"k8s.io/client-go/tools/clientcmd"

	crdjinghzhuv1 "github.com/jinghzhu/KubernetesCRD/pkg/crd/jinghzhu/v1"
	jinghzhuv1client "github.com/jinghzhu/KubernetesCRD/pkg/crd/jinghzhu/v1/client"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	ctx := types.GetCtx()
	cfg := config.GetConfig()
	kubeconfigPath := cfg.GetKubeconfigPath()

	// Use kubeconfig to create client config.
	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}

	apiextensionsClientSet, err := apiextensionsclient.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	// Init a CRD kind.
	if _, err = crdjinghzhuv1.CreateCustomResourceDefinition(apiextensionsClientSet); err != nil {
		panic(err)
	}

	// Create a CRD client interface for Jinghzhu v1.
	crdClient, err := jinghzhuv1client.NewClient(ctx, kubeconfigPath, cfg.GetCRDNamespace())
	if err != nil {
		panic(err)
	}

	// Create an instance of CRD.
	instanceName := "jinghzhu-example-"
	exampleInstance := &crdjinghzhuv1.Jinghzhu{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: instanceName,
		},
		Spec: crdjinghzhuv1.JinghzhuSpec{
			Desired: 1,
			Current: 0,
			PodList: make([]string, 0),
		},
		Status: crdjinghzhuv1.JinghzhuStatus{
			State:   types.StatePending,
			Message: "Created but not processed yet",
		},
	}
	result, err := crdClient.CreateDefault(exampleInstance)
	if err != nil && apierrors.IsAlreadyExists(err) {
		fmt.Printf("ALREADY EXISTS: %#v\n", result)
	} else if err != nil {
		panic(err)
	}
	crdInstanceName := result.GetName()
	fmt.Println("CREATED: " + result.String())

	// Wait until the CRD object is handled by controller and its status is changed to Processed.
	err = crdClient.WaitForInstanceProcessed(crdInstanceName)
	if err != nil {
		panic(err)
	}
	fmt.Println("Processed " + crdInstanceName)

	// Get the list of CRs.
	exampleList, err := crdClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("LIST: %#v\n", exampleList)
}
