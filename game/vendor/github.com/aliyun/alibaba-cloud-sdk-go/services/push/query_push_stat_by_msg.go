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

package push

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// QueryPushStatByMsg invokes the push.QueryPushStatByMsg API synchronously
// api document: https://help.aliyun.com/api/push/querypushstatbymsg.html
func (client *Client) QueryPushStatByMsg(request *QueryPushStatByMsgRequest) (response *QueryPushStatByMsgResponse, err error) {
	response = CreateQueryPushStatByMsgResponse()
	err = client.DoAction(request, response)
	return
}

// QueryPushStatByMsgWithChan invokes the push.QueryPushStatByMsg API asynchronously
// api document: https://help.aliyun.com/api/push/querypushstatbymsg.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryPushStatByMsgWithChan(request *QueryPushStatByMsgRequest) (<-chan *QueryPushStatByMsgResponse, <-chan error) {
	responseChan := make(chan *QueryPushStatByMsgResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.QueryPushStatByMsg(request)
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

// QueryPushStatByMsgWithCallback invokes the push.QueryPushStatByMsg API asynchronously
// api document: https://help.aliyun.com/api/push/querypushstatbymsg.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryPushStatByMsgWithCallback(request *QueryPushStatByMsgRequest, callback func(response *QueryPushStatByMsgResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *QueryPushStatByMsgResponse
		var err error
		defer close(result)
		response, err = client.QueryPushStatByMsg(request)
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

// QueryPushStatByMsgRequest is the request struct for api QueryPushStatByMsg
type QueryPushStatByMsgRequest struct {
	*requests.RpcRequest
	AccessKeyId string           `position:"Query" name:"AccessKeyId"`
	AppKey      requests.Integer `position:"Query" name:"AppKey"`
	MessageId   requests.Integer `position:"Query" name:"MessageId"`
}

// QueryPushStatByMsgResponse is the response struct for api QueryPushStatByMsg
type QueryPushStatByMsgResponse struct {
	*responses.BaseResponse
	RequestId string                       `json:"RequestId" xml:"RequestId"`
	PushStats QueryPushStatByMsgPushStats0 `json:"PushStats" xml:"PushStats"`
}

type QueryPushStatByMsgPushStats0 struct {
	PushStat []QueryPushStatByMsgPushStat1 `json:"PushStat" xml:"PushStat"`
}

type QueryPushStatByMsgPushStat1 struct {
	MessageId              string `json:"MessageId" xml:"MessageId"`
	AcceptCount            int64  `json:"AcceptCount" xml:"AcceptCount"`
	SentCount              int64  `json:"SentCount" xml:"SentCount"`
	ReceivedCount          int64  `json:"ReceivedCount" xml:"ReceivedCount"`
	OpenedCount            int64  `json:"OpenedCount" xml:"OpenedCount"`
	DeletedCount           int64  `json:"DeletedCount" xml:"DeletedCount"`
	SmsSentCount           int64  `json:"SmsSentCount" xml:"SmsSentCount"`
	SmsSkipCount           int64  `json:"SmsSkipCount" xml:"SmsSkipCount"`
	SmsFailedCount         int64  `json:"SmsFailedCount" xml:"SmsFailedCount"`
	SmsReceiveSuccessCount int64  `json:"SmsReceiveSuccessCount" xml:"SmsReceiveSuccessCount"`
	SmsReceiveFailedCount  int64  `json:"SmsReceiveFailedCount" xml:"SmsReceiveFailedCount"`
}

// CreateQueryPushStatByMsgRequest creates a request to invoke QueryPushStatByMsg API
func CreateQueryPushStatByMsgRequest() (request *QueryPushStatByMsgRequest) {
	request = &QueryPushStatByMsgRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Push", "2016-08-01", "QueryPushStatByMsg", "push", "openAPI")
	return
}

// CreateQueryPushStatByMsgResponse creates a response to parse from QueryPushStatByMsg response
func CreateQueryPushStatByMsgResponse() (response *QueryPushStatByMsgResponse) {
	response = &QueryPushStatByMsgResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
