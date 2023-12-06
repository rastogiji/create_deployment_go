package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Authenticates with the Cluster and creates a client set
func getClientset() *kubernetes.Clientset {
	home := homedir.HomeDir()
	kubeconfig := flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "Location to the Kubeconfig file")

	/* Check whether code is running internally or externally and authenticate accordingly
	if kubeconfig exists {
		build config from kubeconfig
	} else {
		Build Config from SA credentials
	}
	*/
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Error Building config from Kubeconfig: %s\n", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error Building Config: %s\n", err.Error())
		}
	}
	flag.Parse()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error Getting clientset: %s\n", err.Error())
	}

	return clientset
}

func int32Ptr(i int32) *int32 {
	return &i
}

// Building the Deployment Object
func buildDeployment(d *Deployment) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: d.name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: d.depLabels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: d.depLabels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									"cpu":    resource.MustParse("0.5"),
									"memory": resource.MustParse("100Mi"),
								},
								Limits: apiv1.ResourceList{
									"cpu":    resource.MustParse("0.5"),
									"memory": resource.MustParse("100Mi"),
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}
func main() {
	c := getClientset()
	ctx := context.Background()

	name := "screen-recorder"
	labels := map[string]string{
		"app": name,
	}
	d := &Deployment{
		// Uses the default namespace
		namespace: apiv1.NamespaceDefault,
		depLabels: labels,
	}
	d.name = name + "-"

	deployment := buildDeployment(d)
	log.Println("Creating Deployment")
	_, err := c.AppsV1().Deployments(d.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error Creating the Deployment: %s\n", err.Error())
	}
}
