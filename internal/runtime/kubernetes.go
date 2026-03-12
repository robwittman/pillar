package runtime

import (
	"context"
	"fmt"
	"log/slog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesConfig holds configuration for the Kubernetes runtime.
type KubernetesConfig struct {
	Context          string
	Namespace        string
	AgentImage       string
	GRPCExternalAddr string
}

// KubernetesRuntime manages agent Deployments in Kubernetes.
type KubernetesRuntime struct {
	client    kubernetes.Interface
	cfg       KubernetesConfig
	logger    *slog.Logger
}

// NewKubernetesRuntime creates a runtime that manages K8s Deployments.
// It loads kubeconfig via standard rules (KUBECONFIG env, ~/.kube/config).
func NewKubernetesRuntime(cfg KubernetesConfig, logger *slog.Logger) (*KubernetesRuntime, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if cfg.Context != "" {
		overrides.CurrentContext = cfg.Context
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	logger.Info("kubernetes client config",
		"host", restConfig.Host,
		"context", cfg.Context,
		"namespace", cfg.Namespace,
	)

	cs, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}

	return newKubernetesRuntime(cs, cfg, logger), nil
}

// newKubernetesRuntime is the internal constructor that accepts a kubernetes.Interface for testing.
func newKubernetesRuntime(client kubernetes.Interface, cfg KubernetesConfig, logger *slog.Logger) *KubernetesRuntime {
	return &KubernetesRuntime{
		client: client,
		cfg:    cfg,
		logger: logger,
	}
}

func (r *KubernetesRuntime) EnsureRunning(ctx context.Context, agentID string) error {
	name := deploymentName(agentID)
	deploymentsClient := r.client.AppsV1().Deployments(r.cfg.Namespace)

	existing, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		deploy := r.buildDeployment(agentID, 1)
		r.logger.Debug("creating deployment", "name", name, "namespace", r.cfg.Namespace)
		if _, err := deploymentsClient.Create(ctx, deploy, metav1.CreateOptions{}); err != nil {
			r.logger.Error("deployment create failed",
				"name", name,
				"namespace", r.cfg.Namespace,
				"error", err,
				"status_reason", apierrors.ReasonForError(err),
			)
			return fmt.Errorf("create deployment %s: %w", name, err)
		}
		r.logger.Info("created agent deployment", "name", name, "agent_id", agentID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("get deployment %s: %w", name, err)
	}

	// Deployment exists — scale to 1 if needed.
	if existing.Spec.Replicas != nil && *existing.Spec.Replicas == 1 {
		return nil
	}
	one := int32(1)
	existing.Spec.Replicas = &one
	if _, err := deploymentsClient.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("scale deployment %s to 1: %w", name, err)
	}
	r.logger.Info("scaled agent deployment to 1", "name", name, "agent_id", agentID)
	return nil
}

func (r *KubernetesRuntime) EnsureStopped(ctx context.Context, agentID string) error {
	name := deploymentName(agentID)
	deploymentsClient := r.client.AppsV1().Deployments(r.cfg.Namespace)

	existing, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get deployment %s: %w", name, err)
	}

	zero := int32(0)
	existing.Spec.Replicas = &zero
	if _, err := deploymentsClient.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("scale deployment %s to 0: %w", name, err)
	}
	r.logger.Info("scaled agent deployment to 0", "name", name, "agent_id", agentID)
	return nil
}

func (r *KubernetesRuntime) Remove(ctx context.Context, agentID string) error {
	name := deploymentName(agentID)
	deploymentsClient := r.client.AppsV1().Deployments(r.cfg.Namespace)

	err := deploymentsClient.Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete deployment %s: %w", name, err)
	}
	r.logger.Info("deleted agent deployment", "name", name, "agent_id", agentID)
	return nil
}

func (r *KubernetesRuntime) buildDeployment(agentID string, replicas int32) *appsv1.Deployment {
	name := deploymentName(agentID)
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "pillar",
		"pillar.io/agent-id":           agentID,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "agent",
							Image:           r.cfg.AgentImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args:            []string{"-addr", r.cfg.GRPCExternalAddr, "-agent-id", agentID},
						},
					},
				},
			},
		},
	}
}

func deploymentName(agentID string) string {
	short := agentID
	if len(short) > 8 {
		short = short[:8]
	}
	return "pillar-agent-" + short
}
