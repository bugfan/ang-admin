package models

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"time"
)

// Certificate represents the certificate configuration model
type Certificate struct {
	Id           int64     `xorm:"pk autoincr" json:"id"`
	CertId       string    `xorm:"'cert_id' varchar(64) unique notnull" json:"cert_id"` // e.g. "id-1", "id-2"
	Type         string    `xorm:"'type' varchar(32) notnull" json:"type"`           // "STD" (标密), "GM" (国密)
	KeyContent   string    `xorm:"'key_content' text notnull" json:"key_content"`     // Private Key (PEM format)
	CertContent  string    `xorm:"'cert_content' text notnull" json:"cert_content"`    // Certificate (PEM format)
	Remark       string    `xorm:"'remark' varchar(255)" json:"remark"`
	SubjectCN    string    `xorm:"'subject_cn' varchar(255)" json:"subject_cn"`    // 解析出的Common Name
	SANs         string    `xorm:"'sans' text" json:"sans"`                  // 解析出的SAN列表，逗号分隔
	NotBefore    time.Time `xorm:"'not_before' datetime" json:"not_before"`        // 证书生效开始时间
	NotAfter     time.Time `xorm:"'not_after' datetime" json:"not_after"`         // 证书过期结束时间
	Issuer       string    `xorm:"'issuer' varchar(255)" json:"issuer"`        // 颁发者
	SerialNumber string    `xorm:"'serial_number' varchar(128)" json:"serial_number"` // 序列号
	Source       string    `xorm:"'source' varchar(32) default 'MANUAL'" json:"source"` // 证书来源: "MANUAL", "ACME", "SELF_SIGNED"
	AcmeAccountId int64    `xorm:"'acme_account_id' int default 0" json:"acme_account_id"`
	AutoRenew     bool     `xorm:"'auto_renew' bool default 0" json:"auto_renew"`
	RenewDays     int      `xorm:"'renew_days' int default 30" json:"renew_days"`
	Domains       string   `xorm:"'domains' text" json:"domains"`
	CreatedAt    time.Time `xorm:"created" json:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" json:"updated_at"`
}

func (Certificate) TableName() string {
	return "certificate"
}

func (c *Certificate) BeforeInsert() {
	c.ParseCertInfo()
}

func (c *Certificate) BeforeUpdate() {
	c.ParseCertInfo()
}

type rawCert struct {
	TBSCertificate rawTBSCertificate
	SignatureAlgorithm struct {
		Algorithm  asn1.ObjectIdentifier
		Parameters asn1.RawValue `asn1:"optional"`
	}
	SignatureValue asn1.BitString
}

type rawTBSCertificate struct {
	Raw                asn1.RawContent
	Version            int `asn1:"optional,explicit,default:0,tag:0"`
	SerialNumber       asn1.RawValue
	SignatureAlgorithm struct {
		Algorithm  asn1.ObjectIdentifier
		Parameters asn1.RawValue `asn1:"optional"`
	}
	Issuer             asn1.RawValue
	Validity           struct {
		NotBefore time.Time
		NotAfter  time.Time
	}
	Subject            asn1.RawValue
	PublicKey          struct {
		Algorithm struct {
			Algorithm  asn1.ObjectIdentifier
			Parameters asn1.RawValue `asn1:"optional"`
		}
		PublicKey asn1.BitString
	}
	UniqueId           asn1.BitString `asn1:"optional,tag:1"`
	SubjectUniqueId    asn1.BitString `asn1:"optional,tag:2"`
	Extensions         []struct {
		Id       asn1.ObjectIdentifier
		Critical bool `asn1:"optional"`
		Value    []byte
	} `asn1:"optional,explicit,tag:3"`
}

func parseRDNString(rdnRaw []byte, targetOID asn1.ObjectIdentifier) string {
	rest := rdnRaw
	for len(rest) > 0 {
		var rawValue asn1.RawValue
		var err error
		var nextRest []byte
		nextRest, err = asn1.Unmarshal(rest, &rawValue)
		if err != nil {
			if len(rest) > 1 {
				rest = rest[1:]
				continue
			}
			break
		}
		rest = nextRest

		if rawValue.Tag == 16 && rawValue.IsCompound {
			var atv struct {
				Type  asn1.ObjectIdentifier
				Value asn1.RawValue
			}
			if _, err := asn1.Unmarshal(rawValue.FullBytes, &atv); err == nil {
				if atv.Type.Equal(targetOID) {
					return string(atv.Value.Bytes)
				}
			}
		}

		if rawValue.IsCompound && len(rawValue.Bytes) > 0 {
			if res := parseRDNString(rawValue.Bytes, targetOID); res != "" {
				return res
			}
		}
	}
	return ""
}

