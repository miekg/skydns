// Copyright (c) 2013 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package server

import (
	"github.com/miekg/dns"
	"strconv"
	"strings"
	"testing"
)

// This test file tests DNS and DNSSEC compliance and is different from
// the other tests in server_test.go which mostly check functionality.
// I'm grateful to Peter van Dijk of PowerDNS fame for copying the testsuite
// of PowerDNS to SkyDNS. See https://github.com/PowerDNS/pdns .

// Each test has a name, a question and a reply. To keep the amount of data
// manageable we input the tests as a subset of the text presentation format.

var compTestCases = []compTestCase{
	{Name: "direct-dnskey",
		Question: "skydns.local. DNSKEY",
		Flags:    "RCODE:0,OPCODE:0,RD,QR,tc,AA,ad",
	},
}

func TestCompliance(t *testing.T) {
	s := newTestServerDNSSEC("", "", "")
	defer s.Stop()
	c := new(dns.Client)
	for _, tc := range compTestCases {
		resp, _, err := c.Exchange(tc.request(), "localhost:"+StrPort)
		if err != nil {
			t.Fatal(err)
		}
		flags, _ := toReply(resp)
		t.Log("name: " + tc.Name)
		t.Log("flags: " + flags)
	}
}

type compTestCase struct {
	Name string
	// domainname qtype, e.g. "example.com. DNSKEY"
	Question string
	// disable DNSSEC, usually false
	NoDNSSEC bool
	// Abbreviated header:
	// RCODE:X,OPCODE:Y,RD,QR,TC,AA,AD
	// Uppercase flag: ON, lowercase: off
	Flags string
	// section domainname qtype RDATA
	// RDATA is a subset of the normal rdata, see the compare functions.
	// e.g. "0 example.com. DNSKEY 256 3 8 ..."
	Reply []string
}

// request creates a dns message with the question section set to Question.
func (c *compTestCase) request() *dns.Msg {
	m := new(dns.Msg)
	parts := strings.Split(c.Question, " ")
	m.SetQuestion(parts[0], dns.StringToType[parts[1]])
	return m
}

// toReply converts a dns message to a Flags string and a Reply string slice.
func toReply(m *dns.Msg) (string, []string) {
	flag := "RCODE:" + strconv.Itoa(m.Rcode) + ",OPCODE:" + strconv.Itoa(m.Opcode)
	// Order: QR,AA,TC,RD,RA,AD
	flag += "," + toFlag("qr,", m.Response)
	flag += toFlag("aa,", m.Authoritative)
	flag += toFlag("tc,", m.Truncated)
	flag += toFlag("rd,", m.RecursionDesired)
	flag += toFlag("ra,", m.RecursionAvailable)
	flag += toFlag("ad", m.AuthenticatedData)
	return flag, nil
}

func toFlag(s string, b bool) string {
	if b {
		return strings.ToUpper(s)
	}
	return strings.ToLower(s)
}

func newTestServerDNSSEC(leader, secret, nameserver string) *Server {
	s := newTestServer(leader, secret, nameserver)
	key, _ := dns.NewRR("skydns.local. IN DNSKEY 256 3 5 AwEAAaXfO+DOBMJsQ5H4TfiabwSpqE4cGL0Qlvh5hrQumrjr9eNSdIOjIHJJKCe56qBU5mH+iBlXP29SVf6UiiMjIrAPDVhClLeWFe0PC+XlWseAyRgiLHdQ8r95+AfkhO5aZgnCwYf9FGGSaT0+CRYN+PyDbXBTLK5FN+j5b6bb7z+d")
	s.dnsKey = key.(*dns.DNSKEY)
	s.keyTag = s.dnsKey.KeyTag()
	s.privKey, _ = s.dnsKey.ReadPrivateKey(strings.NewReader(`
Private-key-format: v1.3
Algorithm: 5 (RSASHA1)
Modulus: pd874M4EwmxDkfhN+JpvBKmoThwYvRCW+HmGtC6auOv141J0g6MgckkoJ7nqoFTmYf6IGVc/b1JV/pSKIyMisA8NWEKUt5YV7Q8L5eVax4DJGCIsd1Dyv3n4B+SE7lpmCcLBh/0UYZJpPT4JFg34/INtcFMsrkU36PlvptvvP50=
PublicExponent: AQAB
PrivateExponent: C6e08GXphbPPx6j36ZkIZf552gs1XcuVoB4B7hU8P/Qske2QTFOhCwbC8I+qwdtVWNtmuskbpvnVGw9a6X8lh7Z09RIgzO/pI1qau7kyZcuObDOjPw42exmjqISFPIlS1wKA8tw+yVzvZ19vwRk1q6Rne+C1romaUOTkpA6UXsE=
Prime1: 2mgJ0yr+9vz85abrWBWnB8Gfa1jOw/ccEg8ZToM9GLWI34Qoa0D8Dxm8VJjr1tixXY5zHoWEqRXciTtY3omQDQ==
Prime2: wmxLpp9rTzU4OREEVwF43b/TxSUBlUq6W83n2XP8YrCm1nS480w4HCUuXfON1ncGYHUuq+v4rF+6UVI3PZT50Q==
Exponent1: wkdTngUcIiau67YMmSFBoFOq9Lldy9HvpVzK/R0e5vDsnS8ZKTb4QJJ7BaG2ADpno7pISvkoJaRttaEWD3a8rQ==
Exponent2: YrC8OglEXIGkV3tm2494vf9ozPL6+cBkFsPPg9dXbvVCyyuW0pGHDeplvfUqs4nZp87z8PsoUL+LAUqdldnwcQ==
Coefficient: mMFr4+rDY5V24HZU3Oa5NEb55iQ56ZNa182GnNhWqX7UqWjcUUGjnkCy40BqeFAQ7lp52xKHvP5Zon56mwuQRw==
Created: 20140126132645
Publish: 20140126132645
Activate: 20140126132645`), "stdin")
	return s
}
