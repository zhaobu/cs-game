package green

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

// DeleteImageLib invokes the green.DeleteImageLib API synchronously
// api document: https://help.aliyun.com/api/green/deleteimagelib.html
func (client *Client) DeleteImageLib(request *DeleteImageLibRequest) (response *DeleteImageLibResponse, err error) {
	response = CreateDeleteImageLibResponse()
	err = client.DoAction(request, response)
	return
}

// DeleteImageLibWithChan invokes the green.DeleteImageLib API asynchronously
// api document: https://help.aliyun.com/api/green/deleteimagelib.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteImageLibWithChan(request *DeleteImageLibRequest) (<-chan *DeleteImageLibResponse, <-chan error) {
	responseChan := make(chan *DeleteImageLibResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteImageLib(request)
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

// DeleteImageLibWithCallback invokes the green.DeleteImageLib API asynchronously
// api document: https://help.aliyun.com/api/green/deleteimagelib.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteImageLibWithCallback(request *DeleteImageLibRequest, callback func(response *DeleteImageLibResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteImageLibResponse
		var err error
		defer close(result)
		response, err = client.DeleteImageLib(request)
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

// DeleteImageLibRequest is the request struct for api DeleteImageLib
type DeleteImageLibRequest struct {
	*requests.RpcRequest
	SourceIp string           `position:"Query" name:"SourceIp"`
	Id       requests.Integer `position:"Query" name:"Id"`
}

// DeleteImageLibResponse is the response struct for api DeleteImageLib
type DeleteImageLibResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeleteImageLibRequest creates a request to invoke DeleteImageLib API
func CreateDeleteImageLibRequest() (request *DeleteImageLibRequest) {
	request = &DeleteImageLibRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Green", "2017-08-23", "DeleteImageLib", "green", "openAPI")
	return
}

// CreateDeleteImageLibResponse creates a response to parse from DeleteImageLib response
func CreateDeleteImageLibResponse() (response *DeleteImageLibResponse) {
	response = &DeleteImageLibResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
