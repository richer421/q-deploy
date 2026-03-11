package model

// InstanceConfigSnapshot 是从 q-metahub InstanceConfig 派生的运行态快照
// 这里只关心渲染需要的字段

type InstanceConfigSnapshot struct {
	Env             string         // dev/test/gray/prod
	InstanceType    string         // deployment/statefulset/job/cronjob/pod
	Spec            map[string]any // K8s 原生 Spec（DeploymentSpec 等）
	AttachResources map[string]any // 附加资源（ConfigMap/Service/Ingress 等）
}

// ReleaseStrategySnapshot 是从 q-metahub CDConfig.ReleaseStrategy 派生的快照

type ReleaseStrategySnapshot struct {
	DeploymentMode    string
	BatchRule         BatchRuleSnapshot
	CanaryTrafficRule *CanaryTrafficRuleSnapshot
}

type BatchRuleSnapshot struct {
	BatchCount  int
	BatchRatio  []float64
	TriggerType string
	Interval    int
}

type CanaryTrafficRuleSnapshot struct {
	TrafficBatchCount int
	TrafficRatioList  []float64
	ManualAdjust      bool
	AdjustTimeout     int
}

// Input 是策略渲染的统一输入

type Input struct {
	Instance InstanceConfigSnapshot
	Strategy ReleaseStrategySnapshot

	ImageRef    string
	ProjectKey  string
	ServiceName string
	Env         string
}

// Output 是策略渲染的统一输出

type Output struct {
	WorkloadYAML string // 主工作负载 YAML
	ResourceYAML string // 配套资源 YAML（Service/ConfigMap/Ingress 等）
}
