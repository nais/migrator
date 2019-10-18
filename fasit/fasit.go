package fasit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/nais/migrator/models/naisd"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"regexp"
)

const (
	NavTruststoreFasitAlias = "nav_truststore"
)

type ResourcePayload interface{}

type RestResourcePayload struct {
	Alias      string         `json:"alias"`
	Scope      Scope          `json:"scope"`
	Type       string         `json:"type"`
	Properties RestProperties `json:"properties"`
}
type WebserviceResourcePayload struct {
	Alias      string               `json:"alias"`
	Scope      Scope                `json:"scope"`
	Type       string               `json:"type"`
	Properties WebserviceProperties `json:"properties"`
}
type WebserviceProperties struct {
	EndpointUrl   string `json:"endpointUrl"`
	WsdlUrl       string `json:"wsdlUrl"`
	SecurityToken string `json:"securityToken"`
	Description   string `json:"description,omitempty"`
}
type RestProperties struct {
	Url         string `json:"url"`
	Description string `json:"description,omitempty"`
}
type Scope struct {
	EnvironmentClass string `json:"environmentclass"`
	Environment      string `json:"environment,omitempty"`
	Zone             string `json:"zone,omitempty"`
}

type Password struct {
	Ref string `json:"ref"`
}
type ApplicationInstancePayload struct {
	Application      string     `json:"application"`
	Environment      string     `json:"environment"`
	Version          string     `json:"version"`
	ExposedResources []Resource `json:"exposedresources"`
	UsedResources    []Resource `json:"usedresources"`
	ClusterName      string     `json:"clustername"`
	Domain           string     `json:"domain"`
}

type Resource struct {
	Id int `json:"id"`
}

type FasitClient struct {
	FasitUrl string
	Username string
	Password string
}

type FasitClientAdapter interface {
	getScopedResource(resourcesRequest ResourceRequest, fasitEnvironment, application, zone string) (NaisResource, naisd.AppError)
	createResource(resource naisd.ExposedResource, fasitEnvironmentClass, fasitEnvironment, hostname string, deploymentRequest naisd.Deploy) (int, error)
	updateResource(existingResource NaisResource, resource naisd.ExposedResource, fasitEnvironmentClass, fasitEnvironment, hostname string, deploymentRequest naisd.Deploy) (int, error)
	GetFasitEnvironmentClass(environmentName string) (string, error)
	GetFasitApplication(application string) error
	GetScopedResources(resourcesRequests []ResourceRequest, fasitEnvironment string, application string, zone string) (resources []NaisResource, err error)
	getLoadBalancerConfig(application string, fasitEnvironment string) (*NaisResource, error)
	createApplicationInstance(deploymentRequest naisd.Deploy, fasitEnvironment, subDomain string, exposedResourceIds, usedResourceIds []int) error
}

type FasitResource struct {
	Id           int
	Alias        string
	ResourceType string `json:"type"`
	Scope        Scope  `json:"scope"`
	Properties   map[string]string
	Secrets      map[string]map[string]string
	Certificates map[string]interface{} `json:"files"`
}

type FasitIngress struct {
	Host string
	Path string
}

type ResourceRequest struct {
	Alias        string
	ResourceType string
	PropertyMap  map[string]string
}

type NaisResource struct {
	id           int
	name         string
	resourceType string
	scope        Scope
	properties   map[string]string
	propertyMap  map[string]string
	secret       map[string]string
	certificates map[string][]byte
	ingresses    []FasitIngress
}

func DefaultResourceRequests() []ResourceRequest {
	return []ResourceRequest{
		{
			Alias:        NavTruststoreFasitAlias,
			ResourceType: "certificate",
			PropertyMap:  map[string]string{"keystore": "NAV_TRUSTSTORE_PATH"},
		},
	}
}

func (nr NaisResource) Properties() map[string]string {
	return nr.properties
}

func (nr NaisResource) Secret() map[string]string {
	return nr.secret
}

func (nr NaisResource) Certificates() map[string][]byte {
	return nr.certificates
}

