package cicd

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	k8scrdClient "github.com/changqings/k8scrd/client"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	LocalDeployBaseDir = "/data/devops/deploy/"
)

type K8sDeploy struct {
	AppName   string
	Group     string
	Namespace string
	WorkDir   string
	Type      string
	Env       string
	Tag       string
	Image     string
	Client    *kubernetes.Clientset
}

func NewK8sDeploy(name, group, namespace, deployType, tag, env string) *K8sDeploy {
	client := k8scrdClient.GetClient()
	image := filepath.Join(TxTcrHost, TxTcrNamespaceDevopsScq, name+":"+env+"-"+tag)
	timeStampStr := "t_" + time.Now().Local().Format(TimeLayOutSet)
	workDir := filepath.Join(LocalDeployBaseDir, namespace, name, timeStampStr)
	return &K8sDeploy{
		AppName:   name,
		Group:     group,
		Namespace: namespace,
		Type:      deployType,
		Env:       env,
		Tag:       tag,
		Image:     image,
		WorkDir:   workDir,
		Client:    client,
	}

}

// if found on k8s update image
// or depoly use devops/cicd project/deploy/ yaml files, like kustomize/helm/operator
func (k *K8sDeploy) DoDeploy() error {

	deploy, err := k.Client.AppsV1().Deployments(k.Namespace).Get(context.Background(), k.AppName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !errors.IsNotFound(err) {
			slog.Error("DoDeploy get deploy", "namespace", k.Namespace, "name", k.AppName, "msg", err)
			return err
		}
	}

	// deploy为结构体指针，不能使用nil判断，可以随便取一个结构体字段的值来进行判断
	if deploy.Name != "" {
		slog.Info("DoDeploy update deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "founded, update image...")
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

		slog.Info("DoDeploy update deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "success.")

	} else {
		slog.Info("DoDeploy not found on k8s cluster", "namespace", k.Namespace, "name", k.AppName, "msg", "use devops/cicd config...")
		// mkdir -p k.Workdir
		err := os.MkdirAll(k.WorkDir, 0755)
		if err != nil {
			slog.Error("deploy mkdir all", "path", k.WorkDir, "msg", err)
			return err
		}

		cloneCICDCmd := NewCmd(
			"git", "clone",
			"-b", "main",
			"--depth=1",
			GitlabCICDRepoAddr, k.WorkDir)

		err = cloneCICDCmd.Run()
		if err != nil {
			slog.Error("deploy clone cicd", "path", k.WorkDir, "msg", err)
			return err
		}

		// do deploy
		k8sDeployYamlPath := filepath.Join(k.WorkDir, "jobs", k.Group, k.AppName, "deploy", k.Env)
		deployCmd := NewCmd(
			"/bin/sh", "-c", "kubectl apply -f ./*.yaml",
		)
		deployCmd.Dir = k8sDeployYamlPath
		if err := deployCmd.Run(); err != nil {
			slog.Error("deploy run", "cmd", deployCmd.String(), "msg", err)
			return err
		}

		slog.Info("DoDeploy create deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "success.")
	}

	return nil
}
