package webproxy

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"pushtart/config"
	"pushtart/dnsserv"
	"pushtart/logging"
	"pushtart/tartmanager"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/cloudfoundry/bytefmt"
	sigar "github.com/cloudfoundry/gosigar"
)

type statusData struct {
	Name       string
	Config     *config.Config
	Load       *sigar.LoadAverage
	Uptime     *sigar.Uptime
	Mem        *sigar.Mem
	Swap       *sigar.Swap
	Background string
	CacheUsed  int
	CacheUtil  int
	Diskfree   uint64
	Disktotal  uint64
	NumCPUs    int
	ProcMem    runtime.MemStats
}

func getDiskFreeAndTotal() (uint64, uint64) {
	var stat syscall.Statfs_t
	wd, _ := os.Getwd()
	syscall.Statfs(wd, &stat)
	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize), stat.Blocks * uint64(stat.Bsize)
}

func statusPage(w http.ResponseWriter, r *http.Request) {
	concreteSigar := sigar.ConcreteSigar{}
	mem := sigar.Mem{}
	swap := sigar.Swap{}
	mem.Get()
	swap.Get()
	uptime := sigar.Uptime{}
	uptime.Get()
	avg, err := concreteSigar.GetLoadAverage()
	if err != nil {
		fmt.Fprintf(w, "Failed to get load average: "+err.Error())
		return
	}

	statusColor := "#F0F0F0"
	if config.All().OverrideStatusColor != "" {
		statusColor = config.All().OverrideStatusColor
	}

	funcMap := template.FuncMap{
		"bytesFormat": bytefmt.ByteSize,
		"percent": func(sub, total uint64) string {
			if total == 0 {
				return "NaN"
			}

			p := (sub * 100) / total
			return strconv.Itoa(int(p))
		},
		"boolcolour": func(in bool) template.HTML {
			if in {
				return template.HTML("<span style=\"color: #00AA00;\">Yes</span>")
			}
			return template.HTML("<span style=\"color: #AA0000;\">No</span>")
		},
		"runStats": func(in string) *tartmanager.RunMetrics {
			m, err2 := tartmanager.GetStats(in)
			if err2 != nil {
				logging.Error("status-page", "Failed to get tart stats: ", err)
				return nil
			}
			return m
		},
		"timeformat": func(in uint64) string {
			t := time.Millisecond * time.Duration(in)
			s := strconv.Itoa(int(t.Hours())) + " hours, "
			s += strconv.Itoa(int(t.Minutes())%60) + " minutes, "
			s += strconv.Itoa(int(t.Seconds())%60) + " seconds."
			return s
		},
	}
	t, err := template.New("status").Funcs(funcMap).Parse(statusTemplate)
	if err != nil {
		w.Write([]byte("Template Error: " + err.Error()))
		return
	}

	diskFree, diskTotal := getDiskFreeAndTotal()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	err = t.ExecuteTemplate(w, "status", statusData{
		Name:       config.All().Name,
		Config:     config.All(),
		Load:       &avg,
		Uptime:     &uptime,
		Mem:        &mem,
		Swap:       &swap,
		Background: statusColor,
		CacheUsed:  dnsserv.GetCacheUsed(),
		CacheUtil:  dnsserv.GetCacheUsed() * 100 / config.All().DNS.LookupCacheSize,
		Diskfree:   diskFree,
		Disktotal:  diskTotal,
		NumCPUs:    runtime.NumCPU(),
		ProcMem:    memStats,
	})
	if err != nil {
		w.Write([]byte("Template Exec Error: " + err.Error()))
		return
	}
}

