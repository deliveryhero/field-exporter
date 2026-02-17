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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
)

var _ = Describe("ResourceFieldExport controller", func() {
	var (
		testNamespace string
		redisInstance *redisv1beta1.RedisInstance
	)
	BeforeEach(func() {
		ctx := context.Background()

		// generate randomized test namespace name
		testNamespace = fmt.Sprintf("test-%03d", rand.Intn(10000))
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

	Context("for existing source resource (GCP)", func() {
		When("creating a field export", func() {
			It("should succeed", func() {
				from := gdpv1alpha1.ResourceRef{
					APIVersion: redisv1beta1.RedisInstanceGVK.GroupVersion().String(),
					Kind:       redisv1beta1.RedisInstanceGVK.Kind,
					Name:       "redis-instance",
				}
				outputs := []gdpv1alpha1.Output{
					{
						Key:  "display-name",
						Path: ".spec.displayName",
					},
				}
				expectedData := map[string]string{
					"display-name": "test-0001-testdb-default",
				}
				testSuccessfulFieldExport(testNamespace, "test", from, outputs, expectedData)
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

	Context("for existing source resource (AWS DBCluster)", func() {
		var awsDbCluster *unstructured.Unstructured

		BeforeEach(func() {
			ctx := context.Background()
			// create the source aws rds cluster
			awsDbCluster = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "rds.services.k8s.aws/v1alpha1",
					"kind":       "DBCluster",
					"metadata": map[string]interface{}{
						"name":      "aws-db-cluster",
						"namespace": testNamespace,
					},
					"spec": map[string]interface{}{
						"dbClusterIdentifier": "aws-db-cluster",
						"engine":              "aurora-postgresql",
						"masterUsername":      "testuser",
					},
				},
			}
			Expect(k8sClient.Create(ctx, awsDbCluster)).Should(Succeed())

			// Update status with cluster endpoints
			awsDbCluster.Object["status"] = map[string]interface{}{
				"endpoint":       "my-cluster-writer.cluster-random.us-east-1.rds.amazonaws.com",
				"readerEndpoint": "my-cluster-reader.cluster-random.us-east-1.rds.amazonaws.com",
			}
			Expect(k8sClient.Status().Update(ctx, awsDbCluster)).Should(Succeed())
		})

		When("creating a field export for an AWS DBCluster resource", func() {
			It("should succeed and populate the target with cluster endpoints", func() {
				from := gdpv1alpha1.ResourceRef{
					APIVersion: "rds.services.k8s.aws/v1alpha1",
					Kind:       "DBCluster",
					Name:       "aws-db-cluster",
				}
				outputs := []gdpv1alpha1.Output{
					{
						Key:  "cluster-writer-endpoint",
						Path: ".status.endpoint",
					},
					{
						Key:  "cluster-reader-endpoint",
						Path: ".status.readerEndpoint",
					},
				}
				expectedData := map[string]string{
					"cluster-writer-endpoint": "my-cluster-writer.cluster-random.us-east-1.rds.amazonaws.com",
					"cluster-reader-endpoint": "my-cluster-reader.cluster-random.us-east-1.rds.amazonaws.com",
				}
				testSuccessfulFieldExport(testNamespace, "test-aws-cluster", from, outputs, expectedData)
			})
		})
	})

	Context("for existing source resource (AWS)", func() {
		var awsDbInstance *unstructured.Unstructured

		BeforeEach(func() {
			ctx := context.Background()
			// create the source aws rds instance
			awsDbInstance = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "rds.services.k8s.aws/v1alpha1",
					"kind":       "DBInstance",
					"metadata": map[string]interface{}{
						"name":      "aws-db-instance",
						"namespace": testNamespace,
					},
					"spec": map[string]interface{}{
						"dbInstanceIdentifier": "aws-db-instance",
						"dbInstanceClass":      "db.t3.micro",
						"engine":               "postgres",
					},
				},
			}
			// The status field is ignored on Create when a status subresource is present.
			// We create the object first...
			Expect(k8sClient.Create(ctx, awsDbInstance)).Should(Succeed())

			// ...then we add the status field to the object...
			awsDbInstance.Object["status"] = map[string]interface{}{
				"endpoint": map[string]interface{}{
					"address": "my-aws-db.random-chars.us-east-1.rds.amazonaws.com",
				},
			}
			// ...and update the status subresource.
			Expect(k8sClient.Status().Update(ctx, awsDbInstance)).Should(Succeed())
		})

		When("creating a field export for an RDS resource", func() {
			It("should succeed and populate the target", func() {
				from := gdpv1alpha1.ResourceRef{
					APIVersion: "rds.services.k8s.aws/v1alpha1",
					Kind:       "DBInstance",
					Name:       "aws-db-instance",
				}
				outputs := []gdpv1alpha1.Output{
					{
						Key:  "db-endpoint",
						Path: ".status.endpoint.address",
					},
				}
				expectedData := map[string]string{
					"db-endpoint": "my-aws-db.random-chars.us-east-1.rds.amazonaws.com",
				}
				testSuccessfulFieldExport(testNamespace, "test-aws", from, outputs, expectedData)
			})
		})

		When("creating a field export for an RDS resource with required conditions", func() {
			It("should succeed when conditions are met", func() {
				ctx := context.Background()

				// Update status to include conditions
				awsDbInstance.Object["status"].(map[string]interface{})["conditions"] = []interface{}{
					map[string]interface{}{"type": "Ready", "status": "True"},
				}
				Expect(k8sClient.Status().Update(ctx, awsDbInstance)).Should(Succeed())

				rfe := &gdpv1alpha1.ResourceFieldExport{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-aws-conditions-ok",
						Namespace: testNamespace,
					},
					Spec: gdpv1alpha1.ResourceFieldExportSpec{
						From: gdpv1alpha1.ResourceRef{
							APIVersion: "rds.services.k8s.aws/v1alpha1",
							Kind:       "DBInstance",
							Name:       "aws-db-instance",
						},
						To: gdpv1alpha1.DestinationRef{
							Type: gdpv1alpha1.ConfigMap,
							Name: "target-cm",
						},
						RequiredFields: &gdpv1alpha1.RequiredFields{
							StatusConditions: []gdpv1alpha1.StatusCondition{
								{Type: "Ready", Status: "True"},
							},
						},
						Outputs: []gdpv1alpha1.Output{
							{
								Key:  "db-endpoint",
								Path: ".status.endpoint.address",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())

				Eventually(func() string {
					cm := &corev1.ConfigMap{}
					_ = k8sClient.Get(context.Background(), cr.ObjectKey{Namespace: testNamespace, Name: "target-cm"}, cm)
					return cm.Data["db-endpoint"]
				}, "10s").Should(Equal("my-aws-db.random-chars.us-east-1.rds.amazonaws.com"))
			})

			It("should fail when conditions are not met", func() {
				ctx := context.Background()
				rfe := &gdpv1alpha1.ResourceFieldExport{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-aws-conditions-fail",
						Namespace: testNamespace,
					},
					Spec: gdpv1alpha1.ResourceFieldExportSpec{
						From: gdpv1alpha1.ResourceRef{
							APIVersion: "rds.services.k8s.aws/v1alpha1",
							Kind:       "DBInstance",
							Name:       "aws-db-instance",
						},
						To: gdpv1alpha1.DestinationRef{
							Type: gdpv1alpha1.ConfigMap,
							Name: "target-cm",
						},
						RequiredFields: &gdpv1alpha1.RequiredFields{
							StatusConditions: []gdpv1alpha1.StatusCondition{
								// This condition does not exist on the object
								{Type: "Available", Status: "True"},
							},
						},
						Outputs: []gdpv1alpha1.Output{
							// Add a dummy output to satisfy schema validation
							{Key: "dummy", Path: ".status.dummy"},
						},
					},
				}
				Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())

				// Check that the RFE status becomes degraded
				Eventually(func() corev1.ConditionStatus {
					updatedRfe := &gdpv1alpha1.ResourceFieldExport{}
					_ = k8sClient.Get(context.Background(), cr.ObjectKeyFromObject(rfe), updatedRfe)
					if len(updatedRfe.Status.Conditions) > 0 {
						return updatedRfe.Status.Conditions[0].Status
					}
					return corev1.ConditionUnknown
				}, "10s").Should(Equal(corev1.ConditionFalse))
			})
		})
	})

	Context("for existing source resource (AWS DynamoDB)", func() {
		var awsDynamoDBTable *unstructured.Unstructured

		BeforeEach(func() {
			ctx := context.Background()
			// create the source aws dynamodb table
			awsDynamoDBTable = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "dynamodb.services.k8s.aws/v1alpha1",
					"kind":       "Table",
					"metadata": map[string]interface{}{
						"name":      "aws-dynamodb-table",
						"namespace": testNamespace,
					},
					"spec": map[string]interface{}{
						"tableName": "aws-dynamodb-table",
						"keySchema": []interface{}{
							map[string]interface{}{"attributeName": "id", "keyType": "HASH"},
						},
						"attributeDefinitions": []interface{}{
							map[string]interface{}{"attributeName": "id", "attributeType": "S"},
						},
						"provisionedThroughput": map[string]interface{}{
							"readCapacityUnits":  int64(5),
							"writeCapacityUnits": int64(5),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, awsDynamoDBTable)).Should(Succeed())

			// Update status
			awsDynamoDBTable.Object["status"] = map[string]interface{}{
				"ackResourceMetadata": map[string]interface{}{
					"ownerAccountID": "123456789012",
					"region":         "us-west-2",
					"arn":            "arn:aws:dynamodb:us-west-2:123456789012:table/aws-dynamodb-table",
				},
			}
			Expect(k8sClient.Status().Update(ctx, awsDynamoDBTable)).Should(Succeed())
		})

		When("creating a field export for an AWS DynamoDB resource", func() {
			It("should succeed and populate the target", func() {
				from := gdpv1alpha1.ResourceRef{
					APIVersion: "dynamodb.services.k8s.aws/v1alpha1",
					Kind:       "Table",
					Name:       "aws-dynamodb-table",
				}
				outputs := []gdpv1alpha1.Output{
					{
						Key:  "table-arn",
						Path: ".status.ackResourceMetadata.arn",
					},
				}
				expectedData := map[string]string{
					"table-arn": "arn:aws:dynamodb:us-west-2:123456789012:table/aws-dynamodb-table",
				}
				testSuccessfulFieldExport(testNamespace, "test-aws-dynamodb", from, outputs, expectedData)
			})
		})
	})

	Context("for existing source resource (AWS ElastiCache)", func() {
		var awsElastiCacheRG *unstructured.Unstructured

		BeforeEach(func() {
			ctx := context.Background()
			// create the source aws elasticache replication group
			awsElastiCacheRG = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "elasticache.services.k8s.aws/v1alpha1",
					"kind":       "ReplicationGroup",
					"metadata": map[string]interface{}{
						"name":      "aws-elasticache-rg",
						"namespace": testNamespace,
					},
					"spec": map[string]interface{}{
						"replicationGroupID": "aws-elasticache-rg",
						"description":        "test replication group",
						"cacheNodeType":      "cache.t3.micro",
						"engine":             "redis",
					},
				},
			}
			// The status field is ignored on Create when a status subresource is present.
			// We create the object first...
			Expect(k8sClient.Create(ctx, awsElastiCacheRG)).Should(Succeed())

			// ...then we add the status field to the object...
			awsElastiCacheRG.Object["status"] = map[string]interface{}{
				"ackResourceMetadata": map[string]interface{}{
					"ownerAccountID": "123456789012",
					"region":         "us-east-1",
				},
				"nodeGroups": []interface{}{
					map[string]interface{}{
						"primaryEndpoint": map[string]interface{}{
							"address": "my-elasticache-rg.random-chars.us-east-1.cache.amazonaws.com",
						},
					},
				},
			}
			// ...and update the status subresource.
			Expect(k8sClient.Status().Update(ctx, awsElastiCacheRG)).Should(Succeed())
		})

		When("creating a field export for an AWS ElastiCache resource", func() {
			It("should succeed and populate the target", func() {
				from := gdpv1alpha1.ResourceRef{
					APIVersion: "elasticache.services.k8s.aws/v1alpha1",
					Kind:       "ReplicationGroup",
					Name:       "aws-elasticache-rg",
				}
				outputs := []gdpv1alpha1.Output{
					{
						Key:  "redis-endpoint",
						Path: ".status.nodeGroups[0].primaryEndpoint.address",
					},
				}
				expectedData := map[string]string{
					"redis-endpoint": "my-elasticache-rg.random-chars.us-east-1.cache.amazonaws.com",
				}
				testSuccessfulFieldExport(testNamespace, "test-aws-elasticache", from, outputs, expectedData)
			})
		})
	})
})

func testSuccessfulFieldExport(testNamespace, rfeName string, from gdpv1alpha1.ResourceRef, outputs []gdpv1alpha1.Output, expectedData map[string]string) {
	ctx := context.Background()
	rfe := &gdpv1alpha1.ResourceFieldExport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rfeName,
			Namespace: testNamespace,
		},
		Spec: gdpv1alpha1.ResourceFieldExportSpec{
			From: from,
			To: gdpv1alpha1.DestinationRef{
				Type: gdpv1alpha1.ConfigMap,
				Name: "target-cm",
			},
			Outputs: outputs,
		},
	}
	Expect(k8sClient.Create(ctx, rfe)).Should(Succeed())

	for key, value := range expectedData {
		Eventually(func() string {
			cm := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), cr.ObjectKey{Namespace: testNamespace, Name: "target-cm"}, cm)).Should(Succeed())
			return cm.Data[key]
		}, "10s").Should(Equal(value))
	}
}
