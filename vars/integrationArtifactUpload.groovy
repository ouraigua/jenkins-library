import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactUpload.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'cpiApiServiceKeyCredentialsId', env: ['PIPER_apiServiceKey']]
    ]

    // Access the 'integrationFlowId' parameter from the 'parameters' map
    String integrationFlowId = parameters.integrationFlowId ?: 'TestFlow'

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
