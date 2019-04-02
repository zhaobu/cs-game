package cdn

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

// SetDomainGreenManagerConfig invokes the cdn.SetDomainGreenManagerConfig API synchronously
// api document: https://help.aliyun.com/api/cdn/setdomaingreenmanagerconfig.html
func (client *Client) SetDomainGreenManagerConfig(request *SetDomainGreenManagerConfigRequest) (response *SetDomainGreenManagerConfigResponse, err error) {
	response = CreateSetDomainGreenManagerConfigResponse()
	err = client.DoAction(request, response)
	return
}

// SetDomainGreenManagerConfigWithChan invokes the cdn.SetDomainGreenManagerConfig API asynchronously
// api document: https://help.aliyun.com/api/cdn/setdomaingreenmanagerconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SetDomainGreenManagerConfigWithChan(request *SetDomainGreenManagerConfigRequest) (<-chan *SetDomainGreenManagerConfigResponse, <-chan error) {
	responseChan := make(chan *SetDomainGreenManagerConfigResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.SetDomainGreenManagerConfig(request)
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

// SetDomainGreenManagerConfigWithCallback invokes the cdn.SetDomainGreenManagerConfig API asynchronously
// api document: https://help.aliyun.com/api/cdn/setdomaingreenmanagerconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SetDomainGreenManagerConfigWithCallback(request *SetDomainGreenManagerConfigRequest, callback func(response *SetDomainGreenManagerConfigResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *SetDomainGreenManagerConfigResponse
		var err error
		defer close(result)
		response, err = client.SetDomainGreenManagerConfig(request)
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

// SetDomainGreenManagerConfigRequest is the request struct for api SetDomainGreenManagerConfig
type SetDomainGreenManagerConfigRequest struct {
	*requests.RpcRequest
	Enable     string           `position:"Query" name:"Enable"`
	DomainName string           `position:"Query" name:"DomainName"`
	OwnerId    requests.Integer `position:"Query" name:"OwnerId"`
}

// SetDomainGreenManagerConfigResponse is the response struct for api SetDomainGreenManagerConfig
type SetDomainGreenManagerConfigResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateSetDomainGreenManagerConfigRequest creates a request to invoke SetDomainGreenManagerConfig API
func CreateSetDomainGreenManagerConfigRequest() (request *SetDomainGreenManagerConfigRequest) {
	request = &SetDomainGreenManagerConfigRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Cdn", "2018-05-10", "SetDomainGreenManagerConfig", "", "")
	return
}

// CreateSetDomainGreenManagerConfigResponse creates a response to parse from SetDomainGreenManagerConfig response
func CreateSetDomainGreenManagerConfigResponse() (response *SetDomainGreenManagerConfigResponse) {
	response = &SetDomainGreenManagerConfigResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
