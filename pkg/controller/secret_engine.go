package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	"kubevault.dev/operator/apis"
	vsapis "kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/engine"
)

const (
	SecretEnginePhaseSuccess    api.SecretEnginePhase = "Success"
	SecretEngineConditionFailed                       = "Failed"
	SecretEngineFinalizer                             = "secretengine.engine.kubevault.com"
)

func (c *VaultController) initSecretEngineWatcher() {
	c.secretEngineInformer = c.extInformerFactory.Engine().V1alpha1().SecretEngines().Informer()
	c.secretEngineQueue = queue.New(api.ResourceKindSecretEngine, c.MaxNumRequeues, c.NumThreads, c.runSecretEngineInjector)
	c.secretEngineInformer.AddEventHandler(queue.NewObservableHandler(c.secretEngineQueue.GetQueue(), apis.EnableStatusSubresource))
	c.secretEngineLister = c.extInformerFactory.Engine().V1alpha1().SecretEngines().Lister()
}

func (c *VaultController) runSecretEngineInjector(key string) error {
	obj, exist, err := c.secretEngineInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("SecretEngine %s does not exist anymore", key)

	} else {
		secretEngine := obj.(*api.SecretEngine).DeepCopy()

		glog.Infof("Sync/Add/Update for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)

		if secretEngine.DeletionTimestamp != nil {
			if core_util.HasFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer) {
				go c.runSecretEngineFinalizer(secretEngine, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchSecretEngine(c.extClient.EngineV1alpha1(), secretEngine, func(engine *api.SecretEngine) *api.SecretEngine {
					engine.ObjectMeta = core_util.AddFinalizer(engine.ObjectMeta, SecretEngineFinalizer)
					return engine
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set SecretEngine finalizer for %s/%s", secretEngine.Namespace, secretEngine.Name)
				}
			}

			secretEngineClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, secretEngine)
			if err != nil {
				return err
			}
			err = c.reconcileSecretEngine(secretEngineClient, secretEngine)
			if err != nil {
				return errors.Wrapf(err, "for SecretEngine %s/%s:", secretEngine.Namespace, secretEngine.Name)
			}
		}
	}
	return nil
}

//	For vault:
//	  - enable the secrets engine if it is not already enabled
//	  - configure Vault secret engine
func (c *VaultController) reconcileSecretEngine(secretEngineClient *engine.SecretEngine, secretEngine *api.SecretEngine) error {
	status := secretEngine.Status

	// enable the secret engine if it is not already enabled
	err := secretEngineClient.EnableSecretEngine()
	if err != nil {
		status.Conditions = []api.SecretEngineCondition{
			{
				Type:    SecretEngineConditionFailed,
				Status:  core.ConditionTrue,
				Reason:  "FailedToEnableSecretEngine",
				Message: err.Error(),
			},
		}
		err2 := c.updatedSecretEngineStatus(&status, secretEngine)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update secret engine status")
		}
		return errors.Wrap(err, "failed to enable secret engine")
	}

	// Create secret engine config

	return nil
}

func (c *VaultController) updatedSecretEngineStatus(status *api.SecretEngineStatus, secretEngine *api.SecretEngine) error {
	_, err := patchutil.UpdateSecretEngineStatus(c.extClient.EngineV1alpha1(), secretEngine, func(s *api.SecretEngineStatus) *api.SecretEngineStatus {
		s = status
		return s
	}, vsapis.EnableStatusSubresource)
	return err
}
func (c *VaultController) runSecretEngineFinalizer(secretEngine *api.SecretEngine, timeout time.Duration, interval time.Duration) {
	if secretEngine == nil {
		glog.Infoln("SecretEngine is nil")
		return
	}

	id := getSecretEngineId(secretEngine)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)

	// Add key to finalizerInfo, it will prevent other go routine to processing for this SecretEngine
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("SecretEngine %s/%s finalizer: attempt %d\n", secretEngine.Namespace, secretEngine.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			secretEngineClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, secretEngine)
			if err != nil {
				glog.Errorf("SecretEngine %s/%s finalizer: %v", secretEngine.Namespace, secretEngine.Name, err)
			} else {
				err = c.finalizeSecretEngine(secretEngineClient)
				if err != nil {
					glog.Errorf("SecretEngine %s/%s finalizer: %v", secretEngine.Namespace, secretEngine.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeSecretEngineFinalizer(secretEngine)
			if err != nil {
				glog.Errorf("SecretEngine %s/%s finalizer: removing finalizer %v", secretEngine.Namespace, secretEngine.Name, err)
			} else {
				break
			}
		}

		select {
		case <-stopCh:
			timeOutOccured = true
		case <-time.After(interval):
		}
		attempt++
	}

	err := c.removeSecretEngineFinalizer(secretEngine)
	if err != nil {
		glog.Errorf("SecretEngine %s/%s finalizer: removing finalizer %v", secretEngine.Namespace, secretEngine.Name, err)
	} else {
		glog.Infof("Removed finalizer for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// ToDo:
func (c *VaultController) finalizeSecretEngine(secretEngineClient *engine.SecretEngine) error {
	//err := gcpRClient.DeleteRole(gcpRole.RoleName())
	//if err != nil {
	//	return errors.Wrap(err, "failed to delete gcp role")
	//}
	return nil
}

func (c *VaultController) removeSecretEngineFinalizer(secretEngine *api.SecretEngine) error {
	m, err := c.extClient.EngineV1alpha1().SecretEngines(secretEngine.Namespace).Get(secretEngine.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchSecretEngine(c.extClient.EngineV1alpha1(), m, func(secretEngine *api.SecretEngine) *api.SecretEngine {
		secretEngine.ObjectMeta = core_util.RemoveFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer)
		return secretEngine
	})
	return err
}

func getSecretEngineId(secretEngine *api.SecretEngine) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceSecretEngine, secretEngine.Namespace, secretEngine.Name)
}