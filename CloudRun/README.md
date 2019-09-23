# PATH on CodeCard

Cloud Run project to give PATH station information to Code Card.

https://github.com/knative/docs/tree/175313457f94baa036b400f12d162157edef70a7/community/samples/serving/helloworld-rust


Update image via Cloud Builds:

```
~/Documents/CodeCard/CloudRun $ gcloud builds submit --tag gcr.io/codecard/path-codecard
```

Update Cloud Run image via Google Cloud Console:
https://console.cloud.google.com/run/detail/us-east1/hello/revisions?authuser=1&project=codecard
