package cms

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

// CreateMyGroupAlertBatch invokes the cms.CreateMyGroupAlertBatch API synchronously
// api document: https://help.aliyun.com/api/cms/createmygroupalertbatch.html
func (client *Client) CreateMyGroupAlertBatch(request *CreateMyGroupAlertBatchRequest) (response *CreateMyGroupAlertBatchResponse, err error) {
	response = CreateCreateMyGroupAlertBatchResponse()
	err = client.DoAction(request, response)
	return
}

// CreateMyGroupAlertBatchWithChan invokes the cms.CreateMyGroupAlertBatch API asynchronously
// api document: https://help.aliyun.com/api/cms/createmygroupalertbatch.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateMyGroupAlertBatchWithChan(request *CreateMyGroupAlertBatchRequest) (<-chan *CreateMyGroupAlertBatchResponse, <-chan error) {
	responseChan := make(chan *CreateMyGroupAlertBatchResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.CreateMyGroupAlertBatch(request)
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

// CreateMyGroupAlertBatchWithCallback invokes the cms.CreateMyGroupAlertBatch API asynchronously
// api document: https://help.aliyun.com/api/cms/createmygroupalertbatch.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) CreateMyGroupAlertBatchWithCallback(request *CreateMyGroupAlertBatchRequest, callback func(response *CreateMyGroupAlertBatchResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *CreateMyGroupAlertBatchResponse
		var err error
		defer close(result)
		response, err = client.CreateMyGroupAlertBatch(request)
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

// CreateMyGroupAlertBatchRequest is the request struct for api CreateMyGroupAlertBatch
type CreateMyGroupAlertBatchRequest struct {
	*requests.RpcRequest
	GroupId             requests.Integer `position:"Query" name:"GroupId"`
	GroupAlertJsonArray string           `position:"Query" name:"GroupAlertJsonArray"`
}

// CreateMyGroupAlertBatchResponse is the response struct for api CreateMyGroupAlertBatch
type CreateMyGroupAlertBatchResponse struct {
	*responses.BaseResponse
	RequestId    string                             `json:"RequestId" xml:"RequestId"`
	Success      bool                               `json:"Success" xml:"Success"`
	ErrorCode    int                                `json:"ErrorCode" xml:"ErrorCode"`
	ErrorMessage string                             `json:"ErrorMessage" xml:"ErrorMessage"`
	Resources    ResourcesInCreateMyGroupAlertBatch `json:"Resources" xml:"Resources"`
}

// CreateCreateMyGroupAlertBatchRequest creates a request to invoke CreateMyGroupAlertBatch API
func CreateCreateMyGroupAlertBatchRequest() (request *CreateMyGroupAlertBatchRequest) {
	request = &CreateMyGroupAlertBatchRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Cms", "2018-03-08", "CreateMyGroupAlertBatch", "cms", "openAPI")
	return
}

// CreateCreateMyGroupAlertBatchResponse creates a response to parse from CreateMyGroupAlertBatch response
func CreateCreateMyGroupAlertBatchResponse() (response *CreateMyGroupAlertBatchResponse) {
	response = &CreateMyGroupAlertBatchResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
