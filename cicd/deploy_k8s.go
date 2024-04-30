package cicd

import (
	"context"
	"log/slog"

	k8scrdClient "github.com/changqings/k8scrd/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sDeploy struct {
	AppName   string
	Namespace string
	Type      string
	Env       string
	Tag       string
	Image     string
	Client    *kubernetes.Clientset
}

func NewK8sDeploy(name, namespace, deployType, env, tag, image string) *K8sDeploy {
	client := k8scrdClient.GetClient()
	return &K8sDeploy{
		AppName:   name,
		Namespace: namespace,
		Type:      deployType,
		Env:       env,
		Tag:       tag,
		Image:     image,
		Client:    client,
	}

}

// if found on k8s. kubectl update image
// or depoly use devops/cicd project/deploy/kustomize,
func (k *K8sDeploy) DoDeploy() error {

	deploy, err := k.Client.AppsV1().Deployments(k.Namespace).Get(context.Background(), k.AppName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		return err
	}

	//
	if deploy != nil {
		var appContainerIndex int
		for i, c := range deploy.Spec.Template.Spec.Containers {
			if c.Name == "app" {
				appContainerIndex = i
				break
			}
		}

		deploy.ResourceVersion = ""
		deploy.Spec.Template.Spec.Containers[appContainerIndex].Image = k.Image

		_, err := k.Client.AppsV1().Deployments(k.Namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("update deploy", "namespace", k.Namespace, "name", k.AppName, "msg", err)
			return err
		}
	} else {
		// not exit
		slog.Info("create deploy use devops/cicd kustomize", "namespace", k.Namespace, "name", k.AppName)

	}

	return nil
}
