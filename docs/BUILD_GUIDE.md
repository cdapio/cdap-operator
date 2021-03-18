## Build LiveRamp CDAP-OPERATOR Image

Ensure you are in the root of the `cdap-operator` project subdirectory.  

Define global variables: 

```
export BUILD_PROJECT_ID=liveramp-eng-data-ops-dev
export REGISTRY_CDAP_OPERATOR=eu.gcr.io/liveramp-eng-data-ops-dev/cdap-controller
export VERSION_CDAP_OPERATOR=latest
```

#### Enable kaniko on your local gcloud

`gcloud config set builds/use_kaniko True`

#### Build using cloud build 

```
gcloud builds submit --project=$BUILD_PROJECT_ID --async --config cloudbuild.yaml --substitutions _REGISTRY_CDAP_OPERATOR=$REGISTRY_CDAP_OPERATOR,_VERSION_CDAP_OPERATOR=$VERSION_CDAP_OPERATOR
```

#### View Builds:

You can view cloud builds in progress from the GCP console and complete images at `eu.gcr.io/liveramp-eng-data-ops-dev/cdap-controller:${VERSION_CDAP_OPERATOR}`
