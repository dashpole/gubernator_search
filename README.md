# gubernator_search
search http://gcsweb.k8s.io/gcs/kubernetes-jenkins

example usage:  
Search kubelet logs for panic only in the `ci-cri-containerd-e2e-gci-gce-slow` suite  
`go run main.go --pattern="runtime.panic" --file_name="kubelet.log" --url="/gcs/kubernetes-jenkins/logs/ci-cri-containerd-e2e-gci-gce-slow"`
