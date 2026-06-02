package monitor

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (s *Service) BuildStatusReport(ctx context.Context) (string, error) {
	var report string
	report += "🚀 *Bot started and ready!*\n\n"

	nodes, err := s.K8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list nodes: %w", err)
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

	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	list, err := s.Dynamic.Resource(gvr).Namespace(s.Cfg.ArgoCD.Name).List(ctx, metav1.ListOptions{})
	if err != nil {
		report += "- failed to list applications: `" + err.Error() + "`\n"
		return report, nil
	}

	for _, item := range list.Items {
		syncStatus := "unknown"

		status, ok := item.Object["status"].(map[string]interface{})
		if ok {
			syncObj, ok := status["sync"].(map[string]interface{})
			if ok {
				if value, ok := syncObj["status"].(string); ok && value != "" {
					syncStatus = value
				}
			}
		}

		report += "- " + item.GetName() + ": " + syncStatus + "\n"
	}

	return report, nil
}

func (s *Service) SendStartupReport(ctx context.Context) error {
	report, err := s.BuildStatusReport(ctx)
	if err != nil {
		return err
	}

	return s.Notifier.Send(report)
}
