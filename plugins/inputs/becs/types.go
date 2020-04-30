package becs

import (
	"encoding/xml"
)

type envelope struct {
	XMLName xml.Name `xml:"env:Envelope"`
	Env     string   `xml:"xmlns:env,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xsd     string   `xml:"xmlns:xsd,attr"`
	Becs    string   `xml:"xmlns:becs,attr"`
	SoapEnc string   `xml:"xmlns:soap-enc,attr"`
	Header  struct {
		Request struct {
			Type      string `xml:"xsi:type,attr"`
			SessionID struct {
				Type string `xml:"xsi:type,attr"`
				ID   string `xml:",chardata"`
			} `xml:"sessionid"`
		} `xml:"urn:packetfront_becs request"`
	} `xml:"env:Header"`
	Body body
}

type body struct {
	XMLName xml.Name `xml:"env:Body"`
	Env     string   `xml:"env:encodingStyle,attr"`
	ID      string   `xml:"id,attr"`
	Method  interface{}
}

type sessionLogin struct {
	XMLName xml.Name `xml:"becs:sessionLogin"`
	In      struct {
		Type      string `xml:"xsi:type,attr"`
		Username  string `xml:"username"`
		Password  string `xml:"password"`
		Namespace string `xml:"namespace,omitempty"`
	} `xml:"in"`
}

type sessionLoginResponse struct {
	Body struct {
		Response struct {
			Out struct {
				Err       uint   `xml:"err"`
				ErrTxt    string `xml:"errtxt,omitempty"`
				SessionID string `xml:"sessionid"`
			} `xml:"out"`
		} `xml:"sessionLoginResponse"`
	}
}

type applicationList struct {
	XMLName xml.Name `xml:"becs:applicationList"`
	In      struct {
		Type string `xml:"xsi:type,attr"`
	} `xml:"in"`
}

type applicationListResponse struct {
	Body struct {
		Response struct {
			Out struct {
				Err    uint   `xml:"err"`
				ErrTxt string `xml:"errtxt,omitempty"`
				Names  struct {
					Items []string `xml:"item"`
				} `xml:"names"`
			} `xml:"out"`
		} `xml:"applicationListResponse"`
	}
}

type applicationStatusGet struct {
	XMLName xml.Name `xml:"becs:applicationStatusGet"`
	In      struct {
		Type         string `xml:"xsi:type,attr"`
		Name         string `xml:"name"`
		IncludePools bool   `xml:"includepools,omitempty"`
	} `xml:"in"`
}

type applicationStatusGetResponse struct {
	Body struct {
		Response struct {
			Out struct {
				Err          uint   `xml:"err"`
				ErrTxt       string `xml:"errtxt,omitempty"`
				Displayname  string `xml:"displayname"`
				Hostname     string `xml:"hostname"`
				Version      string `xml:"version"`
				UpTime       uint   `xml:"uptime"`
				StartTime    uint   `xml:"starttime"`
				CPUUsage     uint   `xml:"cpuusage"`
				CPUAverage60 uint   `xml:"cpuaverage60"`
				MemoryPools  struct {
					Items []memoryPool `xml:"item"`
				} `xml:"memorypools,omitempty"`
			} `xml:"out"`
		} `xml:"applicationStatusGetResponse"`
	}
}

type metricGet struct {
	XMLName xml.Name `xml:"becs:metricGet"`
	In      struct {
		Type string `xml:"xsi:type,attr"`
		Name string `xml:"name"`
	} `xml:"in"`
}

type metricGetResponse struct {
	Body struct {
		Response struct {
			Out struct {
				Err     uint   `xml:"err"`
				ErrTxt  string `xml:"errtxt,omitempty"`
				Metrics struct {
					Items []metric `xml:"item"`
				} `xml:"metrics,omitempty"`
			} `xml:"out"`
		} `xml:"metricGetResponse"`
	}
}

type memoryPool struct {
	Name       string `xml:"name"`
	Size       uint   `xml:"size"`
	Out        uint   `xml:"out"`
	Pages      uint   `xml:"pages"`
	EmptyPages uint   `xml:"emptypages"`
}

type metric struct {
	Name   string `xml:"name"`
	Type   string `xml:"type"`
	Values struct {
		Items []metricValue `xml:"item"`
	} `xml:"values,omitempty"`
}

type metricValue struct {
	Labels struct {
		Items []metricLabel `xml:"item"`
	} `xml:"labels,omitempty"`
	Value uint64 `xml:"value"`
}

type metricLabel struct {
	Name  string `xml:"name,omitempty"`
	Value string `xml:"value,omitempty"`
}

type clientFind struct {
	XMLName xml.Name `xml:"becs:clientFind"`
	In      struct {
		Type  string `xml:"xsi:type,attr"`
		IP    string `xml:"ip"`
		Limit uint   `xml:"limit"`
	} `xml:"in"`
}

type clientFindResponse struct {
	Body struct {
		Response struct {
			Out struct {
				Err    uint   `xml:"err"`
				ErrTxt string `xml:"errtxt,omitempty"`
				Actual uint   `xml:"actual"`
			} `xml:"out"`
		} `xml:"clientFindResponse"`
	}
}
