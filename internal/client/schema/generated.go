// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package schema

import (
	"context"
	"encoding/json"

	"github.com/Khan/genqlient/graphql"
)

// Deployment includes the GraphQL fields of DagsterCloudDeployment requested by the fragment Deployment.
type Deployment struct {
	DeploymentName string `json:"deploymentName"`
	DeploymentId   int    `json:"deploymentId"`
}

// GetDeploymentName returns Deployment.DeploymentName, and is useful for accessing the field via an interface.
func (v *Deployment) GetDeploymentName() string { return v.DeploymentName }

// GetDeploymentId returns Deployment.DeploymentId, and is useful for accessing the field via an interface.
func (v *Deployment) GetDeploymentId() int { return v.DeploymentId }

// GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment includes the requested fields of the GraphQL type DagsterCloudDeployment.
type GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment struct {
	Deployment `json:"-"`
}

// GetDeploymentName returns GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment.DeploymentName, and is useful for accessing the field via an interface.
func (v *GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment) GetDeploymentName() string {
	return v.Deployment.DeploymentName
}

// GetDeploymentId returns GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment.DeploymentId, and is useful for accessing the field via an interface.
func (v *GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment) GetDeploymentId() int {
	return v.Deployment.DeploymentId
}

func (v *GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment) UnmarshalJSON(b []byte) error {

	if string(b) == "null" {
		return nil
	}

	var firstPass struct {
		*GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment
		graphql.NoUnmarshalJSON
	}
	firstPass.GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment = v

	err := json.Unmarshal(b, &firstPass)
	if err != nil {
		return err
	}

	err = json.Unmarshal(
		b, &v.Deployment)
	if err != nil {
		return err
	}
	return nil
}

type __premarshalGetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment struct {
	DeploymentName string `json:"deploymentName"`

	DeploymentId int `json:"deploymentId"`
}

func (v *GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment) MarshalJSON() ([]byte, error) {
	premarshaled, err := v.__premarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(premarshaled)
}

func (v *GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment) __premarshalJSON() (*__premarshalGetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment, error) {
	var retval __premarshalGetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment

	retval.DeploymentName = v.Deployment.DeploymentName
	retval.DeploymentId = v.Deployment.DeploymentId
	return &retval, nil
}

// GetCurrentDeploymentResponse is returned by GetCurrentDeployment on success.
type GetCurrentDeploymentResponse struct {
	CurrentDeployment GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment `json:"currentDeployment"`
}

// GetCurrentDeployment returns GetCurrentDeploymentResponse.CurrentDeployment, and is useful for accessing the field via an interface.
func (v *GetCurrentDeploymentResponse) GetCurrentDeployment() GetCurrentDeploymentCurrentDeploymentDagsterCloudDeployment {
	return v.CurrentDeployment
}

// The query or mutation executed by GetCurrentDeployment.
const GetCurrentDeployment_Operation = `
query GetCurrentDeployment {
	currentDeployment {
		... Deployment
	}
}
fragment Deployment on DagsterCloudDeployment {
	deploymentName
	deploymentId
}
`

func GetCurrentDeployment(
	ctx_ context.Context,
	client_ graphql.Client,
) (*GetCurrentDeploymentResponse, error) {
	req_ := &graphql.Request{
		OpName: "GetCurrentDeployment",
		Query:  GetCurrentDeployment_Operation,
	}
	var err_ error

	var data_ GetCurrentDeploymentResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}
