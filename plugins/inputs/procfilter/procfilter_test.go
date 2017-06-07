// parse project main.go
package procfilter

// DEBUG/OPTIM
/*import (
	"log"
	"net/http"
	_ "net/http/pprof"
)*/

import (
	"fmt"
	"strings"
	"testing"
	"time"
	//"encoding/json"
)

var testScripts = []string{
	//`m = tag(cmd) field(cpu) <- top(cpu,2,user('^ol'r!))`,
	`# Ressource hogs
     top.cmd.cpu = tag(cmd) field(cpu,rss,vsz,swap,cmd_line) <- top(cpu,3)
     top.cmd.rss = tag(cmd) field(rss) <- top(rss,3)
     top.cmd.swap = tag(cmd) field(cpu,rss,vsz,swap,cmd_line) <- top(swap,3)
     top.cmd.thread_nb = tag(cmd) field(cpu,rss,vsz,swap,thread_nb,cmd_line) <- top(thread_nb,3)

     # Pack by user
     top.by.user.cpu = tag(user) field(cpu) <- top(cpu,3,by(user))
     top.by.user.rss = tag(user) field(rss) <- top(rss,3,by(user))
     top.by.user.swap = tag(user) field(swap) <- top(swap,3,by(user))
     top.by.user.process_nb = tag(user) field(process_nb) <- top(process_nb,3,by(user))

     # Workloads
     wl.omni = field(process_nb,cpu,rss,vsz,swap) <- pack(children(or(cmd('omv_'r),user('omni'))))
     wl.influx = field(process_nb,cpu,rss,vsz,swap) <- pack(cmd('influxdb'r),user('influxdb'))
     wl.telegraf = field(process_nb,cpu,rss,vsz,swap) <- pack(user('telegraf'))
     wl.root = field(process_nb,cpu,rss,vsz,swap) <- pack(user(0))
     wl._other = field(process_nb,cpu,rss,vsz,swap) <- pack(not(filters('^wl[.]'r)))
	`,

	//`m = tag(user,uid) field(rss) <- by(user)`,
	//`m = field(rss) <- top(rss,2,by(user))`,
	//`top_cpu = tag(cmd) field(cpu) <- top(cpu,2)`,
	/*`np_o <- pack(user("ol"r))
	  np_r <- pack( user(0))
	  or <- filters("^np_.*"r)
	  mo = field(process_nb) <-np_o
	  mr = field(process_nb) <-np_r
	  mor = field(process_nb) <-or
	  pmor = field(process_nb) <-pack(or)`,
	*/
	//`top.rss = tag(cmd) field(rss)<-top(rss,2)`,
	//`by.user = tag(user) field(cpu,rss)<-by(user)`,
	//`exceed.cpu = tag(cmd) field(cpu)<-exceed(cpu,5)`,
	/*`
	  o_t <- top(rss,2,user('ol'r))
	  o_t = tag(cmd,exe) fields(rss) <- o_t
	  wl_unm = field(rss,cpu,swap) <- pack(not(o_t))
	  `,*/
	//`m1 = tag(cmd) fields(rss) <- top(rss,2,all)`,
	//`m1 = tag(cmd) fields(rss) <- top(rss,5,all)`,
	//`m2 = tag(cmd) fields(rss) <- exceed(rss,300000)`,
	//`g = field(rss) <- gather(user("ol"r))`,
	//`m2 = tag(cmd) fields(rss,cpu) <- top(3,rss,cmd("chrome"r))`,
	//`m2 = tag(cmd) fields(rss) <- top(rss,3,cmdline("chrome --type=renderer"r))`,
	//`m2 = tag(cmd) fields(rss) <- top(rss,3,args("renderer"r))`,
	//`m2 = tag(cmd) fields(rss) <- top(rss,3,user("ol"r))`,
	//`o<-user("oliv") mx = tag(cmd) fields(rss) <- o mx = tag(cmd) value(cpu) <- o`,
	//`# comment to remove`,
	//`mn = fields(rss) <- top(rss,2,all())`,
	//`mn = fields(rss) <- top(rss,2,all)`,
	//`mn = fields(rss) <- top(rss,2)`,
	/*`a <- top(rss,10)
	  m3 = tag(cmd) fields(rss)<-a`,*/
	//`apache = fields(cpu,rss,vss) <- cmd("apache")`,
	//`# comment to remove
	/*a <- cmd('apa.*')
	  b <- user('omni')
	  m4 = tags(name) values(nb,cpu,rss) <- a`,*/
}

func TestScan(*testing.T) {
	for _, conf := range testScripts {
		fmt.Printf("\nScan: \"%s\"\n", conf)
		r := strings.NewReader(conf)
		scanner := newScanner(r)
		for {
			tok, lit := scanner.scan()
			fmt.Printf("%d:%q ", tok, lit)
			if tok == tTIllegal {
				fmt.Printf("Illegal:%q\n", lit)
				break
			} else if tok == tTEOF {
				fmt.Printf("EOF\n")
				break
			}
		}
	}
}

func TestParse(*testing.T) {
	// test conf

	for _, conf := range testScripts {
		fmt.Printf("\nParset \"%s\"\n", conf)
		r := strings.NewReader(conf)
		parser := NewParser(r)
		err := parser.Parse()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func TestApply(*testing.T) {
	/*go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()*/
	// test conf
	for _, conf := range testScripts {
		fmt.Printf("\nTest \"%s\"\n", conf)
		pf := NewProcFilter()
		pf.Script = conf
		for i := 0; i < 2; i++ {
			err := testGather(pf)
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(2 * time.Second)
			fmt.Printf("Done: %d sample\n", i)
		}
	}
}

// This is the ProcFilter.Gather() sligthly modified for test/debug.
func testGather(p *ProcFilter) error {
	if p.parser == nil {
		// Init and parse the script to build the AST.
		p.parser = NewParser(strings.NewReader(p.Script))
		err := p.parser.Parse()
		if err != nil {
			return err
		}
		p.parseOK = true
	}
	if !p.parseOK {
		// Data stored in the parser may be inconsistent, do not gather.
		return nil
	}

	// Use the ASTs stored in the parser to process all filters then output the measurements.
	parser := p.parser
	if len(parser.measurements) == 0 {
		// No measurement, do nothing!
		return nil
	}
	// Change the current stamp and update all global variables
	newSample()
	for _, m := range parser.measurements {
		fmt.Printf("%s\n", m.name)
		err := m.f.Apply()
		if err != nil {
			return err
		}
		iStats := m.f.Stats()
		for _, ps := range iStats.pid2Stat {
			tags, err := m.getTags(ps, p.Tag_prefix)
			if err != nil {
				return err
			}
			fields, err := m.getFields(ps, p.Field_prefix)
			if err != nil {
				return err
			}
			// Display the resuls
			for n, v := range tags {
				fmt.Printf("  %s=%s\n", n, v)
			}
			for n, v := range fields {
				fmt.Printf("  %s=%v\n", n, v)
			}
		}
		fmt.Println()
	}
	return nil
}
