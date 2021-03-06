package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

const (
	ProductionClientId = "3MVG9A2kN3Bn17huXZp1OQhPe8y4_ozAQZZCKxsWbef9GjSnHGOunHSwhnY1BWz_5vHkTL9BeLMriIX5EUKaw"
	PrereleaseClientId = "3MVG9lKcPoNINVBIRgC7lsz5tIhlg0mtoEqkA9ZjDAwEMbBy43gsnfkzzdTdhFLeNnWS8M4bnRnVv1Qj0k9MD"
	RedirectUri        = "https://force-cli.herokuapp.com/auth/callback"
)

const (
	EndpointProduction = iota
	EndpointTest       = iota
	EndpointPrerelease = iota
)

const (
	apiVersion = "v29.0"   //winter 14
)

var RootCertificates = `
-----BEGIN CERTIFICATE-----
MIICPDCCAaUCEHC65B0Q2Sk0tjjKewPMur8wDQYJKoZIhvcNAQECBQAwXzELMAkG
A1UEBhMCVVMxFzAVBgNVBAoTDlZlcmlTaWduLCBJbmMuMTcwNQYDVQQLEy5DbGFz
cyAzIFB1YmxpYyBQcmltYXJ5IENlcnRpZmljYXRpb24gQXV0aG9yaXR5MB4XDTk2
MDEyOTAwMDAwMFoXDTI4MDgwMTIzNTk1OVowXzELMAkGA1UEBhMCVVMxFzAVBgNV
BAoTDlZlcmlTaWduLCBJbmMuMTcwNQYDVQQLEy5DbGFzcyAzIFB1YmxpYyBQcmlt
YXJ5IENlcnRpZmljYXRpb24gQXV0aG9yaXR5MIGfMA0GCSqGSIb3DQEBAQUAA4GN
ADCBiQKBgQDJXFme8huKARS0EN8EQNvjV69qRUCPhAwL0TPZ2RHP7gJYHyX3KqhE
BarsAx94f56TuZoAqiN91qyFomNFx3InzPRMxnVx0jnvT0Lwdd8KkMaOIG+YD/is
I19wKTakyYbnsZogy1Olhec9vn2a/iRFM9x2Fe0PonFkTGUugWhFpwIDAQABMA0G
CSqGSIb3DQEBAgUAA4GBALtMEivPLCYATxQT3ab7/AoRhIzzKBxnki98tsX63/Do
lbwdj2wsqFHMc9ikwFPwTtYmwHYBV4GSXiHx0bH/59AhWM1pF+NEHJwZRDmJXNyc
AA9WjQKZ7aKQRUzkuxCkPfAyAw7xzvjoyVGM5mKf5p/AfbdynMk2OmufTqj/ZA1k
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDxTCCAq2gAwIBAgIQAqxcJmoLQJuPC3nyrkYldzANBgkqhkiG9w0BAQUFADBs
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSswKQYDVQQDEyJEaWdpQ2VydCBIaWdoIEFzc3VyYW5j
ZSBFViBSb290IENBMB4XDTA2MTExMDAwMDAwMFoXDTMxMTExMDAwMDAwMFowbDEL
MAkGA1UEBhMCVVMxFTATBgNVBAoTDERpZ2lDZXJ0IEluYzEZMBcGA1UECxMQd3d3
LmRpZ2ljZXJ0LmNvbTErMCkGA1UEAxMiRGlnaUNlcnQgSGlnaCBBc3N1cmFuY2Ug
RVYgUm9vdCBDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMbM5XPm
+9S75S0tMqbf5YE/yc0lSbZxKsPVlDRnogocsF9ppkCxxLeyj9CYpKlBWTrT3JTW
PNt0OKRKzE0lgvdKpVMSOO7zSW1xkX5jtqumX8OkhPhPYlG++MXs2ziS4wblCJEM
xChBVfvLWokVfnHoNb9Ncgk9vjo4UFt3MRuNs8ckRZqnrG0AFFoEt7oT61EKmEFB
Ik5lYYeBQVCmeVyJ3hlKV9Uu5l0cUyx+mM0aBhakaHPQNAQTXKFx01p8VdteZOE3
hzBWBOURtCmAEvF5OYiiAhF8J2a3iLd48soKqDirCmTCv2ZdlYTBoSUeh10aUAsg
EsxBu24LUTi4S8sCAwEAAaNjMGEwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQF
MAMBAf8wHQYDVR0OBBYEFLE+w2kD+L9HAdSYJhoIAu9jZCvDMB8GA1UdIwQYMBaA
FLE+w2kD+L9HAdSYJhoIAu9jZCvDMA0GCSqGSIb3DQEBBQUAA4IBAQAcGgaX3Nec
nzyIZgYIVyHbIUf4KmeqvxgydkAQV8GK83rZEWWONfqe/EW1ntlMMUu4kehDLI6z
eM7b41N5cdblIZQB2lWHmiRk9opmzN6cN82oNLFpmyPInngiK3BD41VHMWEZ71jF
hS9OMPagMRYjyOfiZRYzy78aG6A9+MpeizGLYAiJLQwGXFK3xPkKmNEVX58Svnw2
Yzi9RKR/5CYrCsSXaQ3pjOLAEFe4yHYSkVXySGnYvCoCWw9E1CAx2/S6cCZdkGCe
vEsXCS+0yx5DaMkHJ8HSXPfqIbloEpw8nL+e/IBcm2PN7EeqJSdnoDfzAIJ9VNep
+OkuE6N36B9K
-----END CERTIFICATE-----`

