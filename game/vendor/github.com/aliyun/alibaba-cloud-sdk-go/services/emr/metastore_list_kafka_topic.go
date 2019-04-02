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

// MetastoreListKafkaTopic invokes the emr.MetastoreListKafkaTopic API synchronously
// api document: https://help.aliyun.com/api/emr/metastorelistkafkatopic.html
func (client *Client) MetastoreListKafkaTopic(request *MetastoreListKafkaTopicRequest) (response *MetastoreListKafkaTopicResponse, err error) {
	response = CreateMetastoreListKafkaTopicResponse()
	err = client.DoAction(request, response)
	return
}

// MetastoreListKafkaTopicWithChan invokes the emr.MetastoreListKafkaTopic API asynchronously
// api document: https://help.aliyun.com/api/emr/metastorelistkafkatopic.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) MetastoreListKafkaTopicWithChan(request *MetastoreListKafkaTopicRequest) (<-chan *MetastoreListKafkaTopicResponse, <-chan error) {
	responseChan := make(chan *MetastoreListKafkaTopicResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.MetastoreListKafkaTopic(request)
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

// MetastoreListKafkaTopicWithCallback invokes the emr.MetastoreListKafkaTopic API asynchronously
// api document: https://help.aliyun.com/api/emr/metastorelistkafkatopic.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) MetastoreListKafkaTopicWithCallback(request *MetastoreListKafkaTopicRequest, callback func(response *MetastoreListKafkaTopicResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *MetastoreListKafkaTopicResponse
		var err error
		defer close(result)
		response, err = client.MetastoreListKafkaTopic(request)
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

// MetastoreListKafkaTopicRequest is the request struct for api MetastoreListKafkaTopic
type MetastoreListKafkaTopicRequest struct {
	*requests.RpcRequest
	ResourceOwnerId requests.Integer `position:"Query" name:"ResourceOwnerId"`
	PageSize        requests.Integer `position:"Query" name:"PageSize"`
	DataSourceId    string           `position:"Query" name:"DataSourceId"`
	TopicName       string           `position:"Query" name:"TopicName"`
	ClusterId       string           `position:"Query" name:"ClusterId"`
	PageNumber      requests.Integer `position:"Query" name:"PageNumber"`
}

// MetastoreListKafkaTopicResponse is the response struct for api MetastoreListKafkaTopic
type MetastoreListKafkaTopicResponse struct {
	*responses.BaseResponse
	RequestId  string    `json:"RequestId" xml:"RequestId"`
	TotalCount int       `json:"TotalCount" xml:"TotalCount"`
	PageNumber int       `json:"PageNumber" xml:"PageNumber"`
	PageSize   int       `json:"PageSize" xml:"PageSize"`
	TopicList  TopicList `json:"TopicList" xml:"TopicList"`
}

// CreateMetastoreListKafkaTopicRequest creates a request to invoke MetastoreListKafkaTopic API
func CreateMetastoreListKafkaTopicRequest() (request *MetastoreListKafkaTopicRequest) {
	request = &MetastoreListKafkaTopicRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Emr", "2016-04-08", "MetastoreListKafkaTopic", "emr", "openAPI")
	return
}

// CreateMetastoreListKafkaTopicResponse creates a response to parse from MetastoreListKafkaTopic response
func CreateMetastoreListKafkaTopicResponse() (response *MetastoreListKafkaTopicResponse) {
	response = &MetastoreListKafkaTopicResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
