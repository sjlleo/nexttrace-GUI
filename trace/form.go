package trace

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sjlleo/nexttrace-GUI/ipgeo"
	"github.com/sjlleo/nexttrace-GUI/util"
	"github.com/sjlleo/nexttrace-GUI/wshandle"
	_ "github.com/ying32/govcl/pkgs/winappres"
	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/types"
)

var (
	MainForm *TMainForm
)

type TMainForm struct {
	*vcl.TForm
	IPInput   *vcl.TEdit
	TraceView *vcl.TListView
}

func (f *TMainForm) OnFormCreate(sender vcl.IObject) {
	f.SetCaption("NextTrace GUI")
	f.SetPosition(types.PoScreenCenter)
	f.SetWidth(900)
	f.SetHeight(600)
	// 双缓冲
	f.SetDoubleBuffered(true)

	// TracePanel
	f.CreateTopPanel()
	f.CreateTraceListPanel()
}

func AddData(f *TMainForm, res *Result, ttl int) {
	var noData = true
	var ip, asn, hostname, latency_str, address_str string
	var latency, address []string
	for _, v := range res.Hops[ttl] {
		if v.Address != nil {
			ip = v.Address.String()
			latency = append(latency, fmt.Sprintf("%.0f", v.RTT.Seconds()*1000))
			if asn != "" {
				asn = v.Geo.Asnumber
			}
			if len(address) == 0 {
				if v.Geo.Country == "" {
					address = append(address, "局域网")
				}
				if v.Geo.Country != "" {
					address = append(address, v.Geo.Country)
				}
				if v.Geo.Prov != "" {
					address = append(address, v.Geo.Prov)
				}
				if v.Geo.City != "" {
					address = append(address, v.Geo.City)
				}

				if v.Geo.Owner != "" {
					address = append(address, v.Geo.Owner)
				}
			}

			hostname = v.Hostname
			noData = false
		} else {
			latency = append(latency, "*")
		}

		if noData {
			ip = "*"
			asn = "*"
		}

		if asn == "" {
			asn = "*"
		}

		latency_str = strings.Join(latency, "／")
		address_str = strings.Join(address, " ")
	}
	vcl.ThreadSync(func() {
		item := f.TraceView.Items().Add()
		// 第一列为Caption属性所管理
		item.SetCaption(fmt.Sprintf("%d", ttl+1))
		item.SubItems().Add(ip)
		item.SubItems().Add(latency_str)
		item.SubItems().Add(address_str)
		item.SubItems().Add(asn)
		item.SubItems().Add(hostname)
	})
}

// func TestAddData() {
// 	for i := 1; i <= 5; i++ {
// 		vcl.ThreadSync(func() {
// 			item := f.TraceView.Items().Add()
// 			// 第一列为Caption属性所管理
// 			item.SetCaption(fmt.Sprintf("%d", i))
// 			item.SubItems().Add(fmt.Sprintf("值：%d", i))

// 		})
// 		<-time.After(500 * time.Millisecond)
// 	}
// }

func (f *TMainForm) NewTrace() {
	ip := util.DomainLookUp(f.IPInput.Text(), false)
	f.TraceView.Clear()
	// 建立 ws
	w := wshandle.New()
	w.Interrupt = make(chan os.Signal, 1)
	signal.Notify(w.Interrupt, os.Interrupt)
	defer func() {
		w.Conn.Close()
	}()
	// 开始 Traceroute
	var conf = Config{
		BeginHop:         1,
		DestIP:           ip,
		MaxHops:          30,
		PacketInterval:   100,
		TTLInterval:      500,
		NumMeasurements:  3,
		ParallelRequests: 18,
		Lang:             "cn",
		RDns:             true,
		AlwaysWaitRDNS:   false,
		IPGeoSource:      ipgeo.GetSource("LeoMoeAPI"),
		Timeout:          1 * time.Second,
	}
	conf.RealtimePrinter = AddData
	Traceroute(ICMPTrace, conf)
}

func (f *TMainForm) CreateTopPanel() {
	pnlUp := vcl.NewPanel(f)
	pnlUp.SetParent(f)
	pnlUp.SetAlign(types.AcoAutoAppend)
	f.IPInput = vcl.NewEdit(f)
	f.IPInput.SetParent(pnlUp)
	// ipEdit.SetCaption("SetSelected")
	f.IPInput.SetWidth(500)
	f.IPInput.SetTop(10)
	f.IPInput.SetLeft(10)
	f.IPInput.SetOnClick(func(sender vcl.IObject) {
	})

	btnTest2 := vcl.NewButton(f)
	btnTest2.SetParent(pnlUp)
	btnTest2.SetTop(10)
	btnTest2.SetLeft(f.IPInput.Left() + f.IPInput.Width() + 10)
	btnTest2.SetWidth(120)
	btnTest2.SetCaption("Start")
	btnTest2.SetOnClick(func(sender vcl.IObject) {
		go f.NewTrace()
	})

}

func (f *TMainForm) CreateTraceListPanel() {
	f.TraceView = vcl.NewListView(f)
	f.TraceView.SetParent(f)
	f.TraceView.SetAlign(types.AlClient)
	f.TraceView.SetRowSelect(true)
	f.TraceView.SetReadOnly(true)
	f.TraceView.SetViewStyle(types.VsReport)
	f.TraceView.SetGridLines(true)
	//lv1.SetColumnClick(false)
	f.TraceView.SetHideSelection(false)

	col := f.TraceView.Columns().Add()
	col.SetCaption("#")
	col.SetWidth(40)

	col = f.TraceView.Columns().Add()
	col.SetCaption("IP")
	col.SetWidth(130)

	col = f.TraceView.Columns().Add()
	col.SetCaption("时间(ms)")
	col.SetWidth(100)

	col = f.TraceView.Columns().Add()
	col.SetCaption("地址")
	col.SetWidth(280)

	col = f.TraceView.Columns().Add()
	col.SetCaption("AS")
	col.SetWidth(100)

	col = f.TraceView.Columns().Add()
	col.SetCaption("主机名")
	col.SetWidth(200)
}
