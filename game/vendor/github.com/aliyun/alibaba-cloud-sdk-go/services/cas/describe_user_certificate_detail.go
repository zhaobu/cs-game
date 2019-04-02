package cas

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

// DescribeUserCertificateDetail invokes the cas.DescribeUserCertificateDetail API synchronously
// api document: https://help.aliyun.com/api/cas/describeusercertificatedetail.html
func (client *Client) DescribeUserCertificateDetail(request *DescribeUserCertificateDetailRequest) (response *DescribeUserCertificateDetailResponse, err error) {
	response = CreateDescribeUserCertificateDetailResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeUserCertificateDetailWithChan invokes the cas.DescribeUserCertificateDetail API asynchronously
// api document: https://help.aliyun.com/api/cas/describeusercertificatedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeUserCertificateDetailWithChan(request *DescribeUserCertificateDetailRequest) (<-chan *DescribeUserCertificateDetailResponse, <-chan error) {
	responseChan := make(chan *DescribeUserCertificateDetailResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeUserCertificateDetail(request)
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

// DescribeUserCertificateDetailWithCallback invokes the cas.DescribeUserCertificateDetail API asynchronously
// api document: https://help.aliyun.com/api/cas/describeusercertificatedetail.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeUserCertificateDetailWithCallback(request *DescribeUserCertificateDetailRequest, callback func(response *DescribeUserCertificateDetailResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeUserCertificateDetailResponse
		var err error
		defer close(result)
		response, err = client.DescribeUserCertificateDetail(request)
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

// DescribeUserCertificateDetailRequest is the request struct for api DescribeUserCertificateDetail
type DescribeUserCertificateDetailRequest struct {
	*requests.RpcRequest
	SourceIp string           `position:"Query" name:"SourceIp"`
	CertId   requests.Integer `position:"Query" name:"CertId"`
	Lang     string           `position:"Query" name:"Lang"`
}

// DescribeUserCertificateDetailResponse is the response struct for api DescribeUserCertificateDetail
type DescribeUserCertificateDetailResponse struct {
	*responses.BaseResponse
	RequestId   string `json:"RequestId" xml:"RequestId"`
	Id          int    `json:"Id" xml:"Id"`
	Name        string `json:"Name" xml:"Name"`
	Common      string `json:"Common" xml:"Common"`
	Fingerprint string `json:"Fingerprint" xml:"Fingerprint"`
	Issuer      string `json:"Issuer" xml:"Issuer"`
	OrgName     string `json:"OrgName" xml:"OrgName"`
	Province    string `json:"Province" xml:"Province"`
	City        string `json:"City" xml:"City"`
	Country     string `json:"Country" xml:"Country"`
	StartDate   string `json:"StartDate" xml:"StartDate"`
	EndDate     string `json:"EndDate" xml:"EndDate"`
	Sans        string `json:"Sans" xml:"Sans"`
	Expired     bool   `json:"Expired" xml:"Expired"`
	BuyInAliyun bool   `json:"BuyInAliyun" xml:"BuyInAliyun"`
	Cert        string `json:"Cert" xml:"Cert"`
	Key         string `json:"Key" xml:"Key"`
}

// CreateDescribeUserCertificateDetailRequest creates a request to invoke DescribeUserCertificateDetail API
func CreateDescribeUserCertificateDetailRequest() (request *DescribeUserCertificateDetailRequest) {
	request = &DescribeUserCertificateDetailRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("cas", "2018-07-13", "DescribeUserCertificateDetail", "cas", "openAPI")
	return
}

// CreateDescribeUserCertificateDetailResponse creates a response to parse from DescribeUserCertificateDetail response
func CreateDescribeUserCertificateDetailResponse() (response *DescribeUserCertificateDetailResponse) {
	response = &DescribeUserCertificateDetailResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
