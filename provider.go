package libhw

import (
	"context"
	"strings"

	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces
type Provider struct {
	Region          string `json:"region,omitempty"`
	AccessKeyId     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	client          *dns.DnsClient
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	zone, _ = toTopDomain(zone)
	zoneID, err := p.getZoneID(zone)
	if err != nil {
		return nil, err
	}

	records, err := p.getRecords(zoneID)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.SetRecords(ctx, zone, records)
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	zone, sub := toTopDomain(zone)
	zoneID, err := p.getZoneID(zone)
	if err != nil {
		return nil, err
	}

	curRecords, err := p.getRecords(zoneID)
	if err != nil {
		return nil, err
	}
	var deletedRecords []libdns.Record

	for _, record := range records {
		if sub != "" {
			record.Name = record.Name + "." + sub
		}
		found, r := findRecord(curRecords, record)
		if !found || (record.Value != "" && r.Value != record.Value) {
			continue
		}
		err := p.deleteRecord(zoneID, r.ID)
		if err != nil {
			return nil, err
		}
		deletedRecords = append(deletedRecords, r)
	}
	return deletedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	zone, sub := toTopDomain(zone)
	zoneID, err := p.getZoneID(zone)
	if err != nil {
		return nil, err
	}

	var updatedRecords []libdns.Record

	curRecords, err := p.getRecords(zoneID)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if sub != "" {
			record.Name = record.Name + "." + sub
		}
		found, r := findRecord(curRecords, record)
		if found {
			updatedRecord, err := p.updateRecord(zone, zoneID, r.ID, record)
			if err != nil {
				return nil, err
			}
			updatedRecords = append(updatedRecords, updatedRecord)
			continue
		}

		newRecord, err := p.createRecord(zoneID, record, zone)
		if err != nil {
			return nil, err
		}
		updatedRecords = append(updatedRecords, newRecord)
	}

	return updatedRecords, nil
}

// findRecord 查询name和type相同的记录
func findRecord(records []libdns.Record, record libdns.Record) (bool, libdns.Record) {
	for _, r := range records {
		if r.Name == record.Name && r.Type == record.Type {
			return true, r
		}
	}
	return false, libdns.Record{}
}

func toTopDomain(zone string) (top string, sub string) {
	j := len(zone)
	for range 3 {
		if j = strings.LastIndexByte(zone[:j], '.'); j == -1 {
			return zone, ""
		}
	}
	return zone[j+1:], zone[:j]
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
