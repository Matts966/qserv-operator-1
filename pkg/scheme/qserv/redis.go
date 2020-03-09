package qserv

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	qservv1alpha1 "github.com/lsst/qserv-operator/pkg/apis/qserv/v1alpha1"
	"github.com/lsst/qserv-operator/pkg/constants"
	"github.com/lsst/qserv-operator/pkg/util"
	kubedbv1alpha "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

// GenerateRedis generate custom resource specification for Redis database
func GenerateRedis(cr *qservv1alpha1.Qserv, labels map[string]string) *kubedbv1alpha.Redis {
	name := cr.Name + "-" + string(constants.RedisName)
	namespace := cr.Namespace
	labels = util.MergeLabels(labels, util.GetLabels(constants.CzarName, cr.Name))

	var masters int32 = 1
	var replicas int32 = 1
	storageClass := cr.Spec.StorageClass
	storageSize := cr.Spec.StorageCapacity

	rcr := &kubedbv1alpha.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: kubedbv1alpha.RedisSpec{
			Version: "4.0-v2",
			Mode:    kubedbv1alpha.RedisModeCluster,
			Cluster: &kubedbv1alpha.RedisClusterSpec{
				Master:   &masters,
				Replicas: &replicas,
			},
			StorageType: kubedbv1alpha.StorageTypeDurable,
			Storage: &v1.PersistentVolumeClaimSpec{
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				StorageClassName: &storageClass,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						"storage": resource.MustParse(storageSize),
					},
				},
			},
			TerminationPolicy: kubedbv1alpha.TerminationPolicyDelete,
			// UpdateStrategy: v1.StatefulSetUpdateStrategy
		},
	}

	//ss.Spec.Template.Spec.Tolerations = cr.Spec.Tolerations

	return rcr
}
