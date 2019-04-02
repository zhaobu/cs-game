package baas

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

// UpgradeChaincode invokes the baas.UpgradeChaincode API synchronously
// api document: https://help.aliyun.com/api/baas/upgradechaincode.html
func (client *Client) UpgradeChaincode(request *UpgradeChaincodeRequest) (response *UpgradeChaincodeResponse, err error) {
	response = CreateUpgradeChaincodeResponse()
	err = client.DoAction(request, response)
	return
}

// UpgradeChaincodeWithChan invokes the baas.UpgradeChaincode API asynchronously
// api document: https://help.aliyun.com/api/baas/upgradechaincode.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpgradeChaincodeWithChan(request *UpgradeChaincodeRequest) (<-chan *UpgradeChaincodeResponse, <-chan error) {
	responseChan := make(chan *UpgradeChaincodeResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.UpgradeChaincode(request)
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

// UpgradeChaincodeWithCallback invokes the baas.UpgradeChaincode API asynchronously
// api document: https://help.aliyun.com/api/baas/upgradechaincode.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) UpgradeChaincodeWithCallback(request *UpgradeChaincodeRequest, callback func(response *UpgradeChaincodeResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *UpgradeChaincodeResponse
		var err error
		defer close(result)
		response, err = client.UpgradeChaincode(request)
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

// UpgradeChaincodeRequest is the request struct for api UpgradeChaincode
type UpgradeChaincodeRequest struct {
	*requests.RpcRequest
	OrganizationId string `position:"Body" name:"OrganizationId"`
	ChaincodeId    string `position:"Body" name:"ChaincodeId"`
	EndorsePolicy  string `position:"Body" name:"EndorsePolicy"`
	Location       string `position:"Body" name:"Location"`
}

// UpgradeChaincodeResponse is the response struct for api UpgradeChaincode
type UpgradeChaincodeResponse struct {
	*responses.BaseResponse
	RequestId string                   `json:"RequestId" xml:"RequestId"`
	Success   bool                     `json:"Success" xml:"Success"`
	ErrorCode int                      `json:"ErrorCode" xml:"ErrorCode"`
	Result    ResultInUpgradeChaincode `json:"Result" xml:"Result"`
}

// CreateUpgradeChaincodeRequest creates a request to invoke UpgradeChaincode API
func CreateUpgradeChaincodeRequest() (request *UpgradeChaincodeRequest) {
	request = &UpgradeChaincodeRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Baas", "2018-07-31", "UpgradeChaincode", "", "")
	return
}

// CreateUpgradeChaincodeResponse creates a response to parse from UpgradeChaincode response
func CreateUpgradeChaincodeResponse() (response *UpgradeChaincodeResponse) {
	response = &UpgradeChaincodeResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