var statusTemplate = `
<html>
  <head>
    <title>Pushtart Status</title>
  </head>

  <body>
  <style>
    .section-header {
      width: 250px;
    }
    .top {
      font-size: 1.25em;
      font-weight: bold;
    }
  	.main {
      width:100%;
  		border:1px solid #C0C0C0;
  		border-collapse:collapse;
  		padding:5px;
  	}
  	.main th {
  		border:1px solid #C0C0C0;
  		padding:5px;
  		background:{{.Background}};
  	}
  	.main td {
  		border:1px solid #C0C0C0;
  		padding:5px;
  	}
  </style>
  <table class="main top">
  	<thead>
    	<tr>
    		<th>{{.Name}}</th>
    	</tr>
  	</thead>
  	<tbody>
    	<tr>
    		<td>
          <table class="main">
            <thead>
              <tr>
                <th class="section-header">Configuration</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>SSH Listener</td>
                <td>{{.Config.SSH.Listener}}</td>
              </tr>
              <tr>
                <td>Config Path</td>
                <td>{{.Config.Path}}</td>
              </tr>
              <tr>
                <td>Data Path</td>
                <td>{{.Config.DataPath}}</td>
              </tr>
              <tr>
                <td>Deployment Path</td>
                <td>{{.Config.DeploymentPath}}</td>
              </tr>
							<tr>
                <td>Run Sentry Interval</td>
                <td>{{.Config.RunSentryInterval}} seconds</td>
              </tr>
              <tr>
                <td>HTTP Proxy Enabled</td>
                <td>{{boolcolour .Config.Web.Enabled}}</td>
              </tr>
              <tr>
                <td>HTTP Proxy Listener</td>
                <td>{{.Config.Web.Listener}}</td>
              </tr>
              <tr>
                <td>HTTP Proxy Domain</td>
                <td>{{.Config.Web.DefaultDomain}}</td>
              </tr>
              <tr>
                <td>DNS Server Enabled</td>
                <td>{{boolcolour .Config.DNS.Enabled}}</td>
              </tr>
              <tr>
                <td>DNS Server Listener</td>
                <td>{{.Config.DNS.Listener}}</td>
              </tr>
              <tr>
                <td>DNS Forwarding Allowed</td>
                <td>{{boolcolour .Config.DNS.AllowForwarding}}</td>
              </tr>
              <tr>
                <td>DNS Cache Size</td>
                <td>{{.Config.DNS.LookupCacheSize}}</td>
              </tr>
              <tr>
                <td>DNS Cache Used</td>
                <td>{{.CacheUtil}}% ({{.CacheUsed}})</td>
              </tr>
            </tbody>
          </table>
        </td>
    	</tr>

			<tr>
    		<td>
          <table class="main">
            <thead>
              <tr>
                <th class="section-header">TLS</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>TLS Enabled</td>
                <td>{{boolcolour .Config.TLS.Enabled}}</td>
              </tr>
              <tr>
                <td>TLS Listener</td>
                <td>{{.Config.TLS.Listener}}</td>
              </tr>
							<tr>
								<td>Certificates loaded</td>
								<td>{{len .Config.TLS.Certs}}</td>
							</tr>
							<tr>
                <td>Redirect apps to HTTPS</td>
                <td>{{boolcolour .Config.TLS.ForceRedirect}}</td>
              </tr>
            </tbody>
          </table>
        </td>
    	</tr>

      <tr>
        <td>
          <table class="main">
            <thead>
              <tr>
                <th class="section-header">System State</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>Server Uptime</td>
                <td>{{.Uptime.Format}}</td>
              </tr>
              <tr>
                <td>Server Load</td>
                <td>
									{{.Load.One}} -- {{.Load.Five}} -- {{.Load.Fifteen}}<br>
									Num logical cores: {{.NumCPUs}}
								</td>
              </tr>
							<tr>
                <td>Server Memory</td>
                <td>
                Total:        {{bytesFormat .Mem.Total}}<br>
                Used:         {{bytesFormat .Mem.Used}}<br>
                Free:         {{bytesFormat .Mem.Free}}<br>
                Actual Used:  {{bytesFormat .Mem.ActualUsed}}<br>
                Actual Free:  {{bytesFormat .Mem.ActualFree}}<br>
                </td>
              </tr>
							<tr>
								<td>Disk</td>
								<td>
								Free:        {{bytesFormat .Diskfree}} ({{percent .Diskfree .Disktotal}}%)<br>
								Total:        {{bytesFormat .Disktotal}}
								</td>
							</tr>
							<tr>
                <td>Process Memory</td>
                <td>
									Allocated: {{bytesFormat .ProcMem.Alloc}}<br>
									Reserved from system: {{bytesFormat .ProcMem.Sys}}<br>
									<br>
									Num Objects: {{.ProcMem.HeapObjects}}<br>
									GC Fraction: {{.ProcMem.GCCPUFraction}}<br>
									Num GC Runs: {{.ProcMem.NumGC}}<br>
								</td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>



      <tr>
        <td>
          <table class="main">
            <thead>
              <tr>
                <th class="section-header">Tarts</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
							{{$memtotal := .Mem.Total}}
              {{range $key, $value := .Config.Tarts}}
              <tr>
                <td>{{$value.Name}} ({{$key}})</td>
                <td>
                  Running: {{boolcolour $value.IsRunning}}<br>
                  PID: {{$value.PID}}<br>
                  Restart on Stop: {{boolcolour $value.RestartOnStop}}<br>
                  Restart Delay Seconds: {{$value.RestartDelaySecs}}<br>
                  Logging Stdout/Stderr: {{boolcolour $value.LogStdout}}<br>

									{{if $value.IsRunning}}
										{{$stats := runStats $key}}
										<br>
										Real Memory: {{bytesFormat $stats.Mem.Resident}} ({{percent $stats.Mem.Resident $memtotal}}%)<br>
										CPU: {{timeformat $stats.Time.Total}}<br>
									{{end}}

									{{if $value.LastHash}}<br><i>{{$value.LastHash}} - {{$value.LastGitMessage}}</i><br>{{end}}
                </td>
              </tr>
              {{end}}
            </tbody>
          </table>
        </td>
      </tr>


      <tr>
        <td>
          <table class="main">
            <thead>
              <tr>
                <th class="section-header">DNS</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {{range $key, $value := .Config.DNS.ARecord}}
              <tr>
                <td>{{$key}}</td>
                <td>
                  {{$value.Address}}<br>
                  TTL: {{$value.TTL}}<br>
                </td>
              </tr>
              {{end}}
            </tbody>
          </table>
        </td>
      </tr>
  	</tbody>
  </table>
  </body>
</html>
`
