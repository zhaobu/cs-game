package gpdb

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

// DeleteDBInstance invokes the gpdb.DeleteDBInstance API synchronously
// api document: https://help.aliyun.com/api/gpdb/deletedbinstance.html
func (client *Client) DeleteDBInstance(request *DeleteDBInstanceRequest) (response *DeleteDBInstanceResponse, err error) {
	response = CreateDeleteDBInstanceResponse()
	err = client.DoAction(request, response)
	return
}

// DeleteDBInstanceWithChan invokes the gpdb.DeleteDBInstance API asynchronously
// api document: https://help.aliyun.com/api/gpdb/deletedbinstance.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteDBInstanceWithChan(request *DeleteDBInstanceRequest) (<-chan *DeleteDBInstanceResponse, <-chan error) {
	responseChan := make(chan *DeleteDBInstanceResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteDBInstance(request)
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

// DeleteDBInstanceWithCallback invokes the gpdb.DeleteDBInstance API asynchronously
// api document: https://help.aliyun.com/api/gpdb/deletedbinstance.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteDBInstanceWithCallback(request *DeleteDBInstanceRequest, callback func(response *DeleteDBInstanceResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteDBInstanceResponse
		var err error
		defer close(result)
		response, err = client.DeleteDBInstance(request)
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

// DeleteDBInstanceRequest is the request struct for api DeleteDBInstance
type DeleteDBInstanceRequest struct {
	*requests.RpcRequest
	ClientToken  string           `position:"Query" name:"ClientToken"`
	DBInstanceId string           `position:"Query" name:"DBInstanceId"`
	OwnerId      requests.Integer `position:"Query" name:"OwnerId"`
}

// DeleteDBInstanceResponse is the response struct for api DeleteDBInstance
type DeleteDBInstanceResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeleteDBInstanceRequest creates a request to invoke DeleteDBInstance API
func CreateDeleteDBInstanceRequest() (request *DeleteDBInstanceRequest) {
	request = &DeleteDBInstanceRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("gpdb", "2016-05-03", "DeleteDBInstance", "gpdb", "openAPI")
	return
}

// CreateDeleteDBInstanceResponse creates a response to parse from DeleteDBInstance response
func CreateDeleteDBInstanceResponse() (response *DeleteDBInstanceResponse) {
	response = &DeleteDBInstanceResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
