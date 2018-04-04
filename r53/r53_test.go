package r53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/jcmturner/ddns/awsclient"
	"github.com/stretchr/testify/assert"
)

func TestUpdateRecord(t *testing.T) {
	awsCl := new(awsclient.MockClient)
	r53Cl, _ := awsCl.R53()

	err := UpdateRecord(r53Cl, aws.String(awsclient.MockHostedZoneID), awsclient.MockeZoneName, "1.2.3.4", route53.RRTypeA)
	if err != nil {
		t.Fatalf("error updating record: %v", err)
	}
}

func TestZoneID(t *testing.T) {
	awsCl := new(awsclient.MockClient)
	r53Cl, _ := awsCl.R53()

	zid, err := ZoneID(r53Cl, awsclient.MockeZoneName)
	if err != nil {
		t.Fatalf("error getting zone id: %v", err)
	}
	assert.Equal(t, awsclient.MockHostedZoneID, aws.StringValue(zid), "zone ID not expected")
}
