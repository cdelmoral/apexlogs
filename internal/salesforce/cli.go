package salesforce

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// A ScratchOrgInfo contains the necessary information to connect to a Salesforce org.
type ScratchOrgInfo struct {
	AccessToken string
	InstanceUrl string
	ApiVersion  string
	Alias       string
}

// UserInfo contains the Salesforce user information.
type UserInfo struct {
	AccessToken string
	Id          string
	InstanceUrl string
	LoginUrl    string
	OrgId       string
	ProfileName string
	Username    string
	Alias       string
}

// A CommandResponse represents the result of an Salesforce CLI command.
type CommandResponse[T any] struct {
	Result   T
	Warnings []string
	Status   int
}

// GetDefaultUserInfo returns the Salesforce CLI default user.
func GetDefaultUserInfo() (UserInfo, error) {
	cmd := exec.Command("sf", "org", "display", "user", "--json")
	out, err := cmd.Output()
	if err != nil {
		return UserInfo{}, err
	}

	var userInfoResponse CommandResponse[UserInfo]
	err = json.Unmarshal(out, &userInfoResponse)
	if err != nil {
		return UserInfo{}, err
	}

	return userInfoResponse.Result, nil
}

// GetDefaultScratchOrgInfo returns the Salesforce CLI default scratch org.
func GetDefaultScratchOrgInfo() (ScratchOrgInfo, error) {
	var info ScratchOrgInfo

	cmd := exec.Command("sf", "org", "display", "--json")
	out, err := cmd.Output()
	if err != nil {
		return info, err
	}

	var orgDisplay map[string]any
	err = json.Unmarshal(out, &orgDisplay)
	if err != nil {
		return info, err
	}

	orgDisplayResult, ok := orgDisplay["result"].(map[string]any)
	if !ok {
		return info, fmt.Errorf("unexpected error parsing org display result")
	}

	info.AccessToken, _ = orgDisplayResult["accessToken"].(string)
	info.InstanceUrl, _ = orgDisplayResult["instanceUrl"].(string)
	info.ApiVersion, _ = orgDisplayResult["apiVersion"].(string)
	info.Alias, _ = orgDisplayResult["alias"].(string)

	return info, nil
}