func (nr NaisResource) ToEnvironmentVariable(property string) string {
	return strings.ToUpper(nr.ToResourceVariable(property))
}

func (nr NaisResource) ToResourceVariable(property string) string {
	if value, ok := nr.propertyMap[property]; ok {
		property = value
	} else if nr.resourceType != "applicationproperties" {
		property = nr.name + "_" + property
	}

	return strings.ToLower(normalizePropertyName(property))
}

func normalizePropertyName(name string) string {
	if strings.Contains(name, ".") {
		name = strings.Replace(name, ".", "_", -1)
	}

	if strings.Contains(name, ":") {
		name = strings.Replace(name, ":", "_", -1)
	}

	if strings.Contains(name, "-") {
		name = strings.Replace(name, "-", "_", -1)
	}

	return name
}

func (fasit FasitClient) GetScopedResources(resourcesRequests []ResourceRequest, fasitEnvironment string, application string, zone string) (resources []NaisResource, err error) {
	for _, request := range resourcesRequests {
		resource, appErr := fasit.getScopedResource(request, fasitEnvironment, application, zone)
		if appErr != nil {
			return []NaisResource{}, fmt.Errorf("unable to get resource %s (%s). %s", request.Alias, request.ResourceType, appErr)
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (fasit FasitClient) getLoadBalancerConfig(application string, fasitEnvironment string) (*NaisResource, error) {
	req, err := fasit.buildRequest("GET", "/api/v2/resources", map[string]string{
		"environment": fasitEnvironment,
		"application": application,
		"type":        "LoadBalancerConfig",
	})

	body, appErr := fasit.doRequest(req)
	if appErr != nil {
		return nil, err
	}

	ingresses, err := parseLoadBalancerConfig(body)
	if err != nil {
		return nil, err
	}

	if len(ingresses) == 0 {
		return nil, nil
	}

	return &NaisResource{
		name:         "",
		properties:   nil,
		resourceType: "LoadBalancerConfig",
		certificates: nil,
		secret:       nil,
		ingresses:    ingresses,
	}, nil

}

func getResourceIds(usedResources []NaisResource) (usedResourceIds []int) {
	for _, resource := range usedResources {
		if resource.resourceType != "LoadBalancerConfig" {
			usedResourceIds = append(usedResourceIds, resource.id)
		}
	}
	return usedResourceIds
}

func FetchFasitResources(fasit FasitClientAdapter, application string, fasitEnvironment string, zone string, usedResources []naisd.UsedResource) (naisresources []NaisResource, err error) {
	resourceRequests := DefaultResourceRequests()

	for _, resource := range usedResources {
		resourceRequests = append(resourceRequests, ResourceRequest{
			Alias:        resource.Alias,
			ResourceType: resource.ResourceType,
			PropertyMap:  resource.PropertyMap,
		})
	}

	naisresources, err = fasit.GetScopedResources(resourceRequests, fasitEnvironment, application, zone)
	if err != nil {
		return naisresources, err
	}

	if lbResource, e := fasit.getLoadBalancerConfig(application, fasitEnvironment); e == nil {
		if lbResource != nil {
			naisresources = append(naisresources, *lbResource)
		}
	} else {
		log.Warningf("failed getting loadbalancer config for application %s in fasitEnvironment %s: %s ", application, fasitEnvironment, e)
	}

	return naisresources, nil

}
func arrayToString(a []int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", ",", -1), "[]")
}

func (fasit FasitClient) doRequest(r *http.Request) ([]byte, naisd.AppError) {

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		return []byte{}, appError{err, "Error contacting fasit", http.StatusInternalServerError}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, appError{err, "Could not read body", http.StatusInternalServerError}
	}

	if resp.StatusCode == 404 {
		return []byte{}, appError{nil, fmt.Sprintf("item not found in Fasit: %s", string(body)), http.StatusNotFound}
	}

	if resp.StatusCode > 299 {
		return []byte{}, appError{nil, fmt.Sprintf("error contacting Fasit: %s", string(body)), resp.StatusCode}
	}

	return body, nil

}
func (fasit FasitClient) getScopedResource(resourcesRequest ResourceRequest, fasitEnvironment, application, zone string) (NaisResource, naisd.AppError) {
	req, err := fasit.buildRequest("GET", "/api/v2/scopedresource", map[string]string{
		"alias":       resourcesRequest.Alias,
		"type":        resourcesRequest.ResourceType,
		"environment": fasitEnvironment,
		"application": application,
		"zone":        zone,
	})

	if err != nil {
		return NaisResource{}, appError{err, "unable to create request", 500}
	}

	body, appErr := fasit.doRequest(req)
	if appErr != nil {
		return NaisResource{}, appErr
	}

	var fasitResource FasitResource

	err = json.Unmarshal(body, &fasitResource)
	if err != nil {
		return NaisResource{}, appError{err, "could not unmarshal body", 500}
	}

	resource, err := fasit.mapToNaisResource(fasitResource, resourcesRequest.PropertyMap)
	if err != nil {
		return NaisResource{}, appError{err, "unable to map response to Nais resource", 500}
	}
	return resource, nil
}

func SafeMarshal(v interface{}) ([]byte, error) {
	/*	String values encode as JSON strings coerced to valid UTF-8, replacing invalid bytes with the Unicode replacement rune.
		The angle brackets "<" and ">" are escaped to "\u003c" and "\u003e" to keep some browsers from misinterpreting JSON output as HTML.
		Ampersand "&" is also escaped to "\u0026" for the same reason. This escaping can be disabled using an Encoder that had SetEscapeHTML(false) called on it.	*/
	b, err := json.Marshal(v)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	return b, err
}

func (fasit FasitClient) GetFasitEnvironmentClass(environmentName string) (string, error) {
	req, err := http.NewRequest("GET", fasit.FasitUrl+"/api/v2/environments/"+environmentName, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %s", err)
	}

	resp, appErr := fasit.doRequest(req)
	if appErr != nil {
		return "", appErr
	}

	type FasitEnvironment struct {
		EnvironmentClass string `json:"environmentclass"`
	}
	var fasitEnvironment FasitEnvironment
	if err := json.Unmarshal(resp, &fasitEnvironment); err != nil {
		return "", fmt.Errorf("unable to read environmentclass from response: %s", err)
	}

	return fasitEnvironment.EnvironmentClass, nil
}

func (fasit FasitClient) GetFasitApplication(application string) error {
	req, err := http.NewRequest("GET", fasit.FasitUrl+"/api/v2/applications/"+application, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %s", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to contact Fasit: %s", err)
	}
	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("error contacting Fasit: %s", err)
	}

	if resp.StatusCode == 200 {
		return nil
	}
	return fmt.Errorf("could not find application %s in Fasit", application)
}

