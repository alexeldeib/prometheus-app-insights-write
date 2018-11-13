# prometheus-app-insights-write

[![CircleCI](https://circleci.com/gh/alexeldeib/prometheus-app-insights-write.svg?style=svg)](https://circleci.com/gh/alexeldeib/prometheus-app-insights-write)

Experimental remote write adapter from Prometheus to Azure Application Insights.

## Usage

```bash
helm install -n app-insights-connect --set ${instrumentation_key} https://github.com/alexeldeib/prometheus-app-insights-write/releases/download/${VERSION}/prometheus-app-insights-write-${VERSION}.tgz>
helm install -n app-insights-connect --set 5e02e9a5-22df-48ef-8f63-e45bbb63de60 https://github.com/alexeldeib/prometheus-app-insights-write/releases/download/1.0.0-alpha/prometheus-app-insights-write-1.0.0-alpha.tgz # For example
```

Navigate to the application in the Azure portal and go to search -> analytics. Try the following query:

```sql
customMetrics
| summarize count() by name, value
```

The output should be a list of Prometheus metrics. Labels from Prometheus metrics will be attached as customDimensions. Example:

```sql
customMetrics
| where name == "nginx_ingress_controller_requests"
| summarize count() by tostring(customDimensions.status), bin(timestamp, 1h) 
| render timechart
```
