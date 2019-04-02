package ros

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

// DescribeResourceDetail invokes the ros.DescribeResourceDetail API synchronously
// api document: https://help.aliyun.com/api/ros/describeresourcedetail.html
func (client *Client) DescribeResourceDetail(request *DescribeResourceDetailRequest) (response *DescribeResourceDetailResponse, err error) {
	response = CreateDescribeResourceDetailResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeResourceDetailWithChan invokes the ros.DescribeResourceDetail API asynchronously
// api document: https://help.aliyun.com/api/ros/describeresourcedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeResourceDetailWithChan(request *DescribeResourceDetailRequest) (<-chan *DescribeResourceDetailResponse, <-chan error) {
	responseChan := make(chan *DescribeResourceDetailResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeResourceDetail(request)
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

// DescribeResourceDetailWithCallback invokes the ros.DescribeResourceDetail API asynchronously
// api document: https://help.aliyun.com/api/ros/describeresourcedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeResourceDetailWithCallback(request *DescribeResourceDetailRequest, callback func(response *DescribeResourceDetailResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeResourceDetailResponse
		var err error
		defer close(result)
		response, err = client.DescribeResourceDetail(request)
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

// DescribeResourceDetailRequest is the request struct for api DescribeResourceDetail
type DescribeResourceDetailRequest struct {
	*requests.RoaRequest
	StackId      string `position:"Path" name:"StackId"`
	StackName    string `position:"Path" name:"StackName"`
	ResourceName string `position:"Path" name:"ResourceName"`
}

// DescribeResourceDetailResponse is the response struct for api DescribeResourceDetail
type DescribeResourceDetailResponse struct {
	*responses.BaseResponse
}

// CreateDescribeResourceDetailRequest creates a request to invoke DescribeResourceDetail API
func CreateDescribeResourceDetailRequest() (request *DescribeResourceDetailRequest) {
	request = &DescribeResourceDetailRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("ROS", "2015-09-01", "DescribeResourceDetail", "/stacks/[StackName]/[StackId]/resources/[ResourceName]", "", "")
	request.Method = requests.GET
	return
}

// CreateDescribeResourceDetailResponse creates a response to parse from DescribeResourceDetail response
func CreateDescribeResourceDetailResponse() (response *DescribeResourceDetailResponse) {
	response = &DescribeResourceDetailResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