func (fasit FasitClient) mapToNaisResource(fasitResource FasitResource, propertyMap map[string]string) (resource NaisResource, err error) {
	resource.name = fasitResource.Alias
	resource.resourceType = fasitResource.ResourceType
	resource.properties = fasitResource.Properties
	resource.propertyMap = propertyMap
	resource.id = fasitResource.Id
	resource.scope = fasitResource.Scope

	if len(fasitResource.Secrets) > 0 {
		secret, err := resolveSecret(fasitResource.Secrets, fasit.Username, fasit.Password)
		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to resolve secret: %s", err)
		}
		resource.secret = secret
	}

	if fasitResource.ResourceType == "certificate" && len(fasitResource.Certificates) > 0 {
		files, err := resolveCertificates(fasitResource.Certificates)

		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to resolve Certificates: %s", err)
		}

		resource.certificates = files

	} else if fasitResource.ResourceType == "applicationproperties" {
		lineFilter, err := regexp.Compile(`^[\p{L}\d_.]+=.+`)
		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to compile regex: %s", err)
		}

		for _, line := range strings.Split(fasitResource.Properties["applicationProperties"], "\n") {
			line = strings.TrimSpace(line)
			if lineFilter.MatchString(line) {
				parts := strings.SplitN(line, "=", 2)
				resource.properties[parts[0]] = parts[1]
			} else if len(line) > 0 {
				log.Infof("the following string did not match our regex: %s", line)
			}
		}
		delete(resource.properties, "applicationProperties")
	}

	return resource, nil
}
func resolveCertificates(files map[string]interface{}) (map[string][]byte, error) {
	fileContent := make(map[string][]byte)

	fileName, fileUrl, err := parseFilesObject(files)
	if err != nil {
		return fileContent, err
	}

	response, err := http.Get(fileUrl)
	if err != nil {
		return fileContent, fmt.Errorf("error contacting fasit when resolving file: %s", err)
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fileContent, fmt.Errorf("error downloading file: %s", err)
	}

	fileContent[fileName] = bodyBytes
	return fileContent, nil

}

