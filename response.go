package firstCentral

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ResponseObject struct {
	XMLName     xml.Name     `xml:"Envelope"`
	XmlnsXsi    string       `xml:"xmlns:xsi,attr"`
	XmlnsXsd    string       `xml:"xmlns:xsd,attr"`
	XmlnsSoap12 string       `xml:"xmlns:soap,attr"`
	Body        ResponseBody `xml:"Body"`
}

func (r *ResponseObject) IsLoginResponse() bool {
	return r.Body.LoginResponse != nil
}
func (r *ResponseObject) GetLoginResponse() *LoginResponse {
	return r.Body.LoginResponse
}
func (r *ResponseObject) GetConsumerMatchResult() (cmr *ConsumerMatching, err error) {
	err = xml.Unmarshal([]byte(r.Body.ConnectConsumerMatchResponse.ConnectConsumerMatchResult), &cmr)
	if err != nil {
		return
	}
	if len(cmr.MatchedConsumer) < 2 && cmr.MatchedConsumer[0].ConsumerID == "0" {
		err = fmt.Errorf("no matches found")
	}
	return
}
func (r *ResponseObject) GetConsumerFullCreditResponse() (cmr *ConsumerFullCredit, err error) {
	err = xml.Unmarshal([]byte(r.Body.GetConsumerFullCreditReportResponse.GetConsumerFullCreditReportResult), &cmr)
	if err != nil {
		return
	}
	if cmr.CreditAgreementSummary == nil || len(cmr.CreditAgreementSummary) == 0 {
		err = fmt.Errorf("no credit agreements found")
	}
	return
}

type ResponseBody struct {
	XMLName                             xml.Name                            `xml:"Body"`
	LoginResponse                       *LoginResponse                      `xml:"LoginResponse,omitempty"`
	ConnectConsumerMatchResponse        ConnectConsumerMatchResponse        `xml:"ConnectConsumerMatchResponse,omitempty"`
	GetConsumerFullCreditReportResponse GetConsumerFullCreditReportResponse `xml:"GetConsumerFullCreditReportResponse,omitempty"`
}
type LoginResponse struct {
	XMLName     xml.Name `xml:"LoginResponse"`
	Xmlns       string   `xml:"xmlns,attr"`
	LoginResult string   `xml:"LoginResult"`
}
type ConnectConsumerMatchResponse struct {
	XMLName                    xml.Name `xml:"ConnectConsumerMatchResponse"`
	Xmlns                      string   `xml:"xmlns,attr"`
	ConnectConsumerMatchResult string   `xml:"ConnectConsumerMatchResult"`
}
type ConsumerMatching struct {
	XMLName         xml.Name          `xml:"ConsumerMtaching"`
	MatchedConsumer []MatchedConsumer `xml:"MatchedConsumer"`
}
type MatchedConsumer struct {
	XMLName          xml.Name `xml:"MatchedConsumer"`
	MatchingEngineID string   `xml:"MatchingEngineID"`
	EnquiryID        string   `xml:"EnquiryID"`
	ConsumerID       string   `xml:"ConsumerID"`
	Reference        string   `xml:"Reference"`
	MatchingRate     string   `xml:"MatchingRate"`
	FirstName        string   `xml:"FirstName"`
	Surname          string   `xml:"Surname"`
	OtherNames       string   `xml:"OtherNames"`
	AccountNo        string   `xml:"AccountNo"`
}
type GetConsumerFullCreditReportResponse struct {
	XMLName                           xml.Name `xml:"GetConsumerFullCreditReportResponse"`
	Xmlns                             string   `xml:"xmlns,attr"`
	GetConsumerFullCreditReportResult string   `xml:"GetConsumerFullCreditReportResult"`
}
type ConsumerFullCredit struct {
	XMLName                xml.Name                 `xml:"ConsumerFullCredit"`
	CreditAgreementSummary []CreditAgreementSummary `xml:"CreditAgreementSummary"`
	SubjectList            []SubjectList            `xml:"SubjectList"`
}
type CreditAgreementSummary struct {
	XMLName              xml.Name `xml:"CreditAgreementSummary"`
	DateAccountOpened    string   `xml:"DateAccountOpened"`
	SubscriberName       string   `xml:"SubscriberName"`
	AccountNo            string   `xml:"AccountNo"`
	SubAccountNo         string   `xml:"SubAccountNo"`
	IndicatorDescription string   `xml:"IndicatorDescription"`
	OpeningBalanceAmt    string   `xml:"OpeningBalanceAmt"`
	Currency             string   `xml:"Currency"`
	CurrentBalanceAmt    string   `xml:"CurrentBalanceAmt"`
	InstalmentAmount     string   `xml:"InstalmentAmount"`
	AmountOverdue        string   `xml:"AmountOverdue"`
	ClosedDate           string   `xml:"ClosedDate"`
	LoanDuration         string   `xml:"LoanDuration"`
	RepaymentFrequency   string   `xml:"RepaymentFrequency"`
	LastUpdatedDate      string   `xml:"LastUpdatedDate"`
	PerformanceStatus    string   `xml:"PerformanceStatus"`
	AccountStatus        string   `xml:"AccountStatus"`
}
type SubjectList struct {
	XMLName      xml.Name `xml:"SubjectList"`
	ConsumerID   string   `xml:"ConsumerID"`
	SearchOutput string   `xml:"SearchOutput"`
	Reference    string   `xml:"Reference"`
}

