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

// UpdateProjectSetting invokes the emr.UpdateProjectSetting API synchronously
// api document: https://help.aliyun.com/api/emr/updateprojectsetting.html
func (client *Client) UpdateProjectSetting(request *UpdateProjectSettingRequest) (response *UpdateProjectSettingResponse, err error) {
	response = CreateUpdateProjectSettingResponse()
	err = client.DoAction(request, response)
	return
}

// UpdateProjectSettingWithChan invokes the emr.UpdateProjectSetting API asynchronously
// api document: https://help.aliyun.com/api/emr/updateprojectsetting.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpdateProjectSettingWithChan(request *UpdateProjectSettingRequest) (<-chan *UpdateProjectSettingResponse, <-chan error) {
	responseChan := make(chan *UpdateProjectSettingResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.UpdateProjectSetting(request)
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

// UpdateProjectSettingWithCallback invokes the emr.UpdateProjectSetting API asynchronously
// api document: https://help.aliyun.com/api/emr/updateprojectsetting.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpdateProjectSettingWithCallback(request *UpdateProjectSettingRequest, callback func(response *UpdateProjectSettingResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *UpdateProjectSettingResponse
		var err error
		defer close(result)
		response, err = client.UpdateProjectSetting(request)
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

// UpdateProjectSettingRequest is the request struct for api UpdateProjectSetting
type UpdateProjectSettingRequest struct {
	*requests.RpcRequest
	ResourceOwnerId requests.Integer `position:"Query" name:"ResourceOwnerId"`
	DefaultOssPath  string           `position:"Query" name:"DefaultOssPath"`
	ProjectId       string           `position:"Query" name:"ProjectId"`
	OssConfig       string           `position:"Query" name:"OssConfig"`
}

// UpdateProjectSettingResponse is the response struct for api UpdateProjectSetting
type UpdateProjectSettingResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateUpdateProjectSettingRequest creates a request to invoke UpdateProjectSetting API
func CreateUpdateProjectSettingRequest() (request *UpdateProjectSettingRequest) {
	request = &UpdateProjectSettingRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Emr", "2016-04-08", "UpdateProjectSetting", "emr", "openAPI")
	return
}

// CreateUpdateProjectSettingResponse creates a response to parse from UpdateProjectSetting response
func CreateUpdateProjectSettingResponse() (response *UpdateProjectSettingResponse) {
	response = &UpdateProjectSettingResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
