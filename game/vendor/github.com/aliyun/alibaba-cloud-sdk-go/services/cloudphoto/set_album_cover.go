package cloudphoto

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

// SetAlbumCover invokes the cloudphoto.SetAlbumCover API synchronously
// api document: https://help.aliyun.com/api/cloudphoto/setalbumcover.html
func (client *Client) SetAlbumCover(request *SetAlbumCoverRequest) (response *SetAlbumCoverResponse, err error) {
	response = CreateSetAlbumCoverResponse()
	err = client.DoAction(request, response)
	return
}

// SetAlbumCoverWithChan invokes the cloudphoto.SetAlbumCover API asynchronously
// api document: https://help.aliyun.com/api/cloudphoto/setalbumcover.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SetAlbumCoverWithChan(request *SetAlbumCoverRequest) (<-chan *SetAlbumCoverResponse, <-chan error) {
	responseChan := make(chan *SetAlbumCoverResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.SetAlbumCover(request)
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

// SetAlbumCoverWithCallback invokes the cloudphoto.SetAlbumCover API asynchronously
// api document: https://help.aliyun.com/api/cloudphoto/setalbumcover.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) SetAlbumCoverWithCallback(request *SetAlbumCoverRequest, callback func(response *SetAlbumCoverResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *SetAlbumCoverResponse
		var err error
		defer close(result)
		response, err = client.SetAlbumCover(request)
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

// SetAlbumCoverRequest is the request struct for api SetAlbumCover
type SetAlbumCoverRequest struct {
	*requests.RpcRequest
	LibraryId string           `position:"Query" name:"LibraryId"`
	AlbumId   requests.Integer `position:"Query" name:"AlbumId"`
	PhotoId   requests.Integer `position:"Query" name:"PhotoId"`
	StoreName string           `position:"Query" name:"StoreName"`
}

// SetAlbumCoverResponse is the response struct for api SetAlbumCover
type SetAlbumCoverResponse struct {
	*responses.BaseResponse
	Code      string `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	RequestId string `json:"RequestId" xml:"RequestId"`
	Action    string `json:"Action" xml:"Action"`
}

// CreateSetAlbumCoverRequest creates a request to invoke SetAlbumCover API
func CreateSetAlbumCoverRequest() (request *SetAlbumCoverRequest) {
	request = &SetAlbumCoverRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("CloudPhoto", "2017-07-11", "SetAlbumCover", "cloudphoto", "openAPI")
	return
}

// CreateSetAlbumCoverResponse creates a response to parse from SetAlbumCover response
func CreateSetAlbumCoverResponse() (response *SetAlbumCoverResponse) {
	response = &SetAlbumCoverResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
