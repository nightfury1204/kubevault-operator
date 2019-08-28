package framework

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/kubevault/v1alpha1/util"
)

const (
	vaultVersion = "test-v1.2.0"
)

func (f *Invocation) VaultServer(node int32, bs api.BackendStorageSpec) *api.VaultServer {
	return &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("vault-test"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"test": f.app,
			},
		},
		Spec: api.VaultServerSpec{
			Nodes:   node,
			Version: vaultVersion,
			Backend: bs,
		},
	}
}

func (f *Invocation) VaultServerWithUnsealer(node int32, bs api.BackendStorageSpec, us api.UnsealerSpec) *api.VaultServer {
	vs := f.VaultServer(node, bs)
	vs.Spec.Unsealer = &us
	return vs
}

func (f *Framework) CreateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.CSClient.KubevaultV1alpha1().VaultServers(obj.Namespace).Create(obj)
}

func (f *Framework) GetVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.CSClient.KubevaultV1alpha1().VaultServers(obj.Namespace).Get(obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	in, err := f.GetVaultServer(obj)
	if err != nil {
		return nil, err
	}

	vs, _, err := patchutil.PatchVaultServer(f.CSClient.KubevaultV1alpha1(), in, func(vs *api.VaultServer) *api.VaultServer {
		vs.Spec = obj.Spec
		By(fmt.Sprint(vs.Spec))
		return vs
	})
	return vs, err
}

func (f *Framework) DeleteVaultServerObj(obj *api.VaultServer) error {
	err := f.CSClient.EngineV1alpha1().SecretEngines(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
	if kerr.IsNotFound(err) {
		return nil
	}
	return err
}

func (f *Framework) DeleteVaultServer(meta metav1.ObjectMeta) error {
	err := f.CSClient.KubevaultV1alpha1().VaultServers(meta.Namespace).Delete(meta.Name, deleteInBackground())
	if kerr.IsNotFound(err) {
		return nil
	}
	return err
}

func (f *Framework) EventuallyVaultServer(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultServer {
		obj, err := f.CSClient.KubevaultV1alpha1().VaultServers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
