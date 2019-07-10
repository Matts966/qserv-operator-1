package sync

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/scheme/redis"
	"github.com/kube-incubator/redis-operator/pkg/staging/syncer"
)

// NewRedisServiceSyncer returns a new sync.Interface for reconciling Redis Service
func NewRedisServiceSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	svc := redis.GenerateRedisService(r, controllerLabels)
	return syncer.NewObjectSyncer("RedisService", r, svc, c, scheme, noFunc)
}

// NewSentinelServiceSyncer returns a new sync.Interface for reconciling Sentinel Service
func NewSentinelServiceSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	svc := redis.GenerateSentinelService(r, controllerLabels)
	return syncer.NewObjectSyncer("SentinelService", r, svc, c, scheme, noFunc)
}