func parseLoadBalancerConfig(config []byte) ([]FasitIngress, error) {
	jsn, err := gabs.ParseJSON(config)
	if err != nil {
		return nil, fmt.Errorf("error parsing load balancer config: %s ", config)
	}

	ingresses := make([]FasitIngress, 0)
	lbConfigs, _ := jsn.Children()
	if len(lbConfigs) == 0 {
		return nil, nil
	}

	for _, lbConfig := range lbConfigs {
		host, found := lbConfig.Path("properties.url").Data().(string)
		if !found {
			log.Warningf("no host found for loadbalancer config: %s", lbConfig)
			continue
		}
		pathList, _ := lbConfig.Path("properties.contextRoots").Data().(string)
		paths := strings.Split(pathList, ",")
		for _, path := range paths {
			ingresses = append(ingresses, FasitIngress{Host: host, Path: path})
		}
	}

	if len(ingresses) == 0 {
		return nil, fmt.Errorf("no loadbalancer config found for: %s", config)
	}
	return ingresses, nil
}

func parseFilesObject(files map[string]interface{}) (fileName string, fileUrl string, e error) {
	jsn, err := gabs.Consume(files)
	if err != nil {
		return "", "", fmt.Errorf("error parsing fasit json: %s ", files)
	}

	fileName, fileNameFound := jsn.Path("keystore.filename").Data().(string)
	if !fileNameFound {
		return "", "", fmt.Errorf("error parsing fasit json. Filename not found: %s ", files)
	}

	fileUrl, fileUrlfound := jsn.Path("keystore.ref").Data().(string)
	if !fileUrlfound {
		return "", "", fmt.Errorf("error parsing fasit json. Fileurl not found: %s ", files)
	}

	return fileName, fileUrl, nil
}

func resolveSecret(secrets map[string]map[string]string, username string, password string) (map[string]string, error) {

	req, err := http.NewRequest("GET", secrets[getFirstKey(secrets)]["ref"], nil)

	if err != nil {
		return map[string]string{}, err
	}

	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]string{}, fmt.Errorf("error contacting fasit when resolving secret: %s", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode > 299 {
		if requestDump, e := httputil.DumpRequest(req, false); e == nil {
			log.Error("Fasit request: ", string(requestDump))
		}
		return map[string]string{}, fmt.Errorf("fasit gave error message when resolving secret: %s (HTTP %v)", body, strconv.Itoa(resp.StatusCode))
	}

	return map[string]string{"password": string(body)}, nil
}

func getFirstKey(m map[string]map[string]string) string {
	if len(m) > 0 {
		for key := range m {
			return key
		}
	}
	return ""
}

func (fasit FasitClient) buildRequest(method, path string, queryParams map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, fasit.FasitUrl+path, nil)

	if err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}

	q := req.URL.Query()

	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func generateScope(resource naisd.ExposedResource, existingResource NaisResource, fasitEnvironmentClass, fasitEnvironment, zone string) Scope {
	if resource.AllZones {
		return Scope{
			EnvironmentClass: fasitEnvironmentClass,
			Environment:      fasitEnvironment,
		}
	}
	if existingResource.id > 0 {
		return existingResource.scope
	}
	return Scope{
		EnvironmentClass: fasitEnvironmentClass,
		Environment:      fasitEnvironment,
		Zone:             zone,
	}
}

