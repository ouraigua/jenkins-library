@Grab(group='org.apache.httpcomponents', module='httpclient', version='4.5.13')
import org.apache.http.HttpResponse
import org.apache.http.client.methods.HttpGet
import org.apache.http.impl.client.HttpClients
import org.apache.http.util.EntityUtils
import groovy.xml.*

def call(String username, String password, String apiEndpoint, String packageName) {
    echo "[-------- Getting all artifacts for package: $packageName --------]"

    String apiUrl = "https://$apiEndpoint/api/v1/IntegrationPackages('$packageName')/IntegrationDesigntimeArtifacts"

    def httpClient = HttpClients.createDefault()
    def httpGet = new HttpGet(apiUrl) 
    def credentials = "${username}:${password}".bytes.encodeBase64().toString()
    httpGet.setHeader("Authorization", "Basic ${credentials}")
    def output = []

    try {
      HttpResponse response = httpClient.execute(httpGet)
      if (response.statusLine.statusCode == 200) {
        def responseBody = EntityUtils.toString(response.entity)
        def xmlParser = new XmlSlurper().parseText(responseBody)

        xmlParser.'**'.each { node ->
          if (node.name() == 'Name') {
            // echo "---> ${node.text()}"
            output << node.text()
          }
        }

      } else {
        echo "Request failed with status code: ${response.statusLine.statusCode}"
      }
    } finally {
      // Close the HTTP client
      httpClient.close()
      return output
    }
}