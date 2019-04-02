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

// VoiceCancelScan invokes the green.VoiceCancelScan API synchronously
// api document: https://help.aliyun.com/api/green/voicecancelscan.html
func (client *Client) VoiceCancelScan(request *VoiceCancelScanRequest) (response *VoiceCancelScanResponse, err error) {
	response = CreateVoiceCancelScanResponse()
	err = client.DoAction(request, response)
	return
}

// VoiceCancelScanWithChan invokes the green.VoiceCancelScan API asynchronously
// api document: https://help.aliyun.com/api/green/voicecancelscan.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) VoiceCancelScanWithChan(request *VoiceCancelScanRequest) (<-chan *VoiceCancelScanResponse, <-chan error) {
	responseChan := make(chan *VoiceCancelScanResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.VoiceCancelScan(request)
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

// VoiceCancelScanWithCallback invokes the green.VoiceCancelScan API asynchronously
// api document: https://help.aliyun.com/api/green/voicecancelscan.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) VoiceCancelScanWithCallback(request *VoiceCancelScanRequest, callback func(response *VoiceCancelScanResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *VoiceCancelScanResponse
		var err error
		defer close(result)
		response, err = client.VoiceCancelScan(request)
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

// VoiceCancelScanRequest is the request struct for api VoiceCancelScan
type VoiceCancelScanRequest struct {
	*requests.RoaRequest
	ClientInfo string `position:"Query" name:"ClientInfo"`
}

// VoiceCancelScanResponse is the response struct for api VoiceCancelScan
type VoiceCancelScanResponse struct {
	*responses.BaseResponse
}

// CreateVoiceCancelScanRequest creates a request to invoke VoiceCancelScan API
func CreateVoiceCancelScanRequest() (request *VoiceCancelScanRequest) {
	request = &VoiceCancelScanRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("Green", "2018-05-09", "VoiceCancelScan", "/green/voice/cancelscan", "green", "openAPI")
	request.Method = requests.POST
	return
}

// CreateVoiceCancelScanResponse creates a response to parse from VoiceCancelScan response
func CreateVoiceCancelScanResponse() (response *VoiceCancelScanResponse) {
	response = &VoiceCancelScanResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