type Force struct {
	Credentials ForceCredentials
	Metadata    *ForceMetadata
	Partner     *ForcePartner
}

type ForceCredentials struct {
	AccessToken string
	Id          string
	InstanceUrl string
	IssuedAt    string
	Scope       string
}

type ForceError struct {
	Message   string
	ErrorCode string
}

type ForceEndpoint int

type ForceRecord map[string]interface{}

type ForceSobject map[string]interface{}

type ForceCreateRecordResult struct {
	Errors  []string
	Id      string
	Success bool
}

type ForceQueryResult struct {
	Done      bool
	Records   []ForceRecord
	TotalSize int
}

type ForceSobjectsResult struct {
	Encoding     string
	MaxBatchSize int
	Sobjects     []ForceSobject
}

func NewForce(creds ForceCredentials) (force *Force) {
	force = new(Force)
	force.Credentials = creds
	force.Metadata = NewForceMetadata(force)
	force.Partner = NewForcePartner(force)
	return
}

func ForceLogin(endpoint ForceEndpoint) (creds ForceCredentials, err error) {
	ch := make(chan ForceCredentials)
	port, err := startLocalHttpServer(ch)
	var url string
	switch endpoint {
	case EndpointProduction:
		url = fmt.Sprintf("https://login.salesforce.com/services/oauth2/authorize?response_type=token&client_id=%s&redirect_uri=%s&state=%d&prompt=login", ProductionClientId, RedirectUri, port)
	case EndpointTest:
		url = fmt.Sprintf("https://test.salesforce.com/services/oauth2/authorize?response_type=token&client_id=%s&redirect_uri=%s&state=%d&prompt=login", ProductionClientId, RedirectUri, port)
	case EndpointPrerelease:
		url = fmt.Sprintf("https://prerellogin.pre.salesforce.com/services/oauth2/authorize?response_type=token&client_id=%s&redirect_uri=%s&state=%d&prompt=login", PrereleaseClientId, RedirectUri, port)
	default:
		ErrorAndExit("no such endpoint type")
	}
	err = Open(url)
	creds = <-ch
	return
}

func (f *Force) ListSobjects() (sobjects []ForceSobject, err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects", f.Credentials.InstanceUrl, apiVersion)
	body, err := f.httpGet(url)
	if err != nil {
		return
	}
	var result ForceSobjectsResult
	json.Unmarshal(body, &result)
	sobjects = result.Sobjects
	return
}

func (f *Force) GetSobject(name string) (sobject ForceSobject, err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/describe", f.Credentials.InstanceUrl, apiVersion, name)
	body, err := f.httpGet(url)
	if err != nil {
		return
	}
	json.Unmarshal(body, &sobject)
	return
}

func (f *Force) Query(query string) (records []ForceRecord, err error) {
	url := fmt.Sprintf("%s/services/data/%s/query?q=%s", f.Credentials.InstanceUrl, apiVersion, url.QueryEscape(query))
	body, err := f.httpGet(url)
	if err != nil {
		return
	}
	var result ForceQueryResult
	json.Unmarshal(body, &result)
	records = result.Records
	return
}

func (f *Force) Get(url string) (object ForceRecord, err error) {
	body, err := f.httpGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &object)
	return
}

func (f *Force) GetRecord(sobject, id string) (object ForceRecord, err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", f.Credentials.InstanceUrl, apiVersion, sobject, id)
	body, err := f.httpGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &object)
	return
}

