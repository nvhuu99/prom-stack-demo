# Refs:

* List of metrics: https://kubernetes.io/docs/reference/instrumentation/metrics/#list-of-stable-kubernetes-metrics
* LGTM Helm: https://github.com/grafana/helm-charts/tree/main/charts/lgtm-distributed
* Loki Helm: https://github.com/grafana/helm-charts/blob/main/charts/loki-distributed/values.yaml
* Kube Prom Stack: https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack

# Setup:

### Prepare:

	kubectl create namespace monitoring
	kubectl config set-context --current --namespace monitoring


### Install helm charts:

  # create configs and secrets:
	kubectl delete secret grafana-admin
	kubectl create secret generic grafana-admin --from-literal=admin-user=admin --from-literal=admin-password=admin
	kubectl delete configmap monitoring-endpoints
	kubectl apply -f k8s/configmap/monitoring-endpoints.yaml

	# clean
	kubectl delete all --all -n monitoring

	# add repo
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update
	
	# install kube-prometheus-stack (included Grafana)
	helm install prom-stack prometheus-community/kube-prometheus-stack \
	  -f k8s/values-kube-prometheus-stack.yaml --timeout 10m0s

	# install Tempo (tracing)
	helm install tempo grafana/tempo-distributed \
	  -f k8s/values-tempo.yaml

	# install Loki (logs)
	helm install loki grafana/loki-distributed \
	  -f k8s/values-loki.yaml

	# verify resources
	kubectl get all -l app.kubernetes.io/instance=prom-stack
	kubectl get all -l app.kubernetes.io/name=tempo
	kubectl get all -l app.kubernetes.io/name=loki-distributed

### Deploy SpringBoot & K6 load simulator:

	# spring boot
	docker build -t demo-observability/spring .
	kind load docker-image demo-observability/spring
	kubectl apply -f k8s/spring-app.yaml
	kubectl get all -l app=demo

	# k6
  kubectl delete configmap k6-script
  kubectl create configmap k6-script --from-file=k6/scripts/test.js
	kubectl apply -f k8s/k6.yaml

### Port forwad:

	kubectl port-forward svc/prom-stack-grafana 30000:3000

### Run load simulation (webserver):

	docker run -i -v "C:\Users\huu.nv\Desktop\sources\demo-observability\k6\scripts:/scripts" --cpus=0.5 --memory=500m grafana/k6 run "/scripts/test.js"

### Operations:

	# Accessing Grafana UI
	kubectl port-forward svc/prom-stack-grafana 3000:80

	# Copying dashboard
	curl -s -H "Authorization: Bearer $GRAFANA_API_KEY" http://localhost:3000/api/dashboards/uid/<uid> | jq . > my-dashboard.json

	# Update helm charts
	helm upgrade prom-stack prometheus-community/kube-prometheus-stack -f k8s/values-kube-prometheus-stack.yaml
	helm upgrade tempo grafana/tempo-distributed -f k8s/values-tempo.yaml
	helm upgrade loki grafana/loki-distributed -f k8s/values-loki.yaml

	# Restart pod when necessary
	kubectl rollout restart deploy prom-stack
	kubectl rollout restart pods tempo
