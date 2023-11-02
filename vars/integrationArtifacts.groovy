@Grab(group='org.apache.httpcomponents', module='httpclient', version='4.5.13')
import org.apache.http.HttpResponse
import org.apache.http.client.methods.HttpGet
import org.apache.http.impl.client.HttpClients
import org.apache.http.util.EntityUtils
import groovy.xml.*

def call(String packageName) {
    // Any valid steps can be called from this code, just like
    // in a Scripted Pipeline
    echo "[-------- ERROR --------]: ${message}"


    String username = "S0025779172"
    String password = "zyThaf-zorjon-0zurdo"
    String API_ENDPOINT = "5fce2be4trial.it-cpitrial06.cfapps.us10-001.hana.ondemand.com"
    API_ENDPOINT = "c5aa4a83trial.it-cpitrial06.cfapps.us10-001.hana.ondemand.com"
    String apiUrl = "https://$API_ENDPOINT/api/v1/IntegrationPackages('$packageName')/IntegrationDesigntimeArtifacts"

    def httpClient = HttpClients.createDefault()
    def httpGet = new HttpGet(apiUrl)
    def credentials = "${username}:${password}".bytes.encodeBase64().toString()
    httpGet.setHeader("Authorization", "Basic ${credentials}")
    try {
      HttpResponse response = httpClient.execute(httpGet)
      if (response.statusLine.statusCode == 200) {
        def responseBody = EntityUtils.toString(response.entity)
        def root = new XmlSlurper().parseText(responseBody)
        def properties = root."**".findAll { it.name() == 'properties' }
        properties.each { property ->
          println(property.Name)
          echo "SARAH: ${property.Name}"
        }
      } else {
        println("Request failed with status code: ${response.statusLine.statusCode}")
      }
    } finally {
      // Close the HTTP client
      httpClient.close()
    }
}