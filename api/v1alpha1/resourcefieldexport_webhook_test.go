package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ResourceFieldExport Webhook", func() {
	_ = Context("on create", func() {
		var rfe *ResourceFieldExport
		BeforeEach(func() {
			rfe = &ResourceFieldExport{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-source",
					Namespace: "default",
				},
				Spec: ResourceFieldExportSpec{
					From: ResourceRef{
						APIVersion: "redis.cnrm.cloud.google.com/v1beta1",
						Kind:       "RedisInstance",
						Name:       "sensitive-secret",
					},
					To: DestinationRef{
						Type: ConfigMap,
						Name: "compromised",
					},
					RequiredFields: &RequiredFields{StatusConditions: []StatusCondition{
						{
							Type:   "Ready",
							Status: "True",
						},
					}},
					Outputs: []Output{
						{
							Key:  "ip",
							Path: ".status.ip",
						},
					},
				},
			}
		})

		_ = When("source resource is valid", func() {
			It("succeeds", func() {
				Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())
			})
		})

		_ = When("source resource is invalid", func() {
			It("fails", func() {
				rfe.Spec.From = ResourceRef{
					APIVersion: "v1",
					Kind:       "Secret",
					Name:       "super-sensitive",
				}
				Expect(k8sClient.Create(ctx, rfe)).Should(MatchError(ContainSubstring("unsupported GroupVersion v1")))
			})
		})

		_ = When("source resource is unknown", func() {
			It("fails", func() {
				rfe.Spec.From.Kind = "RedisCluster"
				Expect(k8sClient.Create(ctx, rfe)).Should(MatchError(ContainSubstring("unsupported resource: redis.cnrm.cloud.google.com/v1beta1, Kind=RedisCluster")))
			})
		})

		_ = When("output path is invalid", func() {
			It("fails", func() {
				rfe.Spec.Outputs[0].Key = "**&&&&"
				Expect(k8sClient.Create(ctx, rfe)).Should(MatchError(ContainSubstring(`output key **&&&& is invalid: unexpected token "*"`)))
			})
		})
	})
})
