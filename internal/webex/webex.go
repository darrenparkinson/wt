package webex

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type tokenEnvelope struct {
	Body tokenBody `xml:"Body"`
}
type tokenBody struct {
	Response tokenResponse `xml:"getMeetingTicketResponse"`
	Fault    tokenFault    `xml:"Fault"`
}
type tokenResponse struct {
	Token string `xml:"getMeetingTicketReturn"`
}
type tokenFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// ListingMessage is the top level xml object consisting of the header and the body
type ListingMessage struct {
	ListingHeader ListingHeader `xml:"header"`
	ListingBody   ListingBody   `xml:"body"`
}

// ListingHeader is the header document element of the listing
type ListingHeader struct {
	ListingResult    string `xml:"response>result"`
	ListingStatus    string `xml:"response>gsbStatus"`
	ListingReason    string `xml:"response>reason"`
	ListingException string `xml:"response>exceptionID"`
}

// ListingBody is the body document element of the listing
type ListingBody struct {
	MatchingRecords string             `xml:"bodyContent>matchingRecords>total"`
	ReturnedRecords string             `xml:"bodyContent>matchingRecords>returned"`
	StartFrom       string             `xml:"bodyContent>matchingRecords>startFrom"`
	Recordings      []ListingRecording `xml:"bodyContent>recording"`
}

// ListingRecording is an individual recording item details
type ListingRecording struct {
	RecordingID         string  `xml:"recordingID"`
	HostWebExID         string  `xml:"hostWebExID"`
	RecordingName       string  `xml:"name"`
	RecordingCreateTime string  `xml:"createTime"`
	RecordingTimeZoneID string  `xml:"timeZoneID"`
	RecordingSize       float32 `xml:"size"`
	StreamURL           string  `xml:"streamURL"`
	FileURL             string  `xml:"fileURL"`
	RecordingType       int     `xml:"recordingType"`
	Duration            int     `xml:"duration"`
	Format              string  `xml:"format"`
	ServiceType         string  `xml:"serviceType"`
	ConfID              string  `xml:"confID"`
	Password            string  `xml:"password"`
	PasswordReq         string  `xml:"passwordReq"`
}

// GetHeaders is a struct method to return the Headers for CSV Output
func (r ListingRecording) GetHeaders() []string {
	return []string{"RecordingID", "HostWebExID", "RecordingName", "RecordingCreateTime", "RecordingSize", "StreamURL", "FileURL", "RecordingType", "Duration", "Format", "ServiceType", "ConfID", "Password", "PasswordReq"}
}

// ToSlice is a struct method to return the string representation of each recording
func (r ListingRecording) ToSlice() []string {
	return []string{r.RecordingID, r.HostWebExID, r.RecordingName, r.RecordingCreateTime, strconv.FormatFloat(float64(r.RecordingSize), 'f', 3, 64), r.StreamURL, r.FileURL, strconv.FormatInt(int64(r.RecordingType), 10), strconv.FormatInt(int64(r.Duration), 10), r.Format, r.ServiceType, r.ConfID, r.Password, r.PasswordReq}
}

// GetRecordingList will retrieve a list of recordings from a given date until 28 days later (due to Cisco API restriction).
func GetRecordingList(username, password, tenant, siteID string, fromDate, toDate time.Time) (ListingMessage, error) {
	webexURL := fmt.Sprintf("https://%s.webex.com/WBXService/XMLService", tenant)
	toDateString := toDate.Format("01/02/2006 15:04:05")
	fromDateString := fromDate.Format("01/02/2006 15:04:05")
	dateScope := fmt.Sprintf("<createTimeScope><createTimeStart>%s</createTimeStart><createTimeEnd>%s</createTimeEnd></createTimeScope>", fromDateString, toDateString)
	// clearly should use the xml encoding instead...
	xmlBody := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><serv:message xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:serv=\"http://www.webex.com/schemas/2002/06/service\"><header><securityContext><webExID>%s</webExID><password>%s</password><siteID>%s</siteID></securityContext></header><body><bodyContent xsi:type=\"java:com.webex.service.binding.ep.LstRecording\">%s<listControl><startFrom>0</startFrom><maximumNum>500</maximumNum></listControl></bodyContent></body></serv:message>", username, password, siteID, dateScope)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", webexURL, strings.NewReader(xmlBody))
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "''")
	req.Header.Set("Accept", "application/soap+xml, application/dime, multipart/related, text/*")
	resp, err := client.Do(req)
	if err != nil {
		return ListingMessage{}, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	var soap ListingMessage
	err = xml.Unmarshal([]byte(s), &soap)
	if err != nil {
		return ListingMessage{}, err
	}
	// 000015 is no records found
	if soap.ListingHeader.ListingResult != "SUCCESS" && soap.ListingHeader.ListingException != "000015" {
		return ListingMessage{}, fmt.Errorf("%s: %s", soap.ListingHeader.ListingResult, soap.ListingHeader.ListingReason)
	}

	return soap, nil
}

