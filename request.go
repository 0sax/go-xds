package firstCentral

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestObject struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	XmlnsXsi string `xml:"xmlns:xsi,attr"`
	XmlnsXsd  string `xml:"xmlns:xsd,attr"`
	XmlnsSoap string `xml:"xmlns:soap,attr"`
	Body      *Body  `xml:"soap:Body,omitempty"`
}
type Body struct {
	XMLName xml.Name `xml:"soap:Body"`
	Login *Login `xml:"Login,omitempty"`
	ConnectConsumerMatch *ConnectConsumerMatch          `xml:"ConnectConsumerMatch,omitempty"`
	ConsumerFullCreditReport *ConsumerFullCreditRequest `xml:"GetConsumerFullCreditReport,omitempty"`
}
type Login struct {
	XMLName xml.Name `xml:"Login"`
	Xmlns string `xml:"xmlns,attr"`
	Username string  `xml:"UserName"`
	Password string `xml:"Password"`
}
type ConnectConsumerMatch struct {
	XMLName xml.Name `xml:"ConnectConsumerMatch"`
	Xmlns string `xml:"xmlns,attr"`
	DataTicket string `xml:"DataTicket"`
	EnquiryReason string `xml:"EnquiryReason"`
	ConsumerName string `xml:"ConsumerName"`
	DateOfBirth string `xml:"DateOfBirth"`
	Identification string `xml:"Identification"`
	AccountNumber string `xml:"AccountNumber"`
	ProductID string `xml:"ProductID"`
}
type ConsumerFullCreditRequest struct {
	XMLName xml.Name `xml:"GetConsumerFullCreditReport"`
	Xmlns string `xml:"xmlns,attr"`
	DataTicket string `xml:"DataTicket"`
	ConsumerID string `xml:"ConsumerID"`
	ConsumerMergeList string `xml:"consumerMergeList"`
	SubscriberEnquiryEngineID string `xml:"SubscriberEnquiryEngineID"`
	EnquiryID string `xml:"enquiryID"`
}


func (s *Service) makeRequest(method, actionParam string, headers map[string]interface{}, body interface{}) (ro *ResponseObject, err error) {

	//convert request to XML
	b, err := xml.Marshal(body)
	if err != nil {
		fmt.Printf("error at point 2.2: %v\n", err) //debug delete
		return
	}

	ss := xml.Header + string(b)
	req, err := http.NewRequest(method, s.BaseUrl, bytes.NewBuffer([]byte(ss)))
	if err != nil {
		fmt.Printf("error at point 2: %v\n", err) //debug delete
		return
	}

	// Add headers to request
	actionUrl := "https://online.firstcentralcreditbureau.com/FirstCentralNigeriaWebService/"
	req.Header.Add("SOAPAction", actionUrl+actionParam)
	req.Header.Add("Content-Type", "text/xml")
	for k, v := range headers {
		req.Header.Set(k, v.(string))
	}


	resp, err := s.Client.Do(req)
	if err != nil {
		fmt.Printf("error at point 3: %v\n", err) //debug delete
		return
	}
	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error at point 4: %v\n", err) //debug delete
		return
	}


	err = xml.Unmarshal(bdy, &ro)
	if err != nil {
		fmt.Printf("error at point 6a: %v\n", err) //debug delete
		return
	}


	return

}
