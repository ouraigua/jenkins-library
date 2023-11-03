@Grab(group='org.apache.httpcomponents', module='httpclient', version='4.5.13')
import org.apache.http.HttpResponse
import org.apache.http.client.methods.HttpGet
import org.apache.http.impl.client.HttpClients
import org.apache.http.util.EntityUtils
import groovy.xml.*

def call(String username, String password, String apiEndpoint, String packageName, String[] result) {
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
        def root = new XmlSlurper().parseText(responseBody)
        def properties = root."**".findAll { it.name() == 'properties' }
        properties.each { property ->
          output << property.Name
          result << property.Name
          echo "---> ${property.Name}"
        }
      } else {
        echo "Request failed with status code: ${response.statusLine.statusCode}"
      }
    } finally {
      // Close the HTTP client
      httpClient.close()
    }

    return output
}