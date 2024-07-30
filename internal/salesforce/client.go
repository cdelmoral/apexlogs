package salesforce

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const DateTimeLayout = "2006-01-02T15:04:05.999Z0700"

type Attributes struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type QueryResponse[T any] struct {
	QueryLocator   string `json:"queryLocator"`
	EntityTypeName string `json:"entityTypeName"`
	Records        []T    `json:"records"`
	Size           int    `json:"size"`
	TotalSize      int    `json:"totalSize"`
	Done           bool   `json:"done"`
}

type PostSObjectResponse struct {
	Id       string
	Errors   []string
	Warnings []string
	Infos    []string
	Success  bool
}

// A Client is a Salesforce API client.
// Client stores the access token, instance URL, API version and alias of a Salesforce org.
type Client struct {
	accessToken string
	instanceUrl string
	apiVersion  string
	alias       string
}

// NewClient creates a new Client.
// NewClient receives a [ScratchOrgInfo] with the Salesforce org information.
// It returns a pointer to a new [Client].
func NewClient(orgInfo ScratchOrgInfo) *Client {
	return &Client{
		accessToken: orgInfo.AccessToken,
		instanceUrl: orgInfo.InstanceUrl,
		apiVersion:  orgInfo.ApiVersion,
		alias:       orgInfo.Alias,
	}
}

func (c *Client) doRequest(
	method, resource, body string,
	queryParams map[string]string,
	headers map[string]string,
) ([]byte, error) {
	u, err := url.Parse(c.instanceUrl)
	if err != nil {
		return nil, fmt.Errorf("unexpected error parsing instance url")
	}

	path := fmt.Sprintf("/services/data/v%s/tooling/%s", c.apiVersion, resource)
	u.Path = path

	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{}

	req, err := http.NewRequest(method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating http request: %s", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %s", err)
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return resBody, fmt.Errorf("error reading response body: %s", err)
	}

	if res.StatusCode > 399 {
		return resBody, fmt.Errorf("request returned error code: %s", res.Status)
	}

	return resBody, nil
}

func (c *Client) doQuery(query string, v any) error {
	resource := "query"
	q := map[string]string{
		"q": query,
	}

	body, err := c.doRequest("GET", resource, "", q, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}

	return nil
}

// DoQuery performs a query request to the Salesforce API.
// The response is typed to the type T, which represents the Salesforce object related to the query.
// An error is returned if the requests fails or if the response cannot be unmarshalled to type T.
func DoQuery[T any](c *Client, query string) (QueryResponse[T], error) {
	resource := "query"
	q := map[string]string{
		"q": query,
	}

	body, err := c.doRequest("GET", resource, "", q, nil)
	if err != nil {
		return QueryResponse[T]{}, err
	}

	var res QueryResponse[T]
	err = json.Unmarshal(body, &res)
	if err != nil {
		return QueryResponse[T]{}, err
	}

	return res, nil
}

// PatchSObject performs an update request to the Salesforce API.
// An error is returned if the payload cannot be serialized or if the request fails.
func PatchSObject(c *Client, resource, id string, payload any) error {
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializing payload: %s", err)
	}

	r := fmt.Sprintf("sobjects/%s/%s", resource, id)
	h := map[string]string{"Content-Type": "application/json"}
	_, err = c.doRequest("PATCH", r, string(serializedPayload), nil, h)
	if err != nil {
		return fmt.Errorf("error sending request to update record: %s", err)
	}

	return nil
}

// PostSObject permforms a create requests to the Salesforce API.
// An error is returned in the following cases:
//   - The payload cannot be serialized
//   - The request fails
//   - The response cannot be deserialized
//   - The response contains an error (the response is also returned in this case)
func PostSObject(c *Client, resource string, payload any) (PostSObjectResponse, error) {
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return PostSObjectResponse{}, fmt.Errorf("error serializing payload: %s", err)
	}

	r := fmt.Sprintf("sobjects/%s", resource)
	h := map[string]string{"Content-Type": "application/json"}
	body, err := c.doRequest("POST", r, string(serializedPayload), nil, h)
	if err != nil {
		return PostSObjectResponse{}, fmt.Errorf("error sending request to create new record: %s", err)
	}

	var unserializedBody PostSObjectResponse
	err = json.Unmarshal(body, &unserializedBody)
	if err != nil {
		return PostSObjectResponse{}, fmt.Errorf("unexpected error parsing response body: %s", err)
	}

	if !unserializedBody.Success {
		return unserializedBody, fmt.Errorf("error creating new record: %s", unserializedBody.Errors)
	}

	return unserializedBody, nil
}

// GetSObjectBody performs a request to retrieve the body of an Object of type resource with the given id.
// An error is returned if the request fails.
func GetSObjectBody(c *Client, resource, id string) (string, error) {
	r := fmt.Sprintf("sobjects/%s/%s/Body", resource, id)

	body, err := c.doRequest("GET", r, "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("error sending request to retrieve record body: %s", err)
	}

	return string(body), nil
}
