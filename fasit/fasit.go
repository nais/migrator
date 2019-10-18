package fasit

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/nais/migrator/models/naisd"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
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
	GetFasitEnvironmentClass(environmentName string) (string, error)
	GetFasitApplication(application string) error
	GetScopedResources(resourcesRequests []ResourceRequest, fasitEnvironment string, application string, zone string) (resources []NaisResource, err error)
	getLoadBalancerConfig(application string, fasitEnvironment string) (*NaisResource, error)
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
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	ResourceType string            `json:"resourceType"`
	Scope        Scope             `json:"scope"`
	Properties   map[string]string `json:"properties"`
	PropertyMap  map[string]string `json:"propertyMap"`
	Secret       map[string]string `json:"secret"`
	Certificates map[string][]byte `json:"certificates"`
	Ingresses    []FasitIngress    `json:"ingresses"`
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

func (nr NaisResource) ToEnvironmentVariable(property string) string {
	return strings.ToUpper(nr.ToResourceVariable(property))
}

func (nr NaisResource) ToResourceVariable(property string) string {
	if value, ok := nr.PropertyMap[property]; ok {
		property = value
	} else if nr.ResourceType != "applicationproperties" {
		property = nr.Name + "_" + property
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
		Name:         "",
		Properties:   nil,
		ResourceType: "LoadBalancerConfig",
		Certificates: nil,
		Secret:       nil,
		Ingresses:    ingresses,
	}, nil

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
	resource.Name = fasitResource.Alias
	resource.ResourceType = fasitResource.ResourceType
	resource.Properties = fasitResource.Properties
	resource.PropertyMap = propertyMap
	resource.ID = fasitResource.Id
	resource.Scope = fasitResource.Scope

	if len(fasitResource.Secrets) > 0 {
		secret, err := resolveSecret(fasitResource.Secrets, fasit.Username, fasit.Password)
		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to resolve secret: %s", err)
		}
		resource.Secret = secret
	}

	if fasitResource.ResourceType == "certificate" && len(fasitResource.Certificates) > 0 {
		files, err := resolveCertificates(fasitResource.Certificates)

		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to resolve Certificates: %s", err)
		}

		resource.Certificates = files

	} else if fasitResource.ResourceType == "applicationproperties" {
		lineFilter, err := regexp.Compile(`^[\p{L}\d_.]+=.+`)
		if err != nil {
			return NaisResource{}, fmt.Errorf("unable to compile regex: %s", err)
		}

		for _, line := range strings.Split(fasitResource.Properties["applicationProperties"], "\n") {
			line = strings.TrimSpace(line)
			if lineFilter.MatchString(line) {
				parts := strings.SplitN(line, "=", 2)
				resource.Properties[parts[0]] = parts[1]
			} else if len(line) > 0 {
				log.Infof("the following string did not match our regex: %s", line)
			}
		}
		delete(resource.Properties, "applicationProperties")
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
