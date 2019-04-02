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

package push

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// QueryDevicesByAlias invokes the push.QueryDevicesByAlias API synchronously
// api document: https://help.aliyun.com/api/push/querydevicesbyalias.html
func (client *Client) QueryDevicesByAlias(request *QueryDevicesByAliasRequest) (response *QueryDevicesByAliasResponse, err error) {
	response = CreateQueryDevicesByAliasResponse()
	err = client.DoAction(request, response)
	return
}

// QueryDevicesByAliasWithChan invokes the push.QueryDevicesByAlias API asynchronously
// api document: https://help.aliyun.com/api/push/querydevicesbyalias.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryDevicesByAliasWithChan(request *QueryDevicesByAliasRequest) (<-chan *QueryDevicesByAliasResponse, <-chan error) {
	responseChan := make(chan *QueryDevicesByAliasResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.QueryDevicesByAlias(request)
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

// QueryDevicesByAliasWithCallback invokes the push.QueryDevicesByAlias API asynchronously
// api document: https://help.aliyun.com/api/push/querydevicesbyalias.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryDevicesByAliasWithCallback(request *QueryDevicesByAliasRequest, callback func(response *QueryDevicesByAliasResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *QueryDevicesByAliasResponse
		var err error
		defer close(result)
		response, err = client.QueryDevicesByAlias(request)
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

// QueryDevicesByAliasRequest is the request struct for api QueryDevicesByAlias
type QueryDevicesByAliasRequest struct {
	*requests.RpcRequest
	AccessKeyId string           `position:"Query" name:"AccessKeyId"`
	AppKey      requests.Integer `position:"Query" name:"AppKey"`
	Alias       string           `position:"Query" name:"Alias"`
}

// QueryDevicesByAliasResponse is the response struct for api QueryDevicesByAlias
type QueryDevicesByAliasResponse struct {
	*responses.BaseResponse
	RequestId string   `json:"RequestId" xml:"RequestId"`
	DeviceIds []string `json:"DeviceIds" xml:"DeviceIds"`
}

// CreateQueryDevicesByAliasRequest creates a request to invoke QueryDevicesByAlias API
func CreateQueryDevicesByAliasRequest() (request *QueryDevicesByAliasRequest) {
	request = &QueryDevicesByAliasRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Push", "2016-08-01", "QueryDevicesByAlias", "push", "openAPI")
	return
}

// CreateQueryDevicesByAliasResponse creates a response to parse from QueryDevicesByAlias response
func CreateQueryDevicesByAliasResponse() (response *QueryDevicesByAliasResponse) {
	response = &QueryDevicesByAliasResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
