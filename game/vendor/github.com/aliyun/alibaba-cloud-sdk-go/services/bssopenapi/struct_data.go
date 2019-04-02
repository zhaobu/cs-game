package bssopenapi

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

// Data is a nested struct in bssopenapi response
type Data struct {
	BusinessType               string                                 `json:"BusinessType" xml:"BusinessType"`
	TradePrice                 float64                                `json:"TradePrice" xml:"TradePrice"`
	HostId                     string                                 `json:"HostId" xml:"HostId"`
	OriginalPrice              float64                                `json:"OriginalPrice" xml:"OriginalPrice"`
	OrderId                    string                                 `json:"OrderId" xml:"OrderId"`
	TotalCount                 int                                    `json:"TotalCount" xml:"TotalCount"`
	BillingCycle               string                                 `json:"BillingCycle" xml:"BillingCycle"`
	Uid                        int                                    `json:"Uid" xml:"Uid"`
	OutstandingAmount          float64                                `json:"OutstandingAmount" xml:"OutstandingAmount"`
	InvalidTimeStamp           int                                    `json:"InvalidTimeStamp" xml:"InvalidTimeStamp"`
	Quantity                   int                                    `json:"Quantity" xml:"Quantity"`
	AvailableCashAmount        string                                 `json:"AvailableCashAmount" xml:"AvailableCashAmount"`
	EffectTimeStamp            int                                    `json:"EffectTimeStamp" xml:"EffectTimeStamp"`
	PrimaryAccount             string                                 `json:"PrimaryAccount" xml:"PrimaryAccount"`
	HostName                   string                                 `json:"HostName" xml:"HostName"`
	TotalOutstandingAmount     float64                                `json:"TotalOutstandingAmount" xml:"TotalOutstandingAmount"`
	Status                     string                                 `json:"Status" xml:"Status"`
	UserId                     int                                    `json:"UserId" xml:"UserId"`
	NewInvoiceAmount           float64                                `json:"NewInvoiceAmount" xml:"NewInvoiceAmount"`
	Numerator                  int                                    `json:"Numerator" xml:"Numerator"`
	AvailableAmount            string                                 `json:"AvailableAmount" xml:"AvailableAmount"`
	PageSize                   int                                    `json:"PageSize" xml:"PageSize"`
	Amount                     string                                 `json:"Amount" xml:"Amount"`
	MybankCreditAmount         string                                 `json:"MybankCreditAmount" xml:"MybankCreditAmount"`
	CreditAmount               string                                 `json:"CreditAmount" xml:"CreditAmount"`
	ThresholdType              int                                    `json:"ThresholdType" xml:"ThresholdType"`
	AccountID                  string                                 `json:"AccountID" xml:"AccountID"`
	InstanceId                 string                                 `json:"InstanceId" xml:"InstanceId"`
	ItemCode                   string                                 `json:"ItemCode" xml:"ItemCode"`
	ThresholdAmount            string                                 `json:"ThresholdAmount" xml:"ThresholdAmount"`
	InvoiceApplyId             int                                    `json:"InvoiceApplyId" xml:"InvoiceApplyId"`
	Boolean                    bool                                   `json:"Boolean" xml:"Boolean"`
	PageNum                    int                                    `json:"PageNum" xml:"PageNum"`
	Bid                        string                                 `json:"Bid" xml:"Bid"`
	Currency                   string                                 `json:"Currency" xml:"Currency"`
	DiscountPrice              float64                                `json:"DiscountPrice" xml:"DiscountPrice"`
	AccountName                string                                 `json:"AccountName" xml:"AccountName"`
	Denominator                int                                    `json:"Denominator" xml:"Denominator"`
	ModuleList                 ModuleList                             `json:"ModuleList" xml:"ModuleList"`
	InstanceList               []Instance                             `json:"InstanceList" xml:"InstanceList"`
	OrderList                  OrderListInQueryOrders                 `json:"OrderList" xml:"OrderList"`
	Modules                    ModulesInQueryInstanceGaapCost         `json:"Modules" xml:"Modules"`
	Items                      ItemsInQueryInstanceBill               `json:"Items" xml:"Items"`
	ResourcePackages           ResourcePackages                       `json:"ResourcePackages" xml:"ResourcePackages"`
	ProductList                ProductList                            `json:"ProductList" xml:"ProductList"`
	ModuleDetails              ModuleDetailsInGetSubscriptionPrice    `json:"ModuleDetails" xml:"ModuleDetails"`
	PromotionDetails           PromotionDetailsInGetSubscriptionPrice `json:"PromotionDetails" xml:"PromotionDetails"`
	CustomerInvoiceList        CustomerInvoiceList                    `json:"CustomerInvoiceList" xml:"CustomerInvoiceList"`
	EvaluateList               EvaluateList                           `json:"EvaluateList" xml:"EvaluateList"`
	Promotions                 Promotions                             `json:"Promotions" xml:"Promotions"`
	CustomerInvoiceAddressList CustomerInvoiceAddressList             `json:"CustomerInvoiceAddressList" xml:"CustomerInvoiceAddressList"`
	AttributeList              AttributeList                          `json:"AttributeList" xml:"AttributeList"`
}
