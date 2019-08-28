package framework

import (
	"fmt"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"kubevault.dev/operator/apis/kubevault/v1alpha1"
)

const (
	updatedNodePort = 30089
)

func (f *Framework) UpdateServiceAndAppBinding(vServer *v1alpha1.VaultServer) error {
	svc, err := f.KubeClient.CoreV1().Services(vServer.Namespace).Get(vServer.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	label := vServer.OffshootLabels()
	svc.Spec = core.ServiceSpec{
		Selector: label,
		Ports: []core.ServicePort{
			{
				Name:     "http",
				Protocol: core.ProtocolTCP,
				Port:     8200,
				NodePort: updatedNodePort,
			},
		},
		Type: core.ServiceTypeNodePort,
	}
	_, err = f.KubeClient.CoreV1().Services(vServer.Namespace).Update(svc)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to update service:%s/%s", vServer.Namespace, vServer.Name))
	}

	vApp, err := f.AppcatClient.AppBindings(vServer.Namespace).Get(vServer.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	nodePortIP, err := f.GetNodePortIP(label)
	if err != nil {
		return errors.Wrap(err, "failed to get NodePortIP")
	}
	url := fmt.Sprintf("http://%s:%d", nodePortIP, updatedNodePort)

	vApp.Spec.ClientConfig = appcat.ClientConfig{
		URL:                   &url,
		InsecureSkipTLSVerify: true,
	}
	_, err = f.AppcatClient.AppBindings(vServer.Namespace).Update(vApp)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to update AppBinding:%s/%s", vServer.Namespace, vServer.Name))
	}
	return nil
}
