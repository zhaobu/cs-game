package dds

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

// VpcItem is a nested struct in dds response
type VpcItem struct {
	VpcId       string    `json:"VpcId" xml:"VpcId"`
	VpcName     string    `json:"VpcName" xml:"VpcName"`
	Bid         string    `json:"Bid" xml:"Bid"`
	AliUid      string    `json:"AliUid" xml:"AliUid"`
	RegionNo    string    `json:"RegionNo" xml:"RegionNo"`
	CidrBlock   string    `json:"CidrBlock" xml:"CidrBlock"`
	IsDefault   bool      `json:"IsDefault" xml:"IsDefault"`
	Status      string    `json:"Status" xml:"Status"`
	GmtCreate   string    `json:"GmtCreate" xml:"GmtCreate"`
	GmtModified string    `json:"GmtModified" xml:"GmtModified"`
	VSwitchs    []VSwitch `json:"VSwitchs" xml:"VSwitchs"`
}