func (cfc *ConsumerFullCredit) GetCleanReport(bvn string) (cr *CleanedReport) {

	if cfc.SubjectList == nil || len(cfc.SubjectList) == 0 {
		return &CleanedReport{
			BVN:     bvn,
			NoHit:   true,
			ID:      "",
			Records: nil,
		}
	}

	return &CleanedReport{
		BVN:     bvn,
		NoHit:   false,
		ID:      cfc.SubjectList[0].Reference,
		Records: cfc.GetCleanRecords(),
	}
}
func (cfc *ConsumerFullCredit) GetCleanRecords() (crs []Record) {

	if cfc.CreditAgreementSummary == nil || len(cfc.CreditAgreementSummary) == 0 {
		return []Record{}
	}

	for _, facility := range cfc.CreditAgreementSummary {
		if crs == nil {
			crs = []Record{}
		}
		crs = append(crs, facility.GetCleanRecord())
	}
	return
}
func (l *CreditAgreementSummary) GetCleanRecord() Record {
	return Record{
		Institution:             l.SubscriberName,
		Amount:                  l.OpeningBalanceAmt,
		Status:                  l.AccountStatus,
		Balance:                 l.CurrentBalanceAmt,
		AmountOverdue:           l.AmountOverdue,
		Classification:          l.PerformanceStatus,
		DisbursalDate:           l.DateAccountOpened,
		MaturityDate:            l.ClosedDate,
		Source:                  "xds",
		ReportDate:              l.LastUpdatedDate,
		RefreshedOn:             time.Now().Format("02-Jan-2006"),
		BureauIdentifierEntry:   l.AccountNo,
		BureauIdentifierAccount: l.AccountNo,
	}
}

func (l *LoginResponse) GetTicket() string {
	return l.LoginResult
}

func processMatchedCustomersForFullCreditReport(mcs []MatchedConsumer, threshold int) (cId, mergeList, enqEngID, enqID string, err error) {
	ml := []string{}
	for _, mc := range mcs {
		var mr int
		mr, err = strconv.Atoi(mc.MatchingRate)
		if err != nil {
			return
		}
		if mr < threshold {
			continue
		}

		ml = append(ml, mc.ConsumerID)
	}

	if len(ml) == 0 {
		err = fmt.Errorf("no matches met the threshold")
		return
	}

	hc := getMatchWithHighestConfidence(mcs)
	cId = hc.ConsumerID
	enqEngID = hc.MatchingEngineID
	enqID = hc.EnquiryID
	mergeList = strings.Join(ml, ",")

	return
}
func getMatchWithHighestConfidence(mcs []MatchedConsumer) (mc MatchedConsumer) {
	hcs := 0
	for _, y := range mcs {
		cs, err := strconv.Atoi(y.MatchingRate)
		if err != nil {
			continue
		}
		if cs >= hcs {
			mc = y
			hcs = cs
		}
	}
	return
}
