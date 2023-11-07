package cmd

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

// import (
// 	"bytes"
// 	"encoding/base64"

// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// )

type XMLResult struct {
	XMLName xml.Name `xml:"result"`
	Ids     []string `xml:"Id"`
}

// def call(Map parameters) {
//     def artifact = parameters.artifact ?: 'Undefined'
//     // Your custom step logic here, using integrationFlowId
// }

func integrationArtifactsGet(config integrationArtifactsGetOptions, telemetryData *telemetry.CustomData) []string {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	output, err := runIntegrationArtifactsGet(&config, telemetryData, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

	return output
}

func runIntegrationArtifactsGet(config *integrationArtifactsGetOptions, telemetryData *telemetry.CustomData, httpClient piperhttp.Sender) ([]string, error) {

	header := make(http.Header)
	header.Add("Accept", "application/zip")

	// Add Basic Authentication credentials
	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return nil, err
	}
	
	basicAuth := serviceKey.OAuth.Username + ":" + serviceKey.OAuth.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(basicAuth))
	header.Add("Authorization", authHeader)

	getArtifactsURL := fmt.Sprintf("%s/api/v1/IntegrationPackages('%s')/IntegrationDesigntimeArtifacts", serviceKey.OAuth.Host, config.PackageID)
	
	// clientOptions := piperhttp.ClientOptions{}
	// tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	// token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to fetch Bearer Token")
	// }
	// clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	// httpClient.SetOptions(clientOptions)
	
	log.Entry().
			WithField("Get packages", serviceKey.OAuth.Username).
			Info("JALAL JALAL")


	httpMethod := "GET"
	response, httpErr := httpClient.SendRequest(httpMethod, getArtifactsURL, nil, header, nil)
	if httpErr != nil {
		return nil, errors.Wrapf(httpErr, "HTTP %v request to %v failed with error", httpMethod, getArtifactsURL)
	}
	if response == nil {
		return nil, errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	output := []string{}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if response.StatusCode == 200 {
		responseBody, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return nil, errors.Wrapf(readErr, "HTTP response body could not be read, Response status code : %v", response.StatusCode)
		}

		var result XMLResult
		err = xml.Unmarshal(responseBody, &result)
		if err != nil {
			return nil, errors.Errorf("Error parsing XML: %v\n", err)
		}

		output = result.Ids
		return output, nil
	}

	return nil, errors.Errorf("get integration artifacts by package id: %v failed, Response Status code: %v", config.PackageID, response.StatusCode)
}
