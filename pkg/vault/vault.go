/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vault

import (
	"context"

	vaultauth "kubevault.dev/operator/pkg/vault/auth"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	vaultutil "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func NewClient(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, vAppRef *appcat.AppReference) (*vaultapi.Client, error) {

	vApp, err := appc.AppBindings(vAppRef.Namespace).Get(context.TODO(), vAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	authInfo, err := authtype.GetAuthInfoFromAppBinding(kc, vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authentication information")
	}

	return NewClientWithAppBinding(kc, authInfo)
}

func NewClientWithAppBinding(kc kubernetes.Interface, authInfo *authtype.AuthInfo) (*vaultapi.Client, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}
	if authInfo.VaultApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	auth, err := vaultauth.NewAuth(kc, authInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth method")
	}

	token, err := auth.Login()
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	cfg, err := vaultutil.VaultConfigFromAppBinding(authInfo.VaultApp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vault client config")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	vc.SetToken(token)
	return vc, nil
}
