package cloudwf

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

// GetAccountConfig invokes the cloudwf.GetAccountConfig API synchronously
// api document: https://help.aliyun.com/api/cloudwf/getaccountconfig.html
func (client *Client) GetAccountConfig(request *GetAccountConfigRequest) (response *GetAccountConfigResponse, err error) {
	response = CreateGetAccountConfigResponse()
	err = client.DoAction(request, response)
	return
}

// GetAccountConfigWithChan invokes the cloudwf.GetAccountConfig API asynchronously
// api document: https://help.aliyun.com/api/cloudwf/getaccountconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetAccountConfigWithChan(request *GetAccountConfigRequest) (<-chan *GetAccountConfigResponse, <-chan error) {
	responseChan := make(chan *GetAccountConfigResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetAccountConfig(request)
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

// GetAccountConfigWithCallback invokes the cloudwf.GetAccountConfig API asynchronously
// api document: https://help.aliyun.com/api/cloudwf/getaccountconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetAccountConfigWithCallback(request *GetAccountConfigRequest, callback func(response *GetAccountConfigResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetAccountConfigResponse
		var err error
		defer close(result)
		response, err = client.GetAccountConfig(request)
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

// GetAccountConfigRequest is the request struct for api GetAccountConfig
type GetAccountConfigRequest struct {
	*requests.RpcRequest
	Id requests.Integer `position:"Query" name:"Id"`
}

// GetAccountConfigResponse is the response struct for api GetAccountConfig
type GetAccountConfigResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Success   bool   `json:"Success" xml:"Success"`
	Message   string `json:"Message" xml:"Message"`
	Data      string `json:"Data" xml:"Data"`
	ErrorCode int    `json:"ErrorCode" xml:"ErrorCode"`
	ErrorMsg  string `json:"ErrorMsg" xml:"ErrorMsg"`
}

// CreateGetAccountConfigRequest creates a request to invoke GetAccountConfig API
func CreateGetAccountConfigRequest() (request *GetAccountConfigRequest) {
	request = &GetAccountConfigRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("cloudwf", "2017-03-28", "GetAccountConfig", "cloudwf", "openAPI")
	return
}

// CreateGetAccountConfigResponse creates a response to parse from GetAccountConfig response
func CreateGetAccountConfigResponse() (response *GetAccountConfigResponse) {
	response = &GetAccountConfigResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
