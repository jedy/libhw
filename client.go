package libhw

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	dnsRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"

	"github.com/libdns/libdns"
)

type libRR = libdns.RR

type Record struct {
	libRR
	ID string
}

func (p *Provider) init() error {
	if p.client != nil {
		return nil
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(p.AccessKeyId).
		WithSk(p.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return err
	}

	var region *region.Region
	if p.Region != "" {
		region, err = dnsRegion.SafeValueOf(p.Region)
		if err != nil {
			return err
		}
	}

	builder := dns.DnsClientBuilder().WithCredential(auth)
	if region == nil {
		// https://developer.huaweicloud.com/endpoint?DNS
		builder.WithEndpoints([]string{"https://dns.myhuaweicloud.com"})
	} else {
		builder.WithRegion(region)
	}

	hcClient, err := builder.SafeBuild()
	if err != nil {
		return err
	}

	p.client = dns.NewDnsClient(hcClient)
	return nil
}

func parseRecordSet(set model.ListRecordSets) []libdns.Record {
	records := make([]libdns.Record, 0, len(*set.Records))

	var ttl time.Duration
	if set.Ttl != nil {
		ttl = time.Duration(*set.Ttl) * time.Second
	}
	name := libdns.RelativeName(getString(set.Name), getString(set.ZoneName))
	for _, value := range *set.Records {
		t := getString(set.Type)
		if t == "TXT" {
			value = unquote(value)
		}
		records = append(records, Record{
			libRR: libdns.RR{
				Name: name,
				Data: value,
				Type: t,
				TTL:  ttl,
			},
			ID: getString(set.Id),
		})
	}
	return records
}

func parseLibdnsRecord(zone string, r libdns.Record) model.ListRecordSets {
	var s model.ListRecordSets
	rr := r.RR()
	ttl := int32(rr.TTL / time.Second)
	if ttl > 0 {
		s.Ttl = &ttl
	}
	t := rr.Type
	s.Type = &t
	name := libdns.AbsoluteName(rr.Name, zone)
	s.Name = &name
	v := rr.Data
	if t == "TXT" {
		v = quote(v)
	}
	s.Records = &[]string{v}
	return s
}

func getString(n *string) string {
	if n == nil {
		return ""
	}
	return *n
}

func (p *Provider) getRecords(zoneID string) ([]libdns.Record, error) {
	request := &model.ListRecordSetsByZoneRequest{}
	request.ZoneId = zoneID

	var records []libdns.Record

	getRecordResult, err := p.client.ListRecordSetsByZone(request)
	if err != nil {
		return records, err
	}

	recordSets := *getRecordResult.Recordsets
	for _, s := range recordSets {
		records = append(records, parseRecordSet(s)...)
	}

	return records, nil
}

func (p *Provider) getZoneID(zoneName string) (string, error) {
	request := &model.ListPublicZonesRequest{}
	limitRequest := int32(1)
	request.Limit = &limitRequest
	request.Name = &zoneName

	getZoneResult, err := p.client.ListPublicZones(request)
	if err != nil {
		return "", err
	}

	var matchingZones []model.PublicZoneResp
	if l := len(*getZoneResult.Zones); l > 0 {
		zones := *getZoneResult.Zones
		for z := 0; z < l; z++ {
			if *zones[z].Name == zoneName {
				matchingZones = append(matchingZones, zones[z])
			}
		}
	}

	if len(matchingZones) >= 1 {
		return *matchingZones[0].Id, nil
	}

	return "", fmt.Errorf("HostedZoneNotFound: No zones found for the domain %s", zoneName)
}

func (p *Provider) createRecord(zoneID string, record libdns.Record, zone string) (libdns.Record, error) {
	recordSet := parseLibdnsRecord(zone, record)
	requestBody := model.CreateRecordSetRequestBody{
		Records: *recordSet.Records,
		Type:    *recordSet.Type,
		Name:    *recordSet.Name,
		Ttl:     recordSet.Ttl,
	}
	request := &model.CreateRecordSetRequest{
		ZoneId: zoneID,
		Body:   &requestBody,
	}
	_, err := p.client.CreateRecordSet(request)
	if err != nil {
		return record, err
	}

	return record, nil
}

func (p *Provider) updateRecord(zone, zoneID, recordSetID string, record libdns.Record) (libdns.Record, error) {
	recordSet := parseLibdnsRecord(zone, record)
	request := &model.UpdateRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordSetID,
		Body: &model.UpdateRecordSetReq{
			Records: recordSet.Records,
			Type:    *recordSet.Type,
			Name:    *recordSet.Name,
			Ttl:     recordSet.Ttl,
		},
	}
	_, err := p.client.UpdateRecordSet(request)
	if err != nil {
		return record, err
	}
	return record, nil
}

func (p *Provider) deleteRecord(zoneID string, recordSetID string) error {
	request := &model.DeleteRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: recordSetID,
	}
	_, err := p.client.DeleteRecordSet(request)
	return err
}