// GetRecordingListForUser will retrieve a list of recordings for a given userid (hostWebExID).
func GetRecordingListForUser(username, password, userid, tenant, siteID string, fromDate, toDate time.Time) (ListingMessage, error) {
	webexURL := fmt.Sprintf("https://%s.webex.com/WBXService/XMLService", tenant)
	toDateString := toDate.Format("01/02/2006 15:04:05")
	fromDateString := fromDate.Format("01/02/2006 15:04:05")
	dateScope := fmt.Sprintf("<createTimeScope><createTimeStart>%s</createTimeStart><createTimeEnd>%s</createTimeEnd></createTimeScope>", fromDateString, toDateString)
	xmlBody := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><serv:message xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:serv=\"http://www.webex.com/schemas/2002/06/service\"><header><securityContext><webExID>%s</webExID><password>%s</password><siteID>%s</siteID></securityContext></header><body><bodyContent xsi:type=\"java:com.webex.service.binding.ep.LstRecording\">%s<listControl><startFrom>0</startFrom><maximumNum>500</maximumNum></listControl><hostWebExID>%s</hostWebExID></bodyContent></body></serv:message>", username, password, siteID, dateScope, userid)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", webexURL, strings.NewReader(xmlBody))
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "''")
	req.Header.Set("Accept", "application/soap+xml, application/dime, multipart/related, text/*")
	resp, err := client.Do(req)
	if err != nil {
		return ListingMessage{}, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	var soap ListingMessage
	err = xml.Unmarshal([]byte(s), &soap)
	if err != nil {
		return ListingMessage{}, err
	}
	// 000015 is no records found
	if soap.ListingHeader.ListingResult != "SUCCESS" && soap.ListingHeader.ListingException != "000015" {
		return ListingMessage{}, fmt.Errorf("%s: %s", soap.ListingHeader.ListingResult, soap.ListingHeader.ListingReason)
	}

	return soap, nil
}

// DownloadRecording is used to download a recording with a given recordingID.
func DownloadRecording(username, password, tenant, siteID, domain, recordingID string) (string, error) {
	webexURL := fmt.Sprintf("https://%s.webex.com/nbr/services/", domain)
	token, err := GetWebExToken(username, password, tenant, siteID, domain)
	if err != nil {
		return "", err
	}
	xmlBody := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><soapenv:Body><ns1:downloadNBRStorageFile soapenv:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\" xmlns:ns1=\"NBRStorageService\"><siteId xsi:type=\"xsd:long\">%s</siteId><recordId xsi:type=\"xsd:long\">%s</recordId><ticket xsi:type=\"xsd:string\">%s</ticket></ns1:downloadNBRStorageFile></soapenv:Body></soapenv:Envelope>", siteID, recordingID, token)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", webexURL, strings.NewReader(xmlBody))
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "''")
	req.Header.Set("Accept", "application/soap+xml, application/dime, multipart/related, text/*")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	mediaType, params, error := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if error != nil {
		return "", error
	}
	if strings.HasPrefix(mediaType, "multipart/") {

		mr := multipart.NewReader(resp.Body, params["boundary"])

		// There are three parts -- the xml file response, a file telling us the size and name, the binary recording
		// filename := ""
		embeddedfilename := ""
		for i := 0; i < 3; i++ {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}
			// XML File
			if i == 0 {
				// do nothing.  could ioutil.ReadAll(p) if necessary
			}

			// File name and Size and Encryption
			if i == 1 {
				slurp, err := ioutil.ReadAll(p)
				if err != nil {
					return "", err
				}
				embeddedfilename = strings.Split(string(slurp), "\n")[0]
				fmt.Println("Downloading " + embeddedfilename)
			}

			// The Recording
			if i == 2 && embeddedfilename != "" {
				fo, err := os.Create(embeddedfilename)
				if err != nil {
					return "", err
				}
				defer fo.Close()
				w := bufio.NewWriter(fo)
				buf := make([]byte, 1024)
				for {
					n, err := p.Read(buf)
					if err != nil && err != io.EOF {
						return "", err
					}
					if n == 0 {
						break
					}
					if _, err := w.Write(buf[:n]); err != nil {
						return "", err
					}
				}

			}
		}

	}
	return "", nil

}

// GetWebExToken gets a token from WebEx for use with further requests
func GetWebExToken(username, password, tenant, siteID, domain string) (string, error) {
	service := "MC"
	tokenURL := fmt.Sprintf("https://%s.webex.com/nbr/services/nbrXmlService?method=getMeetingTicket&siteId=%s&username=%s&password=%s&service=%s", domain, siteID, username, password, service)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", tokenURL, nil)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "''")
	req.Header.Set("Accept", "application/soap+xml, application/dime, multipart/related, text/*")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	var soap tokenEnvelope
	err = xml.Unmarshal([]byte(s), &soap)
	if err != nil {
		return "", err
	}

	if soap.Body.Fault.FaultCode != "" && soap.Body.Response.Token == "" {
		return "", fmt.Errorf("FaultCode: %s FaultString: %s", soap.Body.Fault.FaultCode, soap.Body.Fault.FaultString)
	}
	//TODO: Check all the error codes and make this more intelligent.
	if soap.Body.Response.Token != "" && strings.HasPrefix(soap.Body.Response.Token, "AS") {
		return "", fmt.Errorf("%s", soap.Body.Response.Token)
	}
	return soap.Body.Response.Token, nil
}
