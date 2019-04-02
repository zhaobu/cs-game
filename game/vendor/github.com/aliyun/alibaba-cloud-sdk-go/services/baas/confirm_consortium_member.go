package baas

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

// ConfirmConsortiumMember invokes the baas.ConfirmConsortiumMember API synchronously
// api document: https://help.aliyun.com/api/baas/confirmconsortiummember.html
func (client *Client) ConfirmConsortiumMember(request *ConfirmConsortiumMemberRequest) (response *ConfirmConsortiumMemberResponse, err error) {
	response = CreateConfirmConsortiumMemberResponse()
	err = client.DoAction(request, response)
	return
}

// ConfirmConsortiumMemberWithChan invokes the baas.ConfirmConsortiumMember API asynchronously
// api document: https://help.aliyun.com/api/baas/confirmconsortiummember.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ConfirmConsortiumMemberWithChan(request *ConfirmConsortiumMemberRequest) (<-chan *ConfirmConsortiumMemberResponse, <-chan error) {
	responseChan := make(chan *ConfirmConsortiumMemberResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ConfirmConsortiumMember(request)
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

// ConfirmConsortiumMemberWithCallback invokes the baas.ConfirmConsortiumMember API asynchronously
// api document: https://help.aliyun.com/api/baas/confirmconsortiummember.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ConfirmConsortiumMemberWithCallback(request *ConfirmConsortiumMemberRequest, callback func(response *ConfirmConsortiumMemberResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ConfirmConsortiumMemberResponse
		var err error
		defer close(result)
		response, err = client.ConfirmConsortiumMember(request)
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

// ConfirmConsortiumMemberRequest is the request struct for api ConfirmConsortiumMember
type ConfirmConsortiumMemberRequest struct {
	*requests.RpcRequest
	Organization *[]ConfirmConsortiumMemberOrganization `position:"Query" name:"Organization"  type:"Repeated"`
	ConsortiumId string                                 `position:"Query" name:"ConsortiumId"`
}

// ConfirmConsortiumMemberOrganization is a repeated param struct in ConfirmConsortiumMemberRequest
type ConfirmConsortiumMemberOrganization struct {
	Id string `name:"Id"`
}

// ConfirmConsortiumMemberResponse is the response struct for api ConfirmConsortiumMember
type ConfirmConsortiumMemberResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Success   bool   `json:"Success" xml:"Success"`
	ErrorCode int    `json:"ErrorCode" xml:"ErrorCode"`
	Result    bool   `json:"Result" xml:"Result"`
}

// CreateConfirmConsortiumMemberRequest creates a request to invoke ConfirmConsortiumMember API
func CreateConfirmConsortiumMemberRequest() (request *ConfirmConsortiumMemberRequest) {
	request = &ConfirmConsortiumMemberRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Baas", "2018-07-31", "ConfirmConsortiumMember", "", "")
	return
}

// CreateConfirmConsortiumMemberResponse creates a response to parse from ConfirmConsortiumMember response
func CreateConfirmConsortiumMemberResponse() (response *ConfirmConsortiumMemberResponse) {
	response = &ConfirmConsortiumMemberResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
