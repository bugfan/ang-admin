package models

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"time"
)

// Certificate represents the certificate configuration model
type Certificate struct {
	Id           int64     `xorm:"pk autoincr" json:"id"`
	CertId       string    `xorm:"'cert_id' varchar(64) unique notnull" json:"cert_id"` // e.g. "id-1", "id-2"
	Type         string    `xorm:"'type' varchar(32) notnull" json:"type"`           // "STD" (标密), "GM" (国密), "SELF-STD" (自签标密)
	KeyContent   string    `xorm:"'key_content' text notnull" json:"key_content"`     // Private Key (PEM format)
	CertContent  string    `xorm:"'cert_content' text notnull" json:"cert_content"`    // Certificate (PEM format)
	Remark       string    `xorm:"'remark' varchar(255)" json:"remark"`
	SubjectCN    string    `xorm:"'subject_cn' varchar(255)" json:"subject_cn"`    // 解析出的Common Name
	SANs         string    `xorm:"'sans' text" json:"sans"`                  // 解析出的SAN列表，逗号分隔
	NotBefore    time.Time `xorm:"'not_before' datetime" json:"not_before"`        // 证书生效开始时间
	NotAfter     time.Time `xorm:"'not_after' datetime" json:"not_after"`         // 证书过期结束时间
	Issuer       string    `xorm:"'issuer' varchar(255)" json:"issuer"`        // 颁发者
	SerialNumber string    `xorm:"'serial_number' varchar(128)" json:"serial_number"` // 序列号
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

// ParseCertInfo helper to extract SANs, CN, NotBefore, NotAfter from PEM CertContent
func (c *Certificate) ParseCertInfo() {
	if c.CertContent == "" {
		return
	}
	rest := []byte(c.CertContent)
	var firstCert *x509.Certificate
	var sanList []string

	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				if firstCert == nil {
					firstCert = cert
				}
				if cert.Subject.CommonName != "" {
					sanList = append(sanList, cert.Subject.CommonName)
				}
				sanList = append(sanList, cert.DNSNames...)
				for _, ip := range cert.IPAddresses {
					sanList = append(sanList, ip.String())
				}
			}
		}
	}

	if firstCert != nil {
		c.SubjectCN = firstCert.Subject.CommonName
		c.NotBefore = firstCert.NotBefore
		c.NotAfter = firstCert.NotAfter
		c.Issuer = firstCert.Issuer.CommonName
		if c.Issuer == "" {
			c.Issuer = firstCert.Issuer.String()
		}
		c.SerialNumber = firstCert.SerialNumber.String()
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
