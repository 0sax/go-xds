package firstCentral

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

const (
	//Actions
	LoginAction                       = "Login"
	ConnectConsumerMatchAction        = "ConnectConsumerMatch"
	GetConsumerFullCreditReportAction = "GetConsumerFullCreditReport"
)

func NewFirstCentralService(userName, password string) (s *Service, err error) {
	s = &Service{
		BaseUrl:  "https://online.firstcentralcreditbureau.com/FirstCentralNigeriaWebService/FirstCentralNigeriaWebService.asmx",
		UserName: userName,
		Password: password,
		Client:   http.Client{},
	}
	err = s.Login()
	return
}

type Service struct {
	BaseUrl  string
	UserName string
	Password string
	Client   http.Client
	Ticket   string
}

func (s *Service) Login() (err error) {
	// make login request
	r := RequestObject{
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: &Body{
			Login: &Login{
				Xmlns:    "https://online.firstcentralcreditbureau.com/FirstCentralNigeriaWebService",
				Username: s.UserName,
				Password: s.Password,
			},
		},
	}
	// set service ticket
	var resp *ResponseObject
	resp, err = s.makeRequest(http.MethodPost, LoginAction, nil, r)
	if err != nil {
		return
	}

	if resp.IsLoginResponse() {
		s.Ticket = resp.GetLoginResponse().GetTicket()
	}

	if s.Ticket == "UserNotFound" {
		err = fmt.Errorf("user not found")
	}

	return
}

func (s *Service) ConnectConsumerMatch(reason, bvn, product, name, dob, accountNumber string) (rp *ConsumerMatching, err error) {
	r := RequestObject{
		XMLName:   xml.Name{},
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: &Body{
			ConnectConsumerMatch: &ConnectConsumerMatch{
				Xmlns:          "https://online.firstcentralcreditbureau.com/FirstCentralNigeriaWebService",
				DataTicket:     s.Ticket,
				EnquiryReason:  reason,
				ConsumerName:   name,
				DateOfBirth:    dob,
				Identification: bvn,
				AccountNumber:  accountNumber,
				ProductID:      product,
			},
		},
	}
	var resp *ResponseObject
	resp, err = s.makeRequest(http.MethodPost, ConnectConsumerMatchAction, nil, r)
	if err != nil {
		return
	}

	rp, err = resp.GetConsumerMatchResult()

	return
}

func (s *Service) GetFullCreditReport(mcs []MatchedConsumer, threshold int) (rp *ConsumerFullCredit, err error) {

	if mcs == nil || len(mcs) == 0 {
		rp = &ConsumerFullCredit{
			CreditAgreementSummary: nil,
			SubjectList:            nil,
		}
		return
	}

	cId, ml, enqEngId, enqId, err := processMatchedCustomersForFullCreditReport(mcs, threshold)
	if err != nil {
		return nil, err
	}

	r := RequestObject{
		XMLName:   xml.Name{},
		XmlnsXsi:  "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsXsd:  "http://www.w3.org/2001/XMLSchema",
		XmlnsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: &Body{
			ConsumerFullCreditReport: &ConsumerFullCreditRequest{
				Xmlns:                     "https://online.firstcentralcreditbureau.com/FirstCentralNigeriaWebService",
				DataTicket:                s.Ticket,
				ConsumerID:                cId,
				ConsumerMergeList:         ml,
				SubscriberEnquiryEngineID: enqEngId,
				EnquiryID:                 enqId,
			},
		},
	}
var resp *ResponseObject
	resp, err = s.makeRequest(http.MethodPost, GetConsumerFullCreditReportAction, nil, r)
	if err != nil {
		return
	}

	rp, err = resp.GetConsumerFullCreditResponse()
	return
}

func (s *Service) SearchByBVN(bvn, product, reason string) (cr *CleanedReport, err error) {
	// get matches
	var match *ConsumerMatching
	match, err = s.ConnectConsumerMatch(reason, bvn, product, "", "", "")
	if err != nil {
		return
	}

	var resp *ConsumerFullCredit
	resp, err = s.GetFullCreditReport(match.MatchedConsumer,30)
	if err != nil {
		return
	}

	cr = resp.GetCleanReport(bvn)
return
}

type CleanedReport struct {
	BVN string
	NoHit bool
	ID string
	Records []Record
}

type Record struct {
	Institution    string
	Amount         string
	Status         string
	Balance        string
	AmountOverdue  string
	Classification string
	DisbursalDate  string
	MaturityDate   string
	Source         string
	ReportDate     string
	RefreshedOn    string
	BureauIdentifierEntry string
	BureauIdentifierAccount string
}