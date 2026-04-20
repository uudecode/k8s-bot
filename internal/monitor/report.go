package monitor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (s *Service) SendStartupReport() error {
	var report string
	report += "🚀 *Bot started and ready!*\n\n"

	nodes, err := s.K8s.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	report += "🖥 *Node status:*\n"
	for _, node := range nodes.Items {
		status := "✅ Ready"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status != "True" {
				status = "❌ NotReady"
			}
		}
		report += "- " + node.Name + ": " + status + "\n"
	}

	report += "\n📦 *ArgoCD applications:*\n"
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	list, err := s.Dynamic.Resource(gvr).Namespace(s.Cfg.ArgoCD.Name).List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			syncStatus := item.Object["status"].(map[string]interface{})["sync"].(map[string]interface{})["status"].(string)
			report += "- " + item.GetName() + ": " + syncStatus + "\n"
		}
	}

	return s.Notifier.Send(report)
}
