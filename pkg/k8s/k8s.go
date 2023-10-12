package k8s

import (
	"fmt"
	"strings"

	thanosAlphaV1 "github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	"github.com/ghodss/yaml"
	grafanaBetaV1 "github.com/grafana-operator/grafana-operator/api/v1beta1"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	monitoringV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
)

const (
	KindDeployment     = "Deployment"
	KindStatefulSet    = "StatefulSet"
	KindDaemonSet      = "DaemonSet"
	KindCronJob        = "CronJob"
	KindJob            = "Job"
	KindReplicaSet     = "ReplicaSet"
	KindPod            = "Pod"
	KindGrafana        = "Grafana"
	KindThanos         = "Thanos"
	KindThanosReceiver = "Receiver"
	KindConfigMap      = "ConfigMap"
	kubeKind           = "kind"
)

type (
	Deployments  appsV1.Deployment
	ConfigMap    coreV1.ConfigMap
	StatefulSets appsV1.StatefulSet
	DaemonSets   appsV1.DaemonSet
	ReplicaSets  appsV1.ReplicaSet
	CronJob      batchV1.CronJob
	Job          batchV1.Job
	Pod          coreV1.Pod
	Kind         map[string]interface{}
	containers   struct {
		containers []coreV1.Container
	}
	AlertManager   monitoringV1.Alertmanager
	Prometheus     monitoringV1.Prometheus
	ThanosRuler    monitoringV1.ThanosRuler
	Grafana        grafanaBetaV1.Grafana
	Thanos         thanosAlphaV1.Thanos
	ThanosReceiver thanosAlphaV1.Receiver
)

type KindInterface interface {
	Get(dataMap string) (string, error)
}

type ImagesInterface interface {
	Get(dataMap string) (*Image, error)
}

type Image struct {
	Kind  string   `json:"kind,omitempty"`
	Name  string   `json:"name,omitempty"`
	Image []string `json:"image,omitempty"`
}

func (kin *Kind) Get(dataMap string) (string, error) {
	var kindYaml map[string]interface{}

	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, ok := kindYaml[kubeKind].(string)
		if !ok {
			return "", &imgErrors.ImageError{Message: "failed to get name from the manifest, 'kind' is not type string"}
		}

		return value, nil
	}

	return "", nil
}

func (cm *ConfigMap) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &cm); err != nil {
		return nil, err
	}

	cmContainers := containers{}
	for k, v := range cm.Data {
		if strings.Contains(strings.ToLower(k), "image") {
			cmContainers.containers = append(cmContainers.containers, coreV1.Container{Image: v})
		}
	}

	images := &Image{
		Kind:  KindConfigMap,
		Name:  cm.Name,
		Image: cmContainers.getImages(),
	}

	return images, nil
}

func (dep *Deployments) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	for _, container := range dep.Spec.Template.Spec.Containers {
		for _, ev := range container.Env {
			println(ev.Name)
			if strings.Contains(strings.ToLower(ev.Name), "image") {
				depContainers.containers = append(depContainers.containers, coreV1.Container{Image: ev.Value})
			}

		}
	}

	images := &Image{
		Kind:  KindDeployment,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *StatefulSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindStatefulSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *DaemonSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindDaemonSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *CronJob) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.JobTemplate.Spec.Template.Spec.Containers,
		dep.Spec.JobTemplate.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindCronJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *Job) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *ReplicaSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindReplicaSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *Pod) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindPod,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *AlertManager) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	images := &Image{
		Kind:  monitoringV1.AlertmanagersKind,
		Name:  dep.Name,
		Image: []string{*dep.Spec.Image},
	}

	return images, nil
}

func (dep *Prometheus) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, *dep.Spec.Image)

	images := &Image{
		Kind:  monitoringV1.PrometheusesKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

func (dep *ThanosRuler) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, dep.Spec.Image)

	images := &Image{
		Kind:  monitoringV1.ThanosRulerKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

func (dep *Grafana) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	if dep.APIVersion == "integreatly.org/v1alpha1" {
		return nil, &imgErrors.GrafanaAPIVersionSupportError{
			Message: fmt.Sprintf("plugin supports the latest api version and '%s' is not supported", dep.APIVersion),
		}
	}

	grafanaDeployment := dep.Spec.Deployment.Spec.Template.Spec
	depContainers := containers{append(grafanaDeployment.Containers, grafanaDeployment.InitContainers...)}

	images := &Image{
		Kind:  KindGrafana,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *Thanos) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	thanosContainers := make([]coreV1.Container, 0)
	thanosContainers = append(thanosContainers, dep.Spec.Rule.StatefulsetOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.Rule.StatefulsetOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.Query.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.Query.DeploymentOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.StoreGateway.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.StoreGateway.DeploymentOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.QueryFrontend.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.QueryFrontend.DeploymentOverrides.Spec.Template.Spec.InitContainers...)

	depContainers := containers{thanosContainers}

	images := &Image{
		Kind:  KindThanos,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *ThanosReceiver) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	receiverGroupTotalContainers := make([]coreV1.Container, 0)

	for _, receiverGroup := range dep.Spec.ReceiverGroups {
		receiverGroupTotalContainers = append(receiverGroupTotalContainers, receiverGroup.StatefulSetOverrides.Spec.Template.Spec.Containers...)
		receiverGroupTotalContainers = append(receiverGroupTotalContainers,
			receiverGroup.StatefulSetOverrides.Spec.Template.Spec.InitContainers...)
	}

	depContainers := containers{receiverGroupTotalContainers}

	images := &Image{
		Kind:  KindThanosReceiver,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func NewDeployment() ImagesInterface {
	return &Deployments{}
}

func NewConfigMap() ImagesInterface {
	return &ConfigMap{}
}

func NewStatefulSet() ImagesInterface {
	return &StatefulSets{}
}

func NewDaemonSet() ImagesInterface {
	return &DaemonSets{}
}

func NewReplicaSets() ImagesInterface {
	return &ReplicaSets{}
}

func NewCronjob() ImagesInterface {
	return &CronJob{}
}

func NewJob() ImagesInterface {
	return &Job{}
}

func NewPod() ImagesInterface {
	return &Pod{}
}

func NewAlertManager() ImagesInterface {
	return &AlertManager{}
}

func NewPrometheus() ImagesInterface {
	return &Prometheus{}
}

func NewThanosRuler() ImagesInterface {
	return &ThanosRuler{}
}

func NewGrafana() ImagesInterface {
	return &Grafana{}
}

func NewThanos() ImagesInterface {
	return &Thanos{}
}

func NewThanosReceiver() ImagesInterface {
	return &ThanosReceiver{}
}

func NewKind() KindInterface {
	return &Kind{}
}

func SupportedKinds() []string {
	kinds := []string{
		KindDeployment, KindStatefulSet, KindDaemonSet,
		KindCronJob, KindJob, KindReplicaSet, KindPod,
		monitoringV1.AlertmanagersKind, monitoringV1.PrometheusesKind, monitoringV1.ThanosRulerKind,
		KindGrafana, KindThanos, KindThanosReceiver, KindConfigMap,
	}

	return kinds
}

func (cont containers) getImages() []string {
	images := make([]string, 0)
	for _, container := range cont.containers {
		images = append(images, container.Image)
	}

	return images
}
