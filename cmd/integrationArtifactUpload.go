package cmd

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"



	"github.com/Jeffail/gabs/v2"
	"github.com/SAP/jenkins-library/pkg/cpi"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func integrationArtifactUpload(config integrationArtifactUploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	httpClient := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}
	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runIntegrationArtifactUpload(&config, telemetryData, fileUtils, httpClient)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runIntegrationArtifactUpload(config *integrationArtifactUploadOptions, telemetryData *telemetry.CustomData, fileUtils piperutils.FileUtils, httpClient piperhttp.Sender) error {

	serviceKey, err := cpi.ReadCpiServiceKey(config.APIServiceKey)
	if err != nil {
		return err
	}

	clientOptions := piperhttp.ClientOptions{}
	header := make(http.Header)
	header.Add("Accept", "application/json")
	tokenParameters := cpi.TokenParameters{TokenURL: serviceKey.OAuth.OAuthTokenProviderURL, Username: serviceKey.OAuth.ClientID, Password: serviceKey.OAuth.ClientSecret, Client: httpClient}
	token, err := cpi.CommonUtils.GetBearerToken(tokenParameters)
	if err != nil {
		return errors.Wrap(err, "failed to fetch Bearer Token")
	}
	clientOptions.Token = fmt.Sprintf("Bearer %s", token)
	httpClient.SetOptions(clientOptions)
	httpMethod := "GET"

	// Jalal Addition
	// we first check if package exists
	integrationPackageURL := fmt.Sprintf("%s/api/v1/IntegrationPackages(Id='%s')", serviceKey.OAuth.Host, config.PackageID)
	integrationPackageResp, httpErr := httpClient.SendRequest(httpMethod, integrationPackageURL, nil, header, nil)
	if integrationPackageResp != nil && integrationPackageResp.Body != nil {
		defer integrationPackageResp.Body.Close()
	}
	if integrationPackageResp.StatusCode == 200 {
		log.Entry().
			WithField("PackageID", config.PackageID).
			Info("PackageId DOES exist...")
	} else {
		log.Entry().
			WithField("PackageID", config.PackageID).
			Info("PackageId DOES NOT exist...")

		create_packge_if_required(config, serviceKey.OAuth.Username, serviceKey.OAuth.Password, serviceKey.OAuth.Host, config.PackageID)
		
	}


	//Check availability of integration artefact in CPI design time
	iFlowStatusServiceURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')", serviceKey.OAuth.Host, config.IntegrationFlowID, "Active")
	iFlowStatusResp, httpErr := httpClient.SendRequest("GET", iFlowStatusServiceURL, nil, header, nil)

	if iFlowStatusResp != nil && iFlowStatusResp.Body != nil {
		defer iFlowStatusResp.Body.Close()
	}
	if iFlowStatusResp.StatusCode == 200 {
		return UpdateIntegrationArtifact(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	} else if httpErr != nil && iFlowStatusResp.StatusCode == 404 {
		return UploadIntegrationArtifact(config, httpClient, fileUtils, serviceKey.OAuth.Host)
	}

	if iFlowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if httpErr != nil {
		responseBody, readErr := io.ReadAll(iFlowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", iFlowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, iFlowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, iFlowStatusServiceURL, string(responseBody))
	}
	return errors.Errorf("Failed to check integration flow availability, Response Status code: %v", iFlowStatusResp.StatusCode)
}

// UploadIntegrationArtifact - Upload new integration artifact
func UploadIntegrationArtifact(config *integrationArtifactUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "POST"
	uploadIflowStatusURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts", apiHost)
	header := make(http.Header)
	header.Add("content-type", "application/json")
	payload, jsonError := GetJSONPayloadAsByteArray(config, "create", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
	}

	uploadIflowStatusResp, httpErr := httpClient.SendRequest(httpMethod, uploadIflowStatusURL, payload, header, nil)

	if uploadIflowStatusResp != nil && uploadIflowStatusResp.Body != nil {
		defer uploadIflowStatusResp.Body.Close()
	}

	if uploadIflowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if uploadIflowStatusResp.StatusCode == http.StatusCreated {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully created integration flow artefact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := io.ReadAll(uploadIflowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", uploadIflowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", responseBody, uploadIflowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, uploadIflowStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to create Integration Flow artefact, Response Status code: %v", uploadIflowStatusResp.StatusCode)
}

// UpdateIntegrationArtifact - Update existing integration artifact
func UpdateIntegrationArtifact(config *integrationArtifactUploadOptions, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils, apiHost string) error {
	httpMethod := "PUT"
	header := make(http.Header)
	header.Add("content-type", "application/json")
	updateIflowStatusURL := fmt.Sprintf("%s/api/v1/IntegrationDesigntimeArtifacts(Id='%s',Version='%s')", apiHost, config.IntegrationFlowID, "Active")
	payload, jsonError := GetJSONPayloadAsByteArray(config, "update", fileUtils)
	if jsonError != nil {
		return errors.Wrapf(jsonError, "Failed to get json payload for file %v, failed with error", config.FilePath)
	}
	updateIflowStatusResp, httpErr := httpClient.SendRequest(httpMethod, updateIflowStatusURL, payload, header, nil)

	if updateIflowStatusResp != nil && updateIflowStatusResp.Body != nil {
		defer updateIflowStatusResp.Body.Close()
	}

	if updateIflowStatusResp == nil {
		return errors.Errorf("did not retrieve a HTTP response: %v", httpErr)
	}

	if updateIflowStatusResp.StatusCode == http.StatusOK {
		log.Entry().
			WithField("IntegrationFlowID", config.IntegrationFlowID).
			Info("Successfully updated integration flow artefact in CPI designtime")
		return nil
	}
	if httpErr != nil {
		responseBody, readErr := io.ReadAll(updateIflowStatusResp.Body)
		if readErr != nil {
			return errors.Wrapf(readErr, "HTTP response body could not be read, Response status code: %v", updateIflowStatusResp.StatusCode)
		}
		log.Entry().Errorf("a HTTP error occurred! Response body: %v, Response status code: %v", string(responseBody), updateIflowStatusResp.StatusCode)
		return errors.Wrapf(httpErr, "HTTP %v request to %v failed with error: %v", httpMethod, updateIflowStatusURL, string(responseBody))
	}
	return errors.Errorf("Failed to update Integration Flow artefact, Response Status code: %v", updateIflowStatusResp.StatusCode)
}

// GetJSONPayloadAsByteArray -return http payload as byte array
func GetJSONPayloadAsByteArray(config *integrationArtifactUploadOptions, mode string, fileUtils piperutils.FileUtils) (*bytes.Buffer, error) {
	fileContent, readError := fileUtils.FileRead(config.FilePath)
	if readError != nil {
		return nil, errors.Wrapf(readError, "Error reading file")
	}
	jsonObj := gabs.New()
	if mode == "create" {
		jsonObj.Set(config.IntegrationFlowName, "Name")
		jsonObj.Set(config.IntegrationFlowID, "Id")
		jsonObj.Set(config.PackageID, "PackageId")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else if mode == "update" {
		jsonObj.Set(config.IntegrationFlowName, "Name")
		jsonObj.Set(b64.StdEncoding.EncodeToString(fileContent), "ArtifactContent")
	} else {
		return nil, fmt.Errorf("Unkown node: '%s'", mode)
	}

	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, errors.Wrapf(jsonErr, "json payload is invalid for integration flow artifact %q", config.IntegrationFlowID)
	}
	return bytes.NewBuffer(jsonBody), nil
}

func GetPackageJSONPayloadAsByteArray(config *integrationArtifactUploadOptions) (*bytes.Buffer, error) {

	jsonObj := gabs.New()

	jsonObj.Set(config.PackageID, "Name")
	jsonObj.Set(config.PackageID, "Id")
	jsonObj.Set(config.PackageID, "Description")
	jsonObj.Set(config.PackageID, "ShortText")
	jsonObj.Set("1.0.0", "Version")

	jsonObj.Set("SAP Cloud Integration", "SupportedPlatform")
	
	jsonBody, jsonErr := json.Marshal(jsonObj)

	if jsonErr != nil {
		return nil, errors.Wrapf(jsonErr, "json payload is invalid for integration flow artifact %q", config.IntegrationFlowID)
	}
	fmt.Printf("Package creation JSON body: %s\n", string(jsonBody))
	return bytes.NewBuffer(jsonBody), nil

}

func fetch_xCSRFToken_and_cookie(username, password, endpoint string) (string, string, error) {
	client := &http.Client{}

	// Create an HTTP request with Basic Authentication
	url := endpoint + "?$top=1"
	fmt.Printf("x-csrf-token URL: %s\n", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return "", "", err }

	req.SetBasicAuth(username, password)
	req.Header.Add("x-csrf-token", "fetch")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil { return "", "", err }
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "","", fmt.Errorf("Request failed with status code: %d", resp.StatusCode)
	}

	// Extract the X-CSRF-Token value from the response headers
	xCSRFToken := resp.Header.Get("X-CSRF-Token")
	csrfCookie := resp.Header.Get("Set-Cookie")

	return xCSRFToken, csrfCookie, nil
}


func create_packge_if_required(config *integrationArtifactUploadOptions, username, password, apiEndpoint, packageId string)  {

	payload, err := GetPackageJSONPayloadAsByteArray(config)

	apiUrl := fmt.Sprintf("%s/api/v1/IntegrationPackages", apiEndpoint)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiUrl, payload)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

		
	// get the x-csrf-token
	csrfToken, csrfCookie, err := fetch_xCSRFToken_and_cookie(username, password, apiUrl)
	if err != nil { fmt.Println("xCSRFToken Error: ", err)}
	fmt.Println("xCSRFToken: ", csrfToken)

	
	req.Header.Add("Cookie", csrfCookie)
	req.Header.Add("x-csrf-token", csrfToken)
	
	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil { fmt.Println("httpRequest Error: ", err) }
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("Response Code: %d\n", resp.StatusCode)
	}

	// Read the response body
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error-5: ", err)
	// }

	// fmt.Println("Response:", string(responseBody))
}