func parseSANExtension(extVal []byte) []string {
	var sans []string
	var rawSeq []asn1.RawValue
	if _, err := asn1.Unmarshal(extVal, &rawSeq); err == nil {
		for _, v := range rawSeq {
			if v.Tag == 2 { // dNSName [2] IA5String
				sans = append(sans, string(v.Bytes))
			} else if v.Tag == 7 { // iPAddress [7] OCTET STRING
				if len(v.Bytes) == 4 || len(v.Bytes) == 16 {
					sans = append(sans, net.IP(v.Bytes).String())
				}
			}
		}
	}
	return sans
}

func parseSingleCertBlock(der []byte) (firstCN string, sans []string, nb, na time.Time, issuer, sn string, ok bool) {
	if cert, err := x509.ParseCertificate(der); err == nil {
		firstCN = cert.Subject.CommonName
		if firstCN != "" {
			sans = append(sans, firstCN)
		}
		sans = append(sans, cert.DNSNames...)
		for _, ip := range cert.IPAddresses {
			sans = append(sans, ip.String())
		}
		nb = cert.NotBefore
		na = cert.NotAfter
		issuer = cert.Issuer.CommonName
		if issuer == "" {
			issuer = cert.Issuer.String()
		}
		sn = cert.SerialNumber.String()
		ok = true
		return
	}

	// Fallback using raw ASN.1 unmarshaling for non-standard / lenient certificates
	var rc rawCert
	if _, err := asn1.Unmarshal(der, &rc); err == nil {
		tbs := rc.TBSCertificate
		nb = tbs.Validity.NotBefore
		na = tbs.Validity.NotAfter
		oidCN := asn1.ObjectIdentifier{2, 5, 4, 3}
		firstCN = parseRDNString(tbs.Subject.FullBytes, oidCN)
		if firstCN != "" {
			sans = append(sans, firstCN)
		}
		issuer = parseRDNString(tbs.Issuer.FullBytes, oidCN)
		if issuer == "" {
			oidO := asn1.ObjectIdentifier{2, 5, 4, 10}
			issuer = parseRDNString(tbs.Issuer.FullBytes, oidO)
		}
		if len(tbs.SerialNumber.Bytes) > 0 {
			sn = new(big.Int).SetBytes(tbs.SerialNumber.Bytes).String()
		}
		for _, ext := range tbs.Extensions {
			if ext.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) {
				sans = append(sans, parseSANExtension(ext.Value)...)
			}
		}
		ok = true
		return
	}
	return
}

// ParseCertInfo helper to extract SANs, CN, NotBefore, NotAfter from PEM CertContent
func (c *Certificate) ParseCertInfo() {
	if c.CertContent == "" {
		return
	}
	content := c.CertContent
	if strings.Contains(content, "\\n") && !strings.Contains(content, "\n") {
		content = strings.ReplaceAll(content, "\\n", "\n")
	}
	rest := []byte(content)
	var firstCN, issuer, sn string
	var sanList []string
	var notBefore, notAfter time.Time
	foundFirst := false

	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cn, sans, nb, na, iss, snum, ok := parseSingleCertBlock(block.Bytes)
			if ok {
				if !foundFirst {
					firstCN = cn
					notBefore = nb
					notAfter = na
					issuer = iss
					sn = snum
					foundFirst = true
				}
				sanList = append(sanList, sans...)
			}
		}
	}

	if foundFirst {
		c.SubjectCN = firstCN
		c.NotBefore = notBefore
		c.NotAfter = notAfter
		c.Issuer = issuer
		c.SerialNumber = sn
	}

	// Remove duplicate SAN entries
	uniqueMap := make(map[string]bool)
	var cleanSANs []string
	for _, item := range sanList {
		item = strings.TrimSpace(item)
		if item != "" && !uniqueMap[item] {
			uniqueMap[item] = true
			cleanSANs = append(cleanSANs, item)
		}
	}
	c.SANs = strings.Join(cleanSANs, ", ")
}
