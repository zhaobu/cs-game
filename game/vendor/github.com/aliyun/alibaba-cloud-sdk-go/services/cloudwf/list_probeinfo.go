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

// ListProbeinfo invokes the cloudwf.ListProbeinfo API synchronously
// api document: https://help.aliyun.com/api/cloudwf/listprobeinfo.html
func (client *Client) ListProbeinfo(request *ListProbeinfoRequest) (response *ListProbeinfoResponse, err error) {
	response = CreateListProbeinfoResponse()
	err = client.DoAction(request, response)
	return
}

// ListProbeinfoWithChan invokes the cloudwf.ListProbeinfo API asynchronously
// api document: https://help.aliyun.com/api/cloudwf/listprobeinfo.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListProbeinfoWithChan(request *ListProbeinfoRequest) (<-chan *ListProbeinfoResponse, <-chan error) {
	responseChan := make(chan *ListProbeinfoResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListProbeinfo(request)
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

// ListProbeinfoWithCallback invokes the cloudwf.ListProbeinfo API asynchronously
// api document: https://help.aliyun.com/api/cloudwf/listprobeinfo.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListProbeinfoWithCallback(request *ListProbeinfoRequest, callback func(response *ListProbeinfoResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListProbeinfoResponse
		var err error
		defer close(result)
		response, err = client.ListProbeinfo(request)
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

// ListProbeinfoRequest is the request struct for api ListProbeinfo
type ListProbeinfoRequest struct {
	*requests.RpcRequest
	OrderCol         string           `position:"Query" name:"OrderCol"`
	SearchUserMac    string           `position:"Query" name:"SearchUserMac"`
	SearchSensorMac  string           `position:"Query" name:"SearchSensorMac"`
	Length           requests.Integer `position:"Query" name:"Length"`
	SearchSensorName string           `position:"Query" name:"SearchSensorName"`
	PageIndex        requests.Integer `position:"Query" name:"PageIndex"`
	OrderDir         string           `position:"Query" name:"OrderDir"`
}

// ListProbeinfoResponse is the response struct for api ListProbeinfo
type ListProbeinfoResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
	Success   bool   `json:"Success" xml:"Success"`
	Message   string `json:"Message" xml:"Message"`
	Data      string `json:"Data" xml:"Data"`
	ErrorCode int    `json:"ErrorCode" xml:"ErrorCode"`
	ErrorMsg  string `json:"ErrorMsg" xml:"ErrorMsg"`
}

// CreateListProbeinfoRequest creates a request to invoke ListProbeinfo API
func CreateListProbeinfoRequest() (request *ListProbeinfoRequest) {
	request = &ListProbeinfoRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("cloudwf", "2017-03-28", "ListProbeinfo", "cloudwf", "openAPI")
	return
}

// CreateListProbeinfoResponse creates a response to parse from ListProbeinfo response
func CreateListProbeinfoResponse() (response *ListProbeinfoResponse) {
	response = &ListProbeinfoResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