func (f *Force) CreateRecord(sobject string, attrs map[string]string) (id string, err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s", f.Credentials.InstanceUrl, apiVersion, sobject)
	body, err := f.httpPost(url, attrs)
	var result ForceCreateRecordResult
	json.Unmarshal(body, &result)
	id = result.Id
	return
}

func (f *Force) UpdateRecord(sobject string, id string, attrs map[string]string) (err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", f.Credentials.InstanceUrl, apiVersion, sobject, id)
	_, err = f.httpPatch(url, attrs)
	return
}

func (f *Force) DeleteRecord(sobject string, id string) (err error) {
	url := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", f.Credentials.InstanceUrl, apiVersion, sobject, id)
	_, err = f.httpDelete(url)
	return
}

func (f *Force) Whoami() (me ForceRecord, err error) {
	parts := strings.Split(f.Credentials.Id, "/")
	me, err = f.GetRecord("User", parts[len(parts)-1])
	return
}

func (f *Force) httpGet(url string) (body []byte, err error) {
	req, err := httpRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", f.Credentials.AccessToken))
	res, err := httpClient().Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		err = errors.New("authorization expired, please run `force login`")
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if res.StatusCode/100 != 2 {
		var messages []ForceError
		json.Unmarshal(body, &messages)
		if len(messages) > 0 {
			err = errors.New(messages[0].Message)
		} else {
			err = errors.New(string(body))
		}
		return
	}
	return
}

func (f *Force) httpPost(url string, attrs map[string]string) (body []byte, err error) {
	rbody, _ := json.Marshal(attrs)
	req, err := httpRequest("POST", url, bytes.NewReader(rbody))
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", f.Credentials.AccessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := httpClient().Do(req)
	defer res.Body.Close()
	if err != nil {
		return
	}
	if res.StatusCode == 401 {
		err = errors.New("authorization expired, please run `force login`")
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if res.StatusCode/100 != 2 {
		var messages []ForceError
		json.Unmarshal(body, &messages)
		err = errors.New(messages[0].Message)
		return
	}
	return
}

func (f *Force) httpPatch(url string, attrs map[string]string) (body []byte, err error) {
	rbody, _ := json.Marshal(attrs)
	req, err := httpRequest("PATCH", url, bytes.NewReader(rbody))
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", f.Credentials.AccessToken))
	req.Header.Add("Content-Type", "application/json")
	res, err := httpClient().Do(req)
	defer res.Body.Close()
	if err != nil {
		return
	}
	if res.StatusCode == 401 {
		err = errors.New("authorization expired, please run `force login`")
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if res.StatusCode/100 != 2 {
		var messages []ForceError
		json.Unmarshal(body, &messages)
		err = errors.New(messages[0].Message)
		return
	}
	return
}

func (f *Force) httpDelete(url string) (body []byte, err error) {
	req, err := httpRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", f.Credentials.AccessToken))
	res, err := httpClient().Do(req)
	defer res.Body.Close()
	if err != nil {
		return
	}
	if res.StatusCode == 401 {
		err = errors.New("authorization expired, please run `force login`")
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if res.StatusCode/100 != 2 {
		var messages []ForceError
		json.Unmarshal(body, &messages)
		err = errors.New(messages[0].Message)
		return
	}
	return
}

func httpClient() (client *http.Client) {
	chain := rootCertificate()
	config := tls.Config{}
	config.RootCAs = x509.NewCertPool()
	for _, cert := range chain.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			panic(err)
		}
		config.RootCAs.AddCert(x509Cert)
	}
	config.BuildNameToCertificate()
	tr := http.Transport{TLSClientConfig: &config}
	client = &http.Client{Transport: &tr}
	return
}

func httpRequest(method, url string, body io.Reader) (request *http.Request, err error) {
	request, err = http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	request.Header.Add("User-Agent", fmt.Sprintf("force/%s (%s-%s)", Version, runtime.GOOS, runtime.GOARCH))
	return
}

func rootCertificate() (cert tls.Certificate) {
	certPEMBlock := []byte(RootCertificates)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return
}

func startLocalHttpServer(ch chan ForceCredentials) (port int, err error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}
	port = listener.Addr().(*net.TCPAddr).Port
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://force-cli.herokuapp.com")
		query := r.URL.Query()
		var creds ForceCredentials
		creds.AccessToken = query.Get("access_token")
		creds.Id = query.Get("id")
		creds.InstanceUrl = query.Get("instance_url")
		creds.IssuedAt = query.Get("issued_at")
		creds.Scope = query.Get("scope")
		ch <- creds
		listener.Close()
	})
	go http.Serve(listener, h)
	return
}
