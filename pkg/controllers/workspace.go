package controllers

import (
	"context"

	kdmv1alpha1 "github.com/kdm/api/v1alpha1"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
)

type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (c *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	var workspaceObj kdmv1alpha1.Workspace
	if err := c.Client.Get(ctx, req.NamespacedName, &workspaceObj); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	//TODO add finalizer

	// Read ResourceSpec
	err := c.applyWorkspaceResource(ctx, workspaceObj.Resource)
	if err != nil {
		return reconcile.Result{}, err
	}
	// TODO apply InferenceSpec
	// TODO apply TrainingSpec

	// Validate

	return reconcile.Result{}, nil
}

func (c *WorkspaceReconciler) applyWorkspaceResource(ctx context.Context, resource kdmv1alpha1.ResourceSpec) error {
	// Check CandidateNodes, if the count, instance type
	err := c.validateCandidateNodes(ctx, resource.CandidateNodes)
	if err != nil {
		return err
	}
	// Create machine CR
	// if nodeCount == count > ok
	//   else wait
	// Add nodes names to the WorkspaceStatus.WorkerNodes[]

	return nil
}

func (c *WorkspaceReconciler) validateCandidateNodes(ctx context.Context, nodeList []string) error {
	return nil
}

func (c *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kdmv1alpha1.Workspace{}).
		Complete(c)
}
