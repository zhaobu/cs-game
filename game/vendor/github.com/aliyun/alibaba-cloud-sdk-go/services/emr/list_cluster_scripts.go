package emr

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

// ListClusterScripts invokes the emr.ListClusterScripts API synchronously
// api document: https://help.aliyun.com/api/emr/listclusterscripts.html
func (client *Client) ListClusterScripts(request *ListClusterScriptsRequest) (response *ListClusterScriptsResponse, err error) {
	response = CreateListClusterScriptsResponse()
	err = client.DoAction(request, response)
	return
}

// ListClusterScriptsWithChan invokes the emr.ListClusterScripts API asynchronously
// api document: https://help.aliyun.com/api/emr/listclusterscripts.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListClusterScriptsWithChan(request *ListClusterScriptsRequest) (<-chan *ListClusterScriptsResponse, <-chan error) {
	responseChan := make(chan *ListClusterScriptsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListClusterScripts(request)
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

// ListClusterScriptsWithCallback invokes the emr.ListClusterScripts API asynchronously
// api document: https://help.aliyun.com/api/emr/listclusterscripts.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ListClusterScriptsWithCallback(request *ListClusterScriptsRequest, callback func(response *ListClusterScriptsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListClusterScriptsResponse
		var err error
		defer close(result)
		response, err = client.ListClusterScripts(request)
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

// ListClusterScriptsRequest is the request struct for api ListClusterScripts
type ListClusterScriptsRequest struct {
	*requests.RpcRequest
	ResourceOwnerId requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ClusterId       string           `position:"Query" name:"ClusterId"`
}

// ListClusterScriptsResponse is the response struct for api ListClusterScripts
type ListClusterScriptsResponse struct {
	*responses.BaseResponse
	RequestId      string         `json:"RequestId" xml:"RequestId"`
	ClusterScripts ClusterScripts `json:"ClusterScripts" xml:"ClusterScripts"`
}

// CreateListClusterScriptsRequest creates a request to invoke ListClusterScripts API
func CreateListClusterScriptsRequest() (request *ListClusterScriptsRequest) {
	request = &ListClusterScriptsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Emr", "2016-04-08", "ListClusterScripts", "emr", "openAPI")
	return
}

// CreateListClusterScriptsResponse creates a response to parse from ListClusterScripts response
func CreateListClusterScriptsResponse() (response *ListClusterScriptsResponse) {
	response = &ListClusterScriptsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
