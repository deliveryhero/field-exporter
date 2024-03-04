package resourcefieldexport

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive

	redisv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/redis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
	cr "sigs.k8s.io/controller-runtime/pkg/client"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ResourceFieldExport controller", func() {
	var (
		testNamespace string
		redisInstance *redisv1beta1.RedisInstance
	)
	BeforeEach(func() {
		ctx := context.Background()

		// generate randomized test namespace name
		testNamespace = fmt.Sprintf("test-%3d", rand.Intn(10000))
		// create the test namespace
		namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
		Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())

		// create the source redis instance
		redisInstance = &redisv1beta1.RedisInstance{
			TypeMeta: metav1.TypeMeta{
				Kind:       redisv1beta1.RedisInstanceGVK.Kind,
				APIVersion: redisv1beta1.RedisInstanceGVK.GroupVersion().String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-instance",
				Namespace: testNamespace,
			},
			Spec: redisv1beta1.RedisInstanceSpec{
				DisplayName:      ptr.To("test-0001-testdb-default"),
				MemorySizeGb:     5,
				ReadReplicasMode: ptr.To("READ_REPLICAS_ENABLED"),
				RedisVersion:     ptr.To("REDIS_6_X"),
			},
		}
		redisInstanceMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(redisInstance.DeepCopy())
		Expect(err).Should(Succeed())

		Expect(k8sClient.Create(ctx, &unstructured.Unstructured{Object: redisInstanceMap})).Should(Succeed())

		// create target configmap
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "target-cm",
				Namespace: testNamespace,
			},
		}
		Expect(k8sClient.Create(ctx, cm)).Should(Succeed())

	})

	AfterEach(func() {
		ctx := context.Background()
		namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
		Expect(k8sClient.Delete(ctx, namespace, cr.PropagationPolicy(apimetav1.DeletePropagationForeground))).Should(Succeed())
	})

	Context("for existing source resource", func() {
		When("creating a field export", func() {
			It("should succeed", func() {
				ctx := context.Background()
				rfe := &gdpv1alpha1.ResourceFieldExport{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: testNamespace,
					},
					Spec: gdpv1alpha1.ResourceFieldExportSpec{
						From: gdpv1alpha1.ResourceRef{
							APIVersion: redisv1beta1.RedisInstanceGVK.GroupVersion().String(),
							Kind:       redisv1beta1.RedisInstanceGVK.Kind,
							Name:       "redis-instance",
						},
						To: gdpv1alpha1.DestinationRef{
							Type: gdpv1alpha1.ConfigMap,
							Name: "target-cm",
						},
						Outputs: []gdpv1alpha1.Output{
							{
								Key:  "display-name",
								Path: ".spec.displayName",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())

				Eventually(func() string {
					ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
					defer cancelFunc()
					cm := &corev1.ConfigMap{}
					Expect(k8sClient.Get(ctx, cr.ObjectKey{Namespace: testNamespace, Name: "target-cm"}, cm)).Should(Succeed())
					return cm.Data["display-name"]
				}, "10s").Should(Equal("test-0001-testdb-default"))
			})
		})

		When("source resource is updated", func() {
			BeforeEach(func() {
				ctx := context.Background()
				rfe := &gdpv1alpha1.ResourceFieldExport{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: testNamespace,
					},
					Spec: gdpv1alpha1.ResourceFieldExportSpec{
						From: gdpv1alpha1.ResourceRef{
							APIVersion: redisv1beta1.RedisInstanceGVK.GroupVersion().String(),
							Kind:       redisv1beta1.RedisInstanceGVK.Kind,
							Name:       "redis-instance",
						},
						To: gdpv1alpha1.DestinationRef{
							Type: gdpv1alpha1.ConfigMap,
							Name: "target-cm",
						},
						Outputs: []gdpv1alpha1.Output{
							{
								Key:  "display-name",
								Path: ".spec.displayName",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())
			})

			It("target should be updated", func() {
				ctx := context.Background()
				riMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(redisInstance)
				Expect(err).Should(BeNil())
				riUnstructured := &unstructured.Unstructured{Object: riMap}
				riUnstructured.SetGroupVersionKind(redisv1beta1.RedisInstanceGVK)

				data := `{"spec":{"displayName":"new-display-name"}}`
				Expect(k8sClient.Patch(ctx, riUnstructured, cr.RawPatch(types.MergePatchType, []byte(data)))).Should(Succeed())
				Eventually(func() string {
					ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
					defer cancelFunc()
					cm := &corev1.ConfigMap{}
					Expect(k8sClient.Get(ctx, cr.ObjectKey{Namespace: testNamespace, Name: "target-cm"}, cm)).Should(Succeed())
					return cm.Data["display-name"]
				}, "10s").Should(Equal("new-display-name"))

			})
		})
	})

})