func buildApplicationInstancePayload(deploymentRequest naisd.Deploy, fasitEnvironment, subDomain string, exposedResourceIds, usedResourceIds []int) ApplicationInstancePayload {
	// Need to make an empty array of Resources in order for json.Marshall to return [] and not null
	// see https://danott.co/posts/json-marshalling-empty-slices-to-empty-arrays-in-go.html for details
	emptyResources := make([]Resource, 0)
	domain := strings.Join(strings.Split(subDomain, ".")[1:], ".")
	applicationInstancePayload := ApplicationInstancePayload{
		Application:      deploymentRequest.Application,
		Environment:      fasitEnvironment,
		Version:          deploymentRequest.Version,
		ClusterName:      "nais",
		Domain:           domain,
		ExposedResources: emptyResources,
		UsedResources:    emptyResources,
	}
	if len(exposedResourceIds) > 0 {
		for _, id := range exposedResourceIds {
			applicationInstancePayload.ExposedResources = append(applicationInstancePayload.ExposedResources, Resource{id})
		}
	}
	if len(usedResourceIds) > 0 {
		for _, id := range usedResourceIds {
			applicationInstancePayload.UsedResources = append(applicationInstancePayload.UsedResources, Resource{id})
		}
	}

	return applicationInstancePayload
}

func buildResourcePayload(resource naisd.ExposedResource, existingResource NaisResource, fasitEnvironmentClass, fasitEnvironment, zone, hostname string) ResourcePayload {
	// Reference of valid resources in Fasit
	// ['DataSource', 'MSSQLDataSource', 'DB2DataSource', 'LDAP', 'BaseUrl', 'Credential', 'Certificate', 'OpenAm', 'Cics', 'RoleMapping', 'QueueManager', 'WebserviceEndpoint', 'SoapService', 'RestService', 'WebserviceGateway', 'EJB', 'Datapower', 'EmailAddress', 'SMTPServer', 'Queue', 'Topic', 'DeploymentManager', 'ApplicationProperties', 'MemoryParameters', 'LoadBalancer', 'LoadBalancerConfig', 'FileLibrary', 'Channel
	if strings.EqualFold("restservice", resource.ResourceType) {
		return RestResourcePayload{
			Type:  "RestService",
			Alias: resource.Alias,
			Properties: RestProperties{
				Url:         "https://" + hostname + resource.Path,
				Description: resource.Description,
			},
			Scope: generateScope(resource, existingResource, fasitEnvironmentClass, fasitEnvironment, zone),
		}

	} else if strings.EqualFold("WebserviceEndpoint", resource.ResourceType) || strings.EqualFold("SoapService", resource.ResourceType) {
		Url, _ := url.Parse("http://maven.adeo.no/nexus/service/local/artifact/maven/redirect")
		wsdlArtifactQuery := url.Values{}
		wsdlArtifactQuery.Add("r", "m2internal")
		wsdlArtifactQuery.Add("g", resource.WsdlGroupId)
		wsdlArtifactQuery.Add("a", resource.WsdlArtifactId)
		wsdlArtifactQuery.Add("v", resource.WsdlVersion)
		wsdlArtifactQuery.Add("e", "zip")
		Url.RawQuery = wsdlArtifactQuery.Encode()

		return WebserviceResourcePayload{
			Type:  resource.ResourceType,
			Alias: resource.Alias,
			Properties: WebserviceProperties{
				EndpointUrl:   "https://" + hostname + resource.Path,
				WsdlUrl:       Url.String(),
				SecurityToken: resource.SecurityToken,
				Description:   resource.Description,
			},
			Scope: generateScope(resource, existingResource, fasitEnvironmentClass, fasitEnvironment, zone),
		}
	} else {
		return nil
	}
}
