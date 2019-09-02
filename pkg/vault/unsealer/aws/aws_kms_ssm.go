package aws

import (
	"fmt"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"
)

const (
	ModeAwsKmsSsm = "aws-kms-ssm"
)

type Options struct {
	api.AwsKmsSsmSpec
}

func NewOptions(s api.AwsKmsSsmSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	var args []string
	var cont core.Container

	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultUnsealerContainerName {
			cont = c
		}
	}

	args = append(args, fmt.Sprintf("--mode=%s", ModeAwsKmsSsm))
	if o.KmsKeyID != "" {
		args = append(args, fmt.Sprintf("--aws.kms-key-id=%s", o.KmsKeyID))
	}
	if o.SsmKeyPrefix != "" {
		args = append(args, fmt.Sprintf("--aws.ssm-key-prefix=%s", o.SsmKeyPrefix))
	}
	cont.Args = append(cont.Args, args...)

	var envs []core.EnvVar
	if o.Region != "" {
		envs = append(envs, core.EnvVar{
			Name:  "AWS_REGION",
			Value: o.Region,
		})
	}
	if o.Endpoint != "" {
		envs = append(envs, core.EnvVar{
			Name:  "AWS_KMS_ENDPOINT",
			Value: o.Endpoint,
		})
	}
	if o.CredentialSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "access_key",
				},
			},
		}, core.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "secret_key",
				},
			},
		})
	}

	cont.Env = core_util.UpsertEnvVars(cont.Env, envs...)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(prefix, namespace string) []rbac.Role {
	return nil
}
