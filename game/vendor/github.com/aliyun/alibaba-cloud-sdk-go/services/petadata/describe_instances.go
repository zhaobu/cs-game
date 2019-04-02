package petadata

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DescribeInstances invokes the petadata.DescribeInstances API synchronously
// api document: https://help.aliyun.com/api/petadata/describeinstances.html
func (client *Client) DescribeInstances(request *DescribeInstancesRequest) (response *DescribeInstancesResponse, err error) {
	response = CreateDescribeInstancesResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeInstancesWithChan invokes the petadata.DescribeInstances API asynchronously
// api document: https://help.aliyun.com/api/petadata/describeinstances.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeInstancesWithChan(request *DescribeInstancesRequest) (<-chan *DescribeInstancesResponse, <-chan error) {
	responseChan := make(chan *DescribeInstancesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeInstances(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DescribeInstancesWithCallback invokes the petadata.DescribeInstances API asynchronously
// api document: https://help.aliyun.com/api/petadata/describeinstances.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeInstancesWithCallback(request *DescribeInstancesRequest, callback func(response *DescribeInstancesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeInstancesResponse
		var err error
		defer close(result)
		response, err = client.DescribeInstances(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DescribeInstancesRequest is the request struct for api DescribeInstances
type DescribeInstancesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	InstanceStatus       string           `position:"Query" name:"InstanceStatus"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	InstanceId           string           `position:"Query" name:"InstanceId"`
	SecurityToken        string           `position:"Query" name:"SecurityToken"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	ChargeType           string           `position:"Query" name:"ChargeType"`
}

// DescribeInstancesResponse is the response struct for api DescribeInstances
type DescribeInstancesResponse struct {
	*responses.BaseResponse
	RequestId  string    `json:"RequestId" xml:"RequestId"`
	PageNumber int       `json:"PageNumber" xml:"PageNumber"`
	PageSize   int       `json:"PageSize" xml:"PageSize"`
	TotalCount int       `json:"TotalCount" xml:"TotalCount"`
	Instances  Instances `json:"Instances" xml:"Instances"`
}

// CreateDescribeInstancesRequest creates a request to invoke DescribeInstances API
func CreateDescribeInstancesRequest() (request *DescribeInstancesRequest) {
	request = &DescribeInstancesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("PetaData", "2016-01-01", "DescribeInstances", "petadata", "openAPI")
	return
}

// CreateDescribeInstancesResponse creates a response to parse from DescribeInstances response
func CreateDescribeInstancesResponse() (response *DescribeInstancesResponse) {
	response = &DescribeInstancesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
