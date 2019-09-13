package qserv

import (
	"fmt"
	"path/filepath"
	"strconv"

	qservv1alpha1 "github.com/lsst/qserv-operator/pkg/apis/qserv/v1alpha1"
	"github.com/lsst/qserv-operator/pkg/constants"
	"github.com/lsst/qserv-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getInitContainer(cr *qservv1alpha1.Qserv, component constants.ComponentName) (v1.Container, VolumeSet) {
	sqlConfigMap := fmt.Sprintf("config-sql-%s", component)

	componentName := string(component)

	container := v1.Container{
		Name:  string(constants.InitDbName),
		Image: getMariadbImage(cr, component),
		Command: []string{
			"/config-start/init.sh",
		},
		Env: []v1.EnvVar{
			{
				Name:  "COMPONENT_NAME",
				Value: componentName,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			getDataVolumeMount(),
			// db startup script, configuration and root passwords are shared
			getEtcVolumeMount(constants.MariadbName),
			getStartVolumeMount(constants.MariadbName),
			getSecretVolumeMount(constants.MariadbName),
			{
				MountPath: filepath.Join("/", "config-sql", componentName),
				Name:      sqlConfigMap,
				ReadOnly:  false,
			},
		},
	}

	var volumes VolumeSet
	volumes.make(nil)

	volumes.addConfigMapVolume(sqlConfigMap)
	volumes.addEtcStartVolumes(constants.MariadbName)
	volumes.addSecretVolume(constants.MariadbName)

	if component == constants.ReplName {
		container.Env = append(container.Env, getXrootdRedirectorDn(cr))
		container.VolumeMounts = append(container.VolumeMounts, getSecretVolumeMount(constants.ReplDbName))
		volumes.addSecretVolume(constants.ReplDbName)
	}

	return container, volumes
}

func getSecretVolumeMount(containerName constants.ContainerName) v1.VolumeMount {
	secretName := GetSecretName(containerName)
	return v1.VolumeMount{
		MountPath: filepath.Join("/", secretName),
		Name:      secretName,
		ReadOnly:  false}
}

func getMariadbContainer(cr *qservv1alpha1.Qserv, component constants.ComponentName) (v1.Container, VolumeSet) {

	var uservice constants.ContainerName
	if component == constants.ReplName {
		uservice = constants.ReplDbName
	} else {
		uservice = constants.MariadbName
	}

	mariadbPortName := string(constants.MariadbName)

	container := v1.Container{
		Name:  string(constants.MariadbName),
		Image: getMariadbImage(cr, component),
		Ports: []v1.ContainerPort{
			{
				Name:          mariadbPortName,
				ContainerPort: constants.MariadbPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		Command:        constants.Command,
		LivenessProbe:  getLivenessProbe(mariadbPortName),
		ReadinessProbe: getReadinessProbe(mariadbPortName),
		VolumeMounts: []v1.VolumeMount{
			getDataVolumeMount(),
			getEtcVolumeMount(uservice),
			getStartVolumeMount(uservice),
			{
				MountPath: "/qserv/run/tmp",
				Name:      "tmp-volume",
				ReadOnly:  false,
			},
		},
	}

	// Volumes
	var volumes VolumeSet
	volumes.make(nil)

	volumes.addEmptyDirVolume("tmp-volume")
	volumes.addEtcStartVolumes(uservice)

	return container, volumes
}

func getMariadbImage(cr *qservv1alpha1.Qserv, component constants.ComponentName) string {
	spec := cr.Spec
	var image string
	if component == constants.ReplName {
		image = spec.Replication.DbImage
	} else {
		image = spec.Worker.Image
	}
	return image
}

func getProxyContainer(cr *qservv1alpha1.Qserv) (v1.Container, VolumeSet) {
	spec := cr.Spec
	container := v1.Container{
		Name:  string(constants.ProxyName),
		Image: spec.Worker.Image,
		Ports: []v1.ContainerPort{
			{
				Name:          string(constants.ProxyName),
				ContainerPort: constants.ProxyPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		LivenessProbe:  getLivenessProbe(constants.ProxyPortName),
		ReadinessProbe: getReadinessProbe(constants.ProxyPortName),
		Command:        constants.Command,
		Env: []v1.EnvVar{
			{
				Name:  "XROOTD_RDR_DN",
				Value: util.GetName(cr, string(constants.XrootdRedirectorName)),
			},
		},
		VolumeMounts: []v1.VolumeMount{
			// Used for mysql socket access
			// TODO move mysql socket in emptyDir?
			getDataVolumeMount(),
			getEtcVolumeMount(constants.ProxyName),
			getStartVolumeMount(constants.ProxyName),
		},
	}

	// Volumes
	var volumes VolumeSet
	volumes.make(nil)

	volumes.addEtcStartVolumes(constants.ProxyName)

	return container, volumes
}

func getReplicationCtlContainer(cr *qservv1alpha1.Qserv) (v1.Container, VolumeSet) {
	spec := cr.Spec

	container := v1.Container{
		Name:    string(constants.ReplCtlName),
		Image:   spec.Replication.Image,
		Command: constants.Command,
		Env: []v1.EnvVar{
			{
				Name:  "WORKER_COUNT",
				Value: strconv.FormatInt(int64(spec.Worker.Replicas), 10),
			},
			{
				Name:  "REPL_DB_DN",
				Value: util.GetName(cr, string(constants.ReplDbName)),
			},
			getXrootdRedirectorDn(cr),
		},
		VolumeMounts: []v1.VolumeMount{
			getEtcVolumeMount(constants.ReplCtlName),
			getStartVolumeMount(constants.ReplCtlName),
		},
	}

	var volumes VolumeSet
	volumes.make(nil)

	volumes.addEtcStartVolumes(constants.ReplCtlName)

	return container, volumes
}

func getReplicationWrkContainer(cr *qservv1alpha1.Qserv) (v1.Container, VolumeSet) {
	spec := cr.Spec

	var runAsUser int64 = 1000

	container := v1.Container{
		Name:    string(constants.ReplWrkName),
		Image:   spec.Replication.Image,
		Command: constants.Command,
		Env: []v1.EnvVar{
			{
				Name:  "CZAR_DN",
				Value: util.GetName(cr, string(constants.CzarName)),
			},
			{
				Name:  "REPL_DB_DN",
				Value: util.GetName(cr, string(constants.ReplDbName)),
			},
		},
		SecurityContext: &v1.SecurityContext{
			RunAsUser: &runAsUser,
		},
		VolumeMounts: []v1.VolumeMount{
			getDataVolumeMount(),
			getEtcVolumeMount(constants.ReplWrkName),
			getStartVolumeMount(constants.ReplWrkName),
		},
	}

	var volumes VolumeSet
	volumes.make(nil)

	volumes.addEtcStartVolumes(constants.ReplWrkName)

	return container, volumes
}

func getWmgrContainer(cr *qservv1alpha1.Qserv) (v1.Container, VolumeSet) {
	spec := cr.Spec
	container := v1.Container{
		Name:  string(constants.WmgrName),
		Image: spec.Worker.Image,
		Ports: []v1.ContainerPort{
			{
				Name:          constants.WmgrPortName,
				ContainerPort: constants.WmgrPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		Command: constants.Command,
		Env: []v1.EnvVar{
			{
				Name:  "CZAR_DN",
				Value: util.GetName(cr, string(constants.CzarName)),
			},
		},
		LivenessProbe:  getLivenessProbe(constants.WmgrPortName),
		ReadinessProbe: getReadinessProbe(constants.WmgrPortName),
		VolumeMounts: []v1.VolumeMount{
			{
				MountPath: filepath.Join("/", "config-dot-qserv"),
				Name:      "config-dot-qserv",
				ReadOnly:  true,
			},
			{
				MountPath: "/qserv/run/tmp",
				Name:      "tmp-volume",
				ReadOnly:  false,
			},
			getSecretVolumeMount(constants.MariadbName),
			getSecretVolumeMount(constants.WmgrName),
			getDataVolumeMount(),
			getEtcVolumeMount(constants.WmgrName),
			getStartVolumeMount(constants.WmgrName),
		},
	}

	// Volumes
	var volumes VolumeSet
	volumes.make(nil)

	volumes.addConfigMapVolume("config-dot-qserv")
	volumes.addSecretVolume(constants.MariadbName)
	volumes.addSecretVolume(constants.WmgrName)
	volumes.addEmptyDirVolume("tmp-volume")
	volumes.addEtcStartVolumes("wmgr")

	// TODO Add volumes
	return container, volumes
}

func getXrootdRedirectorDn(cr *qservv1alpha1.Qserv) v1.EnvVar {
	return v1.EnvVar{
		Name:  "XROOTD_RDR_DN",
		Value: util.GetName(cr, string(constants.XrootdRedirectorName)),
	}
}

func getXrootdContainers(cr *qservv1alpha1.Qserv, component constants.ComponentName) ([]v1.Container, VolumeSet) {

	const (
		CMSD = iota
		XROOTD
	)

	spec := cr.Spec

	volumeMounts := getXrootdVolumeMounts(component)
	xrootdPortName := string(constants.XrootdName)

	containers := []v1.Container{
		{
			Name:    string(constants.CmsdName),
			Image:   spec.Worker.Image,
			Command: constants.Command,
			Args:    []string{"-S", "cmsd"},
			Env: []v1.EnvVar{
				getXrootdRedirectorDn(cr),
			},
			SecurityContext: &v1.SecurityContext{
				Capabilities: &v1.Capabilities{
					Add: []v1.Capability{
						v1.Capability("IPC_LOCK"),
					},
				},
			},
			VolumeMounts: volumeMounts,
		},
		{
			Name:  string(constants.XrootdName),
			Image: spec.Worker.Image,
			Ports: []v1.ContainerPort{
				{
					Name:          xrootdPortName,
					ContainerPort: 1094,
					Protocol:      v1.ProtocolTCP,
				},
			},
			Command: constants.Command,
			Env: []v1.EnvVar{
				getXrootdRedirectorDn(cr),
			},
			LivenessProbe:  getLivenessProbe(xrootdPortName),
			ReadinessProbe: getReadinessProbe(xrootdPortName),
			SecurityContext: &v1.SecurityContext{
				Capabilities: &v1.Capabilities{
					Add: []v1.Capability{
						v1.Capability("IPC_LOCK"),
						v1.Capability("SYS_RESOURCE"),
					},
				},
			},
			VolumeMounts: volumeMounts,
		},
	}

	// Cmsd port is only open on redirectors, not on workers
	if component == constants.XrootdRedirectorName {
		containers[0].Ports = []v1.ContainerPort{
			{
				Name:          string(constants.CmsdName),
				ContainerPort: 2131,
				Protocol:      v1.ProtocolTCP,
			},
		}
		containers[0].LivenessProbe = getLivenessProbe(constants.CmsdPortName)
		containers[0].ReadinessProbe = getReadinessProbe(constants.CmsdPortName)
	}

	// Volumes
	var volumes VolumeSet
	volumes.make(nil)

	volumes.addEtcStartVolumes(constants.XrootdName)
	volumes.addEmptyDirVolume(constants.XrootdAdminPathVolumeName)

	return containers, volumes
}

func getLivenessProbe(portName string) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString(portName),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
	}
}

func getReadinessProbe(portName string) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString(portName),
			},
		},
		InitialDelaySeconds: 10,
		PeriodSeconds:       5,
	}
}
