import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactsGet.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'cpiApiServiceKeyCredentialsId', env: ['PIPER_apiServiceKey']]
    ]

    echo "CALLED call() within integrationArtifactsGet.groovy"
    echo STEP_NAME
    echo METADATA_FILE
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
