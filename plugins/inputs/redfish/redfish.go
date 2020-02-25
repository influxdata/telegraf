package redfish
import (
	"github.com/influxdata/telegraf"
        "github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/base64"
	"crypto/tls"
	"encoding/json"
	"strconv"
)
type Redfish struct{
   Host string `toml:"host"`
   BasicAuthUsername string `toml:"basicauthusername"`
   BasicAuthPassword string `toml:"basicauthpassword"`
   Id string `toml:"id"`
   Server string `toml:"server"`
   parser parsers.Parser `toml:"parser"`
   Timeout internal.Duration `toml:"timeout"`
}
type Hostname struct {
   Hostname string `json:"HostName"`
}
type Cpu struct {
    Name string `json:"Name"`
    Temperature int64 `json:"ReadingCelsius"`
    Status CpuStatus `json:"Status"`
}
type Temperatures struct {
   Temperatures []Cpu `json:"Temperatures"`
}
type CpuStatus struct {
   State string `json:"State"`
   Health string `json:"Health"`
}
type Fans struct {
   Fans []speed `json:"Fans"`
}
type speed struct {
   Name string `json:"Name"`
   Speed  int64 `json:"Reading"`
   Status FansStatus `json:"Status"`
}
type FansStatus struct {
   State string `json:"State"`
   Health string `json:"Health"`
}
type PowerSupplies struct {
   PowerSupplies []watt `json:"PowerSupplies"`
}
type PowerSupplieshp struct {
   PowerSupplieshp  []watthp `json:"PowerSupplies"`
}
type watt struct {
   Name string `json:"Name"`
   PowerInputWatts float64 `json:"PowerInputWatts"`
   PowerCapacityWatts float64 `json:"PowerCapacityWatts"`
   PowerOutputWatts float64 `json:"PowerOutputWatts"`
   Status PowerStatus `json:"Status"`
}
type watthp struct{
   Name string `json:"Name"`
   MemberId string `json:"MemberId"`
   PowerCapacityWatts float64 `json:"PowerCapacityWatts"`
   LastPowerOutputWatts float64 `json:"LastPowerOutputWatts"`
   LineInputVoltage float64 `json:"LineInputVoltage"`
}
type PowerStatus struct {
   State string `json:"State"`
   Health string `json:"Health"`
}
type Voltages struct {
  Voltages []volt `json:"Voltages"`
}
type volt struct{
  Name string `json:"Name"`
  ReadingVolts int64 `json:"ReadingVolts"`
  Status VoltStatus `json:"Status"`
}
type VoltStatus struct {
   State string `json:"State"`
   Health string `json:"Health"`
}
type Location struct{
  Location Address `json:"Location"`
}
type Address struct{
  PostalAddress PostalAddress `json:"PostalAddress"`
  Placement Placement `json:"Placement"`
}
type PostalAddress struct{
  DataCenter string `json:"Building"`
  Room string `json:"Room"`
 }
type Placement struct{
 Rack string `json:"Rack"`
  Row string `json:"Row"`
}


var h Hostname
var t Temperatures
var f Fans
var p PowerSupplies
var v Voltages
var php PowerSupplieshp
var l Location

func basicAuth(username, password string) string {
  auth := username + ":" + password
   return base64.StdEncoding.EncodeToString([]byte(auth))
}

