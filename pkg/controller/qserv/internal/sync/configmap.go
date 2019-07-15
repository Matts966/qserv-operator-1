package sync

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	qservv1alpha1 "github.com/lsst/qserv-operator/pkg/apis/qserv/v1alpha1"
	"github.com/lsst/qserv-operator/pkg/scheme/qserv"
	"github.com/lsst/qserv-operator/pkg/staging/syncer"
)

// NewXrootdEtcConfigMapSyncer returns a new sync.Interface for reconciling XrootdEtc ConfigMap
func NewXrootdEtcConfigMapSyncer(r *qservv1alpha1.Qserv, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	cm := qserv.GenerateConfigMap(r, controllerLabels, "xrootd", "etc")
	return syncer.NewObjectSyncer("XrootdEtcConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}

// NewXrootdStartConfigMapSyncer returns a new sync.Interface for reconciling XrootdStart ConfigMap
func NewXrootStartConfigMapSyncer(r *qservv1alpha1.Qserv, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	cm := qserv.GenerateConfigMap(r, controllerLabels, "xrootd", "start")
	return syncer.NewObjectSyncer("XrootdStartConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}

// // NewRedisShutdownConfigMapSyncer returns a new sync.Interface for reconciling Redis Shutdown ConfigMap
// func NewRedisShutdownConfigMapSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
// 	cm := redis.GenerateRedisShutdownConfigMap(r, controllerLabels)
// 	return syncer.NewObjectSyncer("RedisShutdownConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
// 		return nil
// 	})
// }

// // NewSentinelConfigMapSyncer returns a new sync.Interface for reconciling Sentinel ConfigMap
// func NewSentinelConfigMapSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
// 	cm := redis.GenerateSentinelConfigMap(r, controllerLabels)
// 	return syncer.NewObjectSyncer("SentinelConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
// 		return nil
// 	})
// }
