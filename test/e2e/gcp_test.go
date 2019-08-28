package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	kvapi "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/test/e2e/framework"
)

var _ = Describe("GCP Role", func() {

	var f *framework.Invocation
	var vServer *kvapi.VaultServer
	const (
		VaultKey = "vault-keys-23432"
	)

	var (
		IsVaultServerCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether VaultServer:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "VaultServer is created")
		}

		IsVaultServerDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether VaultServer:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "VaultServer is deleted")
		}

		IsAppBindingCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AppBinding:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.AppcatClient.AppBindings(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "AppBinding is created")
		}

		IsAppBindingDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AppBinding:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.AppcatClient.AppBindings(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AppBinding is deleted")
		}

		IsVaultGCPRoleCreated = func(name string) {
			By("Checking whether vault gcp role is created")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, &v1alpha1.AppReference{
				Name:      vServer.Name,
				Namespace: vServer.Namespace,
			})

			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/gcp/roleset/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault gcp role is created")

		}

		IsVaultGCPRoleDeleted = func(name string) {
			By("Checking whether vault gcp role is deleted")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, &v1alpha1.AppReference{
				Name:      vServer.Name,
				Namespace: vServer.Namespace,
			})
			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/gcp/roleset/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err != nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault gcp role is deleted")

		}

		IsSecretEngineCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is created")
		}
		IsSecretEngineDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is deleted")
		}
		IsSecretEngineSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.SecretEnginePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine status is succeeded")

		}
		IsGCPRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is created")
		}
		IsGCPRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is deleted")
		}
		IsGCPRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.GCPRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole status is succeeded")

		}
		IsGCPAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is created")
		}
		IsGCPAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is deleted")
		}
		IsGCPAKRConditionApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequestConditions-> Type: Approved"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == api.AccessApproved {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type : Approved")
		}
		IsGCPAKRConditionDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequestConditions-> Type: Denied"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == api.AccessDenied {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type: Denied")
		}
		IsGCPAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether GCPAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is created")
		}
		IsGCPAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether GCPAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()

		vServer = f.VaultServerWithUnsealer(1,
			kvapi.BackendStorageSpec{
				Inmem: &kvapi.InmemSpec{},
			},
			kvapi.UnsealerSpec{
				SecretShares:    4,
				SecretThreshold: 2,
				Mode: kvapi.ModeSpec{
					KubernetesSecret: &kvapi.KubernetesSecretSpec{SecretName: VaultKey},
				},
			})
		_, err := f.CreateVaultServer(vServer)
		Expect(err).NotTo(HaveOccurred(), "Create VaultServer")

		IsVaultServerCreated(vServer.Name, vServer.Namespace)
		IsAppBindingCreated(vServer.Name, vServer.Namespace)

		err = f.UpdateServiceAndAppBinding(vServer)
		Expect(err).NotTo(HaveOccurred(), "Update VaultServer's svc and appBinding")
		IsAppBindingCreated(vServer.Name, vServer.Namespace)
	})

	AfterEach(func() {
		err := f.DeleteVaultServerObj(vServer)
		Expect(err).NotTo(HaveOccurred(), "Delete VaultServer")

		err = f.DeleteAppBinding(vServer.Name, vServer.Namespace)
		Expect(err).NotTo(HaveOccurred(), "Delete VaultServer appbinding")

		IsVaultServerDeleted(vServer.Name, vServer.Namespace)
		IsAppBindingDeleted(vServer.Name, vServer.Namespace)
		time.Sleep(20 * time.Second)
	})

	FDescribe("GCPRole", func() {
		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
			gcpSE          api.SecretEngine
		)

		const (
			gcpCredSecret   = "gcp-cred-3224"
			gcpRoleName     = "my-gcp-roleset-4325"
			gcpSecretEngine = "my-gcp-secretengine-3423423"
		)

		BeforeEach(func() {

			credentialAddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			jsonBytes, err := ioutil.ReadFile(credentialAddr)
			Expect(err).NotTo(HaveOccurred(), "Parse gcp credentials")

			gcpCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpCredSecret,
					Namespace: f.Namespace(),
				},
				Data: map[string][]byte{
					"sa.json": jsonBytes,
				},
			}
			_, err = f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&gcpCredentials)
			Expect(err).NotTo(HaveOccurred(), "Create gcp credentials secret")

			gcpRole = api.GCPRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: vServer.Name,
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: ` resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
					roles = ["roles/viewer"]
				}`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			}

			gcpSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: vServer.Name,
					},
					Path: "gcp",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{
							CredentialSecret: gcpCredSecret,
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(gcpCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")
		})

		Context("Create GCPRole", func() {
			var p api.GCPRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = gcpRole
				se = gcpSE
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsVaultGCPRoleDeleted(p.RoleName())
				IsGCPRoleDeleted(p.Name, p.Namespace)

				err = f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Delete(se.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete Secret engine")

				IsSecretEngineDeleted(se.Name, se.Namespace)
			})

			It("Should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Create(&se)
				Expect(err).NotTo(HaveOccurred(), "Create SecretEngine")

				IsSecretEngineCreated(se.Name, se.Namespace)
				IsSecretEngineSucceeded(se.Name, se.Namespace)

				_, err = f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleCreated(p.Name, p.Namespace)
				IsVaultGCPRoleCreated(p.RoleName())
				IsGCPRoleSucceeded(p.Name, p.Namespace)
			})

		})

		//Context("Create GCPRole with invalid vault AppReference", func() {
		//	var p api.GCPRole
		//
		//	BeforeEach(func() {
		//		p = gcpRole
		//		p.Spec.VaultRef = core.LocalObjectReference{
		//			Name: "invalid",
		//		}
		//	})
		//
		//	AfterEach(func() {
		//		err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
		//		Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")
		//
		//		IsVaultGCPRoleDeleted(p.RoleName())
		//		IsGCPRoleDeleted(p.Name, p.Namespace)
		//	})
		//
		//	It("Should be successful", func() {
		//		_, err := f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(&p)
		//		Expect(err).NotTo(HaveOccurred(), "Create GCPRole")
		//
		//		IsGCPRoleCreated(p.Name, p.Namespace)
		//		IsVaultGCPRoleDeleted(p.RoleName())
		//	})
		//})

	})

	Describe("GCPAccessKeyRequest", func() {
		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
			gcpAKReq       api.GCPAccessKeyRequest
		)
		const (
			gcpCredSecret = "gcp-cred-2343"
			gcpRoleName   = "gcp-token-roleset-23432"
			gcpAKReqName  = "gcp-akr-324432"
		)

		BeforeEach(func() {
			credentialAddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			jsonBytes, err := ioutil.ReadFile(credentialAddr)
			Expect(err).NotTo(HaveOccurred(), "Parse gcp credentials")

			gcpCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpCredSecret,
					Namespace: f.Namespace(),
				},
				Data: map[string][]byte{
					"sa.json": jsonBytes,
				},
			}
			_, err = f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&gcpCredentials)
			Expect(err).NotTo(HaveOccurred(), "Create gcp credentials secret")

			gcpRole = api.GCPRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: ` resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
					roles = ["roles/viewer"]
				}`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			}

			gcpAKReq = api.GCPAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpAKReqName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPAccessKeyRequestSpec{
					RoleRef: api.RoleRef{
						Name:      gcpRoleName,
						Namespace: f.Namespace(),
					},
					Subjects: []rbac.Subject{
						{
							Kind:      rbac.ServiceAccountKind,
							Name:      "sa-5576",
							Namespace: f.Namespace(),
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(gcpCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")
		})

		Context("Create, Approve, Deny GCPAccessKeyRequests", func() {
			BeforeEach(func() {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create GCPAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []api.GCPAccessKeyRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")

				IsGCPAKRConditionApproved(gcpAKReq.Name, gcpAKReq.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []api.GCPAccessKeyRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsGCPAKRConditionDenied(gcpAKReq.Name, gcpAKReq.Namespace)
			})
		})

		Context("Create secret where SecretType is access_token", func() {
			var (
				secretName string
			)

			BeforeEach(func() {
				gcpRole.Spec.SecretType = api.GCPSecretAccessToken
				gcpAKReq.Status.Conditions = []api.GCPAccessKeyRequestCondition{
					{
						Type: api.AccessApproved,
					},
				}
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretDeleted(secretName, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create Access Key Secret", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				d, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(gcpAKReq.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get GCPAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

		Context("Create secret where SecretType is service_account_key", func() {
			var (
				secretName string
			)

			BeforeEach(func() {
				gcpRole.Spec.SecretType = api.GCPSecretServiceAccountKey
				gcpAKReq.Status.Conditions = []api.GCPAccessKeyRequestCondition{
					{
						Type: api.AccessApproved,
					},
				}
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretDeleted(secretName, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create Access Key Secret", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				d, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(gcpAKReq.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get GCPAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})
	})
})
