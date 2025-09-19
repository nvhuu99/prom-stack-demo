# Refs:

* List of metrics: https://kubernetes.io/docs/reference/instrumentation/metrics/#list-of-stable-kubernetes-metrics
* LGTM Helm: https://github.com/grafana/helm-charts/tree/main/charts/lgtm-distributed
* Kube Prom Stack: https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack
* Alloy Helm: https://github.com/grafana/alloy/tree/main/operations/helm/charts/alloy

# Setup:

### Prepare:

	cd k8s
	kubectl create namespace monitoring
	kubectl config set-context --current --namespace monitoring

### Create configs and secrets:

	kubectl delete secret grafana-admin
	kubectl create secret generic grafana-admin --from-literal=admin-user=admin --from-literal=admin-password=admin

	kubectl delete configmap alloy-config alloy-endpoints
	kubectl create configmap alloy-config --from-file alloy/config.alloy
	kubectl create configmap alloy-endpoints --from-file alloy/endpoints.json

### Install helm charts:

	# clean
	kubectl delete all --all -n monitoring

	# add repo
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update
	
	# install kube-prometheus-stack (included Grafana)
	helm install prom-stack prometheus-community/kube-prometheus-stack -f values-kube-prometheus-stack.yaml --timeout 10m0s

	# install Tempo (tracing)
	helm install tempo grafana/tempo-distributed -f values-tempo.yaml

	# install Loki (logs)
	helm install loki grafana/loki-distributed -f values-loki.yaml

	# install Alloy (otlp collector)
	helm install alloy grafana/alloy -f values-alloy.yaml

	# verify resources
	kubectl get all -l app.kubernetes.io/instance=prom-stack
	kubectl get all -l app.kubernetes.io/name=tempo
	kubectl get all -l app.kubernetes.io/name=loki-distributed
	kubectl get all -l app.kubernetes.io/name=alloy

### Port forwad:

	kubectl port-forward svc/prom-stack-grafana 30000:3000
	kubectl port-forward svc/alloy 30001:3000

### Run load simulation (webserver):

	docker run -i -v "C:\Users\huu.nv\Desktop\sources\demo-observability\k6\scripts:/scripts" --cpus=0.5 --memory=500m grafana/k6 run "/scripts/test.js"

### Operations:

	# Accessing Grafana UI
	kubectl port-forward svc/prom-stack-grafana 3000:80

	# Copying dashboard
	curl -s -H "Authorization: Bearer $GRAFANA_API_KEY" http://localhost:3000/api/dashboards/uid/<uid> | jq . > my-dashboard.json

	# Update helm charts
	helm upgrade prom-stack prometheus-community/kube-prometheus-stack -f values-kube-prometheus-stack.yaml
	helm upgrade tempo grafana/tempo-distributed -f values-tempo.yaml
	helm upgrade loki grafana/loki-distributed -f values-loki.yaml
	helm upgrade alloy grafana/alloy -f values-alloy.yaml

	# Restart pod when necessary
	kubectl rollout restart deploy lgtm-grafana
	kubectl rollout restart daemonset -l app.kubernetes.io/name=alloy