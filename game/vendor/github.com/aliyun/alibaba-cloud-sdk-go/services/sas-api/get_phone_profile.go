package sas_api

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

// GetPhoneProfile invokes the sas_api.GetPhoneProfile API synchronously
// api document: https://help.aliyun.com/api/sas-api/getphoneprofile.html
func (client *Client) GetPhoneProfile(request *GetPhoneProfileRequest) (response *GetPhoneProfileResponse, err error) {
	response = CreateGetPhoneProfileResponse()
	err = client.DoAction(request, response)
	return
}

// GetPhoneProfileWithChan invokes the sas_api.GetPhoneProfile API asynchronously
// api document: https://help.aliyun.com/api/sas-api/getphoneprofile.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetPhoneProfileWithChan(request *GetPhoneProfileRequest) (<-chan *GetPhoneProfileResponse, <-chan error) {
	responseChan := make(chan *GetPhoneProfileResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetPhoneProfile(request)
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

// GetPhoneProfileWithCallback invokes the sas_api.GetPhoneProfile API asynchronously
// api document: https://help.aliyun.com/api/sas-api/getphoneprofile.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetPhoneProfileWithCallback(request *GetPhoneProfileRequest, callback func(response *GetPhoneProfileResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetPhoneProfileResponse
		var err error
		defer close(result)
		response, err = client.GetPhoneProfile(request)
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

// GetPhoneProfileRequest is the request struct for api GetPhoneProfile
type GetPhoneProfileRequest struct {
	*requests.RpcRequest
	Phone        string           `position:"Query" name:"Phone"`
	SensType     requests.Integer `position:"Query" name:"SensType"`
	DataVersion  string           `position:"Query" name:"DataVersion"`
	BusinessType requests.Integer `position:"Query" name:"BusinessType"`
}

// GetPhoneProfileResponse is the response struct for api GetPhoneProfile
type GetPhoneProfileResponse struct {
	*responses.BaseResponse
	Code      int    `json:"Code" xml:"Code"`
	Message   string `json:"Message" xml:"Message"`
	Success   bool   `json:"Success" xml:"Success"`
	RequestId string `json:"RequestId" xml:"RequestId"`
	Data      Data   `json:"Data" xml:"Data"`
}

// CreateGetPhoneProfileRequest creates a request to invoke GetPhoneProfile API
func CreateGetPhoneProfileRequest() (request *GetPhoneProfileRequest) {
	request = &GetPhoneProfileRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Sas-api", "2017-07-05", "GetPhoneProfile", "sas-api", "openAPI")
	return
}

// CreateGetPhoneProfileResponse creates a response to parse from GetPhoneProfile response
func CreateGetPhoneProfileResponse() (response *GetPhoneProfileResponse) {
	response = &GetPhoneProfileResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}