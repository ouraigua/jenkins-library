import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactsGet.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'cpiApiServiceKeyCredentialsId', env: ['PIPER_apiServiceKey']]
    ]

    def output = piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    echo "Got artifacts: $output"
    return output
}
