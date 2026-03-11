apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ .ProjectKey }}-{{ .ServiceName }}-{{ .Env }}
  namespace: {{ .ArgoNamespace }}
  labels:
    project: {{ .ProjectKey }}
    service: {{ .ServiceName }}
    env: {{ .Env }}
    planId: "{{ .PlanID }}"
    deployment_mode: "{{ .DeploymentMode }}"
spec:
  project: {{ .ArgoProject }}
  source:
    repoURL: {{ .GitOpsRepoURL }}
    targetRevision: {{ .GitOpsBranch }}
    path: {{ .ManifestPath }}
    directory:
      recurse: true
  destination:
    server: {{ .ClusterServer }}
    namespace: {{ .Namespace }}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
