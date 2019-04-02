package ess

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

// DetachInstances invokes the ess.DetachInstances API synchronously
// api document: https://help.aliyun.com/api/ess/detachinstances.html
func (client *Client) DetachInstances(request *DetachInstancesRequest) (response *DetachInstancesResponse, err error) {
	response = CreateDetachInstancesResponse()
	err = client.DoAction(request, response)
	return
}

// DetachInstancesWithChan invokes the ess.DetachInstances API asynchronously
// api document: https://help.aliyun.com/api/ess/detachinstances.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DetachInstancesWithChan(request *DetachInstancesRequest) (<-chan *DetachInstancesResponse, <-chan error) {
	responseChan := make(chan *DetachInstancesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DetachInstances(request)
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

// DetachInstancesWithCallback invokes the ess.DetachInstances API asynchronously
// api document: https://help.aliyun.com/api/ess/detachinstances.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DetachInstancesWithCallback(request *DetachInstancesRequest, callback func(response *DetachInstancesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DetachInstancesResponse
		var err error
		defer close(result)
		response, err = client.DetachInstances(request)
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

// DetachInstancesRequest is the request struct for api DetachInstances
type DetachInstancesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	InstanceId           *[]string        `position:"Query" name:"InstanceId"  type:"Repeated"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ScalingGroupId       string           `position:"Query" name:"ScalingGroupId"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
}

// DetachInstancesResponse is the response struct for api DetachInstances
type DetachInstancesResponse struct {
	*responses.BaseResponse
	ScalingActivityId string `json:"ScalingActivityId" xml:"ScalingActivityId"`
	RequestId         string `json:"RequestId" xml:"RequestId"`
}

// CreateDetachInstancesRequest creates a request to invoke DetachInstances API
func CreateDetachInstancesRequest() (request *DetachInstancesRequest) {
	request = &DetachInstancesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ess", "2014-08-28", "DetachInstances", "ess", "openAPI")
	return
}

// CreateDetachInstancesResponse creates a response to parse from DetachInstances response
func CreateDetachInstancesResponse() (response *DetachInstancesResponse) {
	response = &DetachInstancesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
