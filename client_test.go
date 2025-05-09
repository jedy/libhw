package libhw

import (
	"context"
	"net/netip"
	"os"
	"testing"

	"github.com/libdns/libdns"
)

var AK = os.Getenv("CLOUD_SDK_AK")
var SK = os.Getenv("CLOUD_SDK_SK")
var TOP_DOMAIN = os.Getenv("TOP_DOMAIN") // like example.com.

func TestShow(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	rs, err := p.GetRecords(context.Background(), TOP_DOMAIN)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}

func TestCreate(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	ip, _ := netip.ParseAddr("1.2.3.4")
	rs, err := p.AppendRecords(context.Background(), TOP_DOMAIN, []libdns.Record{
		libdns.Address{
			Name: "jedytest",
			IP:   ip,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}

func TestSet(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	ip, _ := netip.ParseAddr("192.168.1.1")
	rs, err := p.SetRecords(context.Background(), TOP_DOMAIN, []libdns.Record{
		libdns.Address{
			Name: "jedytest",
			IP:   ip,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}

func TestDelete(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	rs, err := p.DeleteRecords(context.Background(), TOP_DOMAIN, []libdns.Record{
		libdns.Address{
			Name: "jedytest",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}

func TestTXT(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	rs, err := p.AppendRecords(context.Background(), TOP_DOMAIN, []libdns.Record{
		libdns.TXT{
			Name: "jedytest",
			Text: "aaa_bbb_ccc",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}

func TestSubdomain(t *testing.T) {
	const Sub = "jedy.test."
	top, sub := toTopDomain(Sub + TOP_DOMAIN)
	if top != TOP_DOMAIN || sub != Sub[:len(Sub)-1] {
		t.Fatalf("unexpected result: [%s], [%s]", top, sub)
	}
	top, sub = toTopDomain(TOP_DOMAIN)
	if top != TOP_DOMAIN || sub != "" {
		t.Fatalf("unexpected result:[%s], [%s]", top, sub)
	}
}

func TestChangeSub(t *testing.T) {
	p := Provider{
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
	ip, _ := netip.ParseAddr("1.2.3.4")
	rs, err := p.AppendRecords(context.Background(), "jedydemo."+TOP_DOMAIN, []libdns.Record{
		libdns.Address{
			Name: "jedytest",
			IP:   ip,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}

	rs, err = p.DeleteRecords(context.Background(), "jedydemo."+TOP_DOMAIN, []libdns.Record{
		libdns.Address{
			Name: "jedytest",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rs {
		r := r.RR()
		t.Logf("%s[%s]: %s", r.Name, r.Type, r.Data)
	}
}
