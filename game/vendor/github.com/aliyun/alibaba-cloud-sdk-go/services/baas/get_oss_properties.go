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

// GetOssProperties invokes the baas.GetOssProperties API synchronously
// api document: https://help.aliyun.com/api/baas/getossproperties.html
func (client *Client) GetOssProperties(request *GetOssPropertiesRequest) (response *GetOssPropertiesResponse, err error) {
	response = CreateGetOssPropertiesResponse()
	err = client.DoAction(request, response)
	return
}

// GetOssPropertiesWithChan invokes the baas.GetOssProperties API asynchronously
// api document: https://help.aliyun.com/api/baas/getossproperties.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetOssPropertiesWithChan(request *GetOssPropertiesRequest) (<-chan *GetOssPropertiesResponse, <-chan error) {
	responseChan := make(chan *GetOssPropertiesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetOssProperties(request)
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

// GetOssPropertiesWithCallback invokes the baas.GetOssProperties API asynchronously
// api document: https://help.aliyun.com/api/baas/getossproperties.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetOssPropertiesWithCallback(request *GetOssPropertiesRequest, callback func(response *GetOssPropertiesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetOssPropertiesResponse
		var err error
		defer close(result)
		response, err = client.GetOssProperties(request)
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

// GetOssPropertiesRequest is the request struct for api GetOssProperties
type GetOssPropertiesRequest struct {
	*requests.RpcRequest
	Bizid string `position:"Body" name:"Bizid"`
}

// GetOssPropertiesResponse is the response struct for api GetOssProperties
type GetOssPropertiesResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Result    Result `json:"Result" xml:"Result"`
}

// CreateGetOssPropertiesRequest creates a request to invoke GetOssProperties API
func CreateGetOssPropertiesRequest() (request *GetOssPropertiesRequest) {
	request = &GetOssPropertiesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Baas", "2018-07-31", "GetOssProperties", "", "")
	return
}

// CreateGetOssPropertiesResponse creates a response to parse from GetOssProperties response
func CreateGetOssPropertiesResponse() (response *GetOssPropertiesResponse) {
	response = &GetOssPropertiesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