func getThermal(host,username,password,id string) (error){
		r := fmt.Sprint(host,"/redfish/v1/Chassis/",id,"/Thermal")
		client := &http.Client{}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest("GET", r, nil)
		req.Header.Add("Authorization","Basic "+ basicAuth(username,password))
		req.Header.Set("Accept","application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
                if err != nil{
                 return err
                 }
		if resp.StatusCode == 200 {
		  body, _ := ioutil.ReadAll(resp.Body)
		  jsonErr := json.Unmarshal(body,&t)
		  if jsonErr != nil {
		   return fmt.Errorf("error parsing input: %v", jsonErr)
		  }
                  jsonErr = json.Unmarshal(body,&f)
                  if jsonErr != nil {
		   return fmt.Errorf("error parsing input: %v", jsonErr)
                  }
		}else {
		  return fmt.Errorf("received status code %d (%s), expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
		}
		return nil
}

func getPower(host,username,password,id,server string)(error){
                r := fmt.Sprint(host,"/redfish/v1/Chassis/",id,"/Power")
                client := &http.Client{}
                http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
                req, err := http.NewRequest("GET", r, nil)
                req.Header.Add("Authorization","Basic "+ basicAuth(username,password))
                req.Header.Set("Accept","application/json")
                req.Header.Set("Content-Type", "application/json")
                resp, err := client.Do(req)
                if err != nil{
                 return err
                 }
                if resp.StatusCode == 200 {
                  body, _ := ioutil.ReadAll(resp.Body)
		  if server == "dell"{
		    jsonErr := json.Unmarshal(body,&p)
                    if jsonErr != nil {
		     return fmt.Errorf("error parsing input: %v", jsonErr)
                     }
                    jsonErr = json.Unmarshal(body,&v)
                    if jsonErr != nil {
		      return fmt.Errorf("error parsing input: %v", jsonErr)
                    }
		   }
		  if server == "hp"{
                   jsonErr := json.Unmarshal(body,&php)
                   if jsonErr != nil {
                     return fmt.Errorf("error parsing input: %v", jsonErr)
                  }
		}
                 return nil
	    }else {
                  return fmt.Errorf("received status code %d (%s), expected 200",
                        resp.StatusCode,
                        http.StatusText(resp.StatusCode))
		}
}

func getHostname(host,username,password,id string) (error){
        url := fmt.Sprint(host,"/redfish/v1/Systems/",id)
        client := &http.Client{}
        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        req, err := http.NewRequest("GET", url, nil)
        req.Header.Add("Authorization","Basic "+ basicAuth(username,password))
        req.Header.Set("Accept","*/*")
        req.Header.Set("Content-Type", "application/json")
        resp, err := client.Do(req)
        if err != nil{
                 return err
             }
	if resp.StatusCode == 200 {
         body, _ := ioutil.ReadAll(resp.Body)
         jsonErr := json.Unmarshal(body,&h)
         if jsonErr != nil {
             return fmt.Errorf("error parsing input: %v", jsonErr)
         }
	} else {
             return fmt.Errorf("received status code %d (%s), expected 200",
             resp.StatusCode,
             http.StatusText(resp.StatusCode))
                }
		return nil
}

func getLocation(host,Username,password,id string)(error){
                r := fmt.Sprint(host,"/redfish/v1/Chassis/",id,"/")
                client := &http.Client{}
                http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
                req, err := http.NewRequest("GET", r, nil)
                req.Header.Add("Authorization","Basic "+ basicAuth(Username,password))
                req.Header.Set("Accept","application/json")
                req.Header.Set("Content-Type", "application/json")
                resp, err := client.Do(req)
                if err != nil{
                 return err
                 }
		if resp.StatusCode == 200 {
                 body, _ := ioutil.ReadAll(resp.Body)
                 jsonErr := json.Unmarshal(body,&l)
                 if jsonErr != nil {
                    return fmt.Errorf("error parsing input: %v", jsonErr)
                }
	    } else {
                    return fmt.Errorf("received status code %d (%s), expected 200",
                     resp.StatusCode,
                     http.StatusText(resp.StatusCode))
              }

		return nil
}



func (r *Redfish) Description() string {
  return "Read CPU, Fans, Powersupply and Voltage metrics of Dell/HP hardware server through redfish APIs"
}

var redfishConfig =`
## Server OOB-IP
host = "https://192.0.0.1"

## Username,Password for hardware server
basicauthusername = "test"
basicauthpassword = "test"
## Server Vendor(dell or hp)
server= "dell"
## Resource Id for redfish APIs
id="System.Embedded.1"

## Amount of time allowed to complete the HTTP request
# timeout = "5s"
`
func (r *Redfish) SampleConfig() string {
  return redfishConfig
}


func (r *Redfish) SetParser(parser parsers.Parser) {
  r.parser = parser
}


func (r *Redfish) Init() error {
  return nil
}

func (r *Redfish) Gather(acc telegraf.Accumulator) error {

	if len(r.Host) > 0 && len(r.BasicAuthUsername) > 0 && len(r.BasicAuthPassword) > 0 && len(r.Server) > 0 && len(r.Id) > 0 && (r.Server == "dell" || r.Server == "hp"){
		err := getThermal(r.Host,r.BasicAuthUsername,r.BasicAuthPassword,r.Id)
		if err != nil {
			return err
		}
		err = getHostname(r.Host,r.BasicAuthUsername,r.BasicAuthPassword,r.Id)
                if err != nil {
                        return err
                }

		err = getPower(r.Host,r.BasicAuthUsername,r.BasicAuthPassword,r.Id,r.Server)
                if err != nil {
                        return err
                }


		if r.Server == "dell" {
		  err = getLocation(r.Host,r.BasicAuthUsername,r.BasicAuthPassword,r.Id)
                if err != nil {
                        return err
                }

		}

		for i := 0; i < len(t.Temperatures); i++ {
			//  Tags
//			tags := map[string]string{"Name": t.Temperatures[i].Name}
			tags := map[string]string{"OOBIP": r.Host,"Name": t.Temperatures[i].Name,"Hostname": h.Hostname,}
			//  Fields
			fields := make(map[string]interface{})
			fields["Temperature"] = strconv.FormatInt(t.Temperatures[i].Temperature,10)
			fields["State"] = t.Temperatures[i].Status.State
			fields["Health"] = t.Temperatures[i].Status.Health
			if r.Server == "dell"{
				fields["Datacenter"] = l.Location.PostalAddress.DataCenter
				fields["Room"] = l.Location.PostalAddress.Room
				fields["Rack"] = l.Location.Placement.Rack
				fields["Row"] = l.Location.Placement.Row
				acc.AddFields("cputemperature", fields, tags)
			}
			if r.Server == "hp"{
			acc.AddFields("cputemperature", fields, tags)
			}
		}
                for i := 0; i < len(f.Fans); i++ {
                        //  Tags
                        tags := map[string]string{"OOBIP": r.Host,"Name": f.Fans[i].Name,"Hostname": h.Hostname}
                        //  Fields
                        fields := make(map[string]interface{})
                        fields["Fanspeed"] = strconv.FormatInt(f.Fans[i].Speed,10)
                        fields["State"] = f.Fans[i].Status.State
                        fields["Health"] = f.Fans[i].Status.Health
                        if r.Server == "dell" {
                                fields["Datacenter"] = l.Location.PostalAddress.DataCenter
                                fields["Room"] = l.Location.PostalAddress.Room
                                fields["Rack"] = l.Location.Placement.Rack
                                fields["Row"] = l.Location.Placement.Row
                                acc.AddFields("fans", fields, tags)
                        }
                        if r.Server == "hp" {
                         acc.AddFields("fans", fields, tags)
			}
		}
		if r.Server == "dell" {
                 for i := 0; i < len(p.PowerSupplies); i++ {
                        //  Tags
                        tags := map[string]string{"OOBIP": r.Host,"Name": p.PowerSupplies[i].Name,"Hostname": h.Hostname}
                        //  Fields
                        fields := make(map[string]interface{})
                        fields["PowerInputWatts"] = strconv.FormatFloat(p.PowerSupplies[i].PowerInputWatts,'f',-1,64)
			fields["PowerCapacityWatts"] = strconv.FormatFloat(p.PowerSupplies[i].PowerCapacityWatts,'f',-1,64)
			fields["PowerOutputWatts"] = strconv.FormatFloat(p.PowerSupplies[i].PowerOutputWatts,'f',-1,64)
                        fields["State"] = p.PowerSupplies[i].Status.State
                        fields["Health"] = p.PowerSupplies[i].Status.Health
			fields["Datacenter"] = l.Location.PostalAddress.DataCenter
                        fields["Room"] = l.Location.PostalAddress.Room
                        fields["Rack"] = l.Location.Placement.Rack
                        fields["Row"] = l.Location.Placement.Row
                        acc.AddFields("powersupply", fields, tags)
			}
		}
                if r.Server == "hp" {
                 for i := 0; i < len(php.PowerSupplieshp); i++ {
                        //  Tags
                        tags := map[string]string{"OOBIP": r.Host,"Name": php.PowerSupplieshp[i].Name,"MemberId" : php.PowerSupplieshp[i].MemberId,"Hostname": h.Hostname}
                        //  Fields
                        fields := make(map[string]interface{})
                        fields["LineInputVoltage"] = strconv.FormatFloat(php.PowerSupplieshp[i].LineInputVoltage,'f',-1,64)
                        fields["PowerCapacityWatts"] = strconv.FormatFloat(php.PowerSupplieshp[i].PowerCapacityWatts,'f',-1,64)
                        fields["LastPowerOutputWatts"] = strconv.FormatFloat(php.PowerSupplieshp[i].LastPowerOutputWatts,'f',-1,64)
                        acc.AddFields("powersupply", fields, tags)
                        }
                }

		if r.Server == "dell" {
		 for i := 0; i < len(v.Voltages); i++ {
                        //  Tags
                        tags := map[string]string{"OOBIP": r.Host,"Name": v.Voltages[i].Name,"Hostname": h.Hostname}
                        //  Fields
                        fields := make(map[string]interface{})
                        fields["Voltage"] = strconv.FormatInt(v.Voltages[i].ReadingVolts,10)
                        fields["State"] = v.Voltages[i].Status.State
                        fields["Health"] = v.Voltages[i].Status.Health
			fields["Datacenter"] = l.Location.PostalAddress.DataCenter
                        fields["Room"] = l.Location.PostalAddress.Room
                        fields["Rack"] = l.Location.Placement.Rack
                        fields["Row"] = l.Location.Placement.Row
                        acc.AddFields("voltages", fields, tags)
			}
		 }
		return nil
    }else {
		return fmt.Errorf("Did not provide all the mandatory fields in the configuration")
		}

}

func init() {
  inputs.Add("redfish", func() telegraf.Input { return &Redfish{} })
}